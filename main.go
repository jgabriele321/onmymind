package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jgabriele321/onmymind/reminder"
	timecalc "github.com/jgabriele321/onmymind/time"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db          *sql.DB
	lastPulled  = make(map[int64]string) // chatID → last pulled text
	lpMutex     = &sync.RWMutex{}        // Protects lastPulled map
	lastDeleted = make(map[int64]struct {
		Text      string
		DeletedAt time.Time
	})
	ldMutex         = &sync.RWMutex{} // Protects lastDeleted map
	timeCalculator  *timecalc.TimeCalculator
	reminderHandler *reminder.Handler
)

func startHealthCheck() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Printf("Starting health check server on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()
}

func initDB() error {
	// Get data directory from environment or use default
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		// Local development fallback
		dataDir = "data"
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Open SQLite database
	dbPath := filepath.Join(dataDir, "mind.db")
	log.Printf("Using database at: %s", dbPath)

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Create tables if they don't exist
	schema := `
	-- Existing tables
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS deleted (
		id INTEGER PRIMARY KEY,
		text TEXT NOT NULL,
		deleted_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- New reminder tables
	CREATE TABLE IF NOT EXISTS reminders (
		id TEXT PRIMARY KEY, -- UUID
		user_id TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		due_time DATETIME NOT NULL,
		recurrence_pattern TEXT,
		priority BOOLEAN DEFAULT 0,
		status TEXT CHECK(status IN ('pending', 'completed', 'cancelled')) DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reminder_logs (
		id TEXT PRIMARY KEY, -- UUID
		reminder_id TEXT NOT NULL,
		notification_type TEXT CHECK(notification_type IN ('telegram_message', 'telegram_call')) NOT NULL,
		status TEXT CHECK(status IN ('success', 'failed')) NOT NULL,
		error_message TEXT,
		attempted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (reminder_id) REFERENCES reminders(id) ON DELETE CASCADE
	);

	-- Indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_reminders_user_id ON reminders(user_id);
	CREATE INDEX IF NOT EXISTS idx_reminders_due_time ON reminders(due_time);
	CREATE INDEX IF NOT EXISTS idx_reminders_status ON reminders(status);
	CREATE INDEX IF NOT EXISTS idx_reminder_logs_reminder_id ON reminder_logs(reminder_id);
	CREATE INDEX IF NOT EXISTS idx_reminder_logs_attempted_at ON reminder_logs(attempted_at);`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %v", err)
	}

	return nil
}

func loadEnv(filename string) error {
	// Skip loading .env file in production (Render)
	if os.Getenv("RENDER") != "" {
		log.Printf("Running in Render, skipping .env file")
		return nil
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", absPath, err)
	}
	defer file.Close()

	log.Printf("Loading environment from: %s", absPath)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				os.Setenv(key, value)
				// Log key but mask the value for sensitive data
				if strings.Contains(strings.ToLower(key), "key") || strings.Contains(strings.ToLower(key), "token") {
					log.Printf("Set %s=***masked***", key)
				} else {
					log.Printf("Set %s=%s", key, value)
				}
			}
		}
	}
	return scanner.Err()
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Start health check server
	startHealthCheck()

	// Load .env file (only in development)
	if err := loadEnv(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get environment variables
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}
	log.Printf("Using token: %s...[last 5 chars hidden]", token[:len(token)-5])

	openRouterKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterKey == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is not set")
	}
	log.Printf("Using OpenRouter key: %s...[last 10 chars hidden]", openRouterKey[:len(openRouterKey)-10])

	// Initialize time calculator
	timeCalculator = timecalc.NewTimeCalculator(openRouterKey)

	// Initialize reminder system
	reminderStore := reminder.NewSQLiteStore(db)
	reminderService := reminder.NewService(reminderStore)
	location, err := time.LoadLocation("Local") // Use system timezone
	if err != nil {
		log.Printf("Warning: Failed to load local timezone: %v", err)
		location = time.UTC
	}
	reminderHandler = reminder.NewHandler(reminderService, location)

	// Create bot instance
	bot, err := tgbot.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Initialize and start the reminder scheduler
	scheduler := reminder.NewScheduler(reminderService, bot, location)
	scheduler.Start()
	defer scheduler.Stop()

	// Configure update parameters
	u := tgbot.NewUpdate(0)
	u.Timeout = 30 // Reduced timeout

	// Get updates with retry
	for {
		updates := bot.GetUpdatesChan(u)
		log.Printf("Started listening for updates...")

		for update := range updates {
			if update.Message == nil {
				continue
			}

			log.Printf("Received message: [%s] %s", update.Message.From.UserName, update.Message.Text)

			if !update.Message.IsCommand() {
				continue
			}

			msg := tgbot.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "remindme":
				response, err := reminderHandler.HandleRemindMe(update.Message)
				if err != nil {
					log.Printf("Error handling remindme: %v", err)
					msg.Text = "❌ An error occurred"
				} else {
					msg.Text = response
				}

			case "reminders":
				response, err := reminderHandler.HandleReminders(update.Message)
				if err != nil {
					log.Printf("Error handling reminders: %v", err)
					msg.Text = "❌ An error occurred"
				} else {
					msg.Text = response
				}

			case "delete":
				response, err := reminderHandler.HandleDelete(update.Message)
				if err != nil {
					log.Printf("Error handling delete: %v", err)
					msg.Text = "❌ An error occurred"
				} else {
					msg.Text = response
				}

			case "complete":
				response, err := reminderHandler.HandleComplete(update.Message)
				if err != nil {
					log.Printf("Error handling complete: %v", err)
					msg.Text = "❌ An error occurred"
				} else {
					msg.Text = response
				}

			case "time":
				query := update.Message.CommandArguments()
				if query == "" {
					msg.Text = "Usage: /time what's the time in New York?"
				} else {
					response, err := timeCalculator.ProcessQuery(query)
					if err != nil {
						log.Printf("Error processing time query: %v", err)
						msg.Text = fmt.Sprintf("Error: %v", err)
					} else {
						msg.Text = response
					}
				}

			case "help":
				msg.Text = `Available commands:
/remindme <time> to <message> [-call] - Set a reminder
/reminders [all|priority|regular] - List your reminders
/delete <reminder_id> - Delete a reminder
/complete <reminder_id> - Mark a reminder as completed
/time <query> - Calculate times and time zones

Examples:
• /remindme in 2 hours to check email
• /remindme tomorrow at 3pm to call mom -call
• /remindme every Sunday at 10am to water plants
• /time what's 2 hours before 3pm?
• /time convert 14:00 to EST`

			default:
				msg.Text = "I don't know that command. Try /help for available commands."
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}

		log.Printf("Updates channel closed, reconnecting...")
		time.Sleep(time.Second) // Wait before reconnecting
	}
}
