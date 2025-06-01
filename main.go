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
	ldMutex        = &sync.RWMutex{} // Protects lastDeleted map
	timeCalculator *timecalc.TimeCalculator
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
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS deleted (
		id INTEGER PRIMARY KEY,
		text TEXT NOT NULL,
		deleted_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %v", err)
	}

	return nil
}

func loadEnv(filename string) error {
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

	// Load .env file
	if err := loadEnv(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

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

	// Simple version to test that the bot works
	bot, err := tgbot.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

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
			case "add":
				text := update.Message.CommandArguments()
				if text == "" {
					msg.Text = "Usage: /add something"
				} else {
					// Store in database
					if _, err := db.Exec("INSERT INTO items (text) VALUES (?)", text); err != nil {
						log.Printf("Error storing item: %v", err)
						msg.Text = "Failed to store item."
					} else {
						msg.Text = fmt.Sprintf("Added: %s ✅", text)
					}
				}
			case "pull":
				// Get random item from database
				var text string
				err := db.QueryRow("SELECT text FROM items ORDER BY RANDOM() LIMIT 1").Scan(&text)
				if err == sql.ErrNoRows {
					msg.Text = "No items available."
				} else if err != nil {
					log.Printf("Error pulling item: %v", err)
					msg.Text = "Failed to pull item."
				} else {
					lpMutex.Lock()
					lastPulled[update.Message.Chat.ID] = text
					lpMutex.Unlock()
					msg.Text = fmt.Sprintf("🎲 %s", text)
				}
			case "delete":
				// Delete last pulled item
				lpMutex.RLock()
				text, ok := lastPulled[update.Message.Chat.ID]
				lpMutex.RUnlock()

				if !ok {
					msg.Text = "Pull an item first using /pull"
				} else {
					tx, err := db.Begin()
					if err != nil {
						log.Printf("Error starting transaction: %v", err)
						msg.Text = "Failed to delete item."
						break
					}

					// Store the item for undo functionality
					ldMutex.Lock()
					lastDeleted[update.Message.Chat.ID] = struct {
						Text      string
						DeletedAt time.Time
					}{
						Text:      text,
						DeletedAt: time.Now(),
					}
					ldMutex.Unlock()

					// Move item to deleted table
					if _, err := tx.Exec("INSERT INTO deleted (text) VALUES (?)", text); err != nil {
						tx.Rollback()
						log.Printf("Error moving item to deleted: %v", err)
						msg.Text = "Failed to delete item."
						break
					}

					// Delete from items table
					if _, err := tx.Exec("DELETE FROM items WHERE text = ?", text); err != nil {
						tx.Rollback()
						log.Printf("Error deleting item: %v", err)
						msg.Text = "Failed to delete item."
						break
					}

					if err := tx.Commit(); err != nil {
						log.Printf("Error committing transaction: %v", err)
						msg.Text = "Failed to delete item."
						break
					}

					lpMutex.Lock()
					delete(lastPulled, update.Message.Chat.ID)
					lpMutex.Unlock()

					msg.Text = fmt.Sprintf("Deleted: %s 🗑️", text)
				}
			case "list":
				// List all items
				rows, err := db.Query("SELECT text FROM items ORDER BY created_at DESC")
				if err != nil {
					log.Printf("Error listing items: %v", err)
					msg.Text = "Failed to list items."
					break
				}
				defer rows.Close()

				var items []string
				for rows.Next() {
					var text string
					if err := rows.Scan(&text); err != nil {
						log.Printf("Error scanning row: %v", err)
						continue
					}
					items = append(items, "• "+text)
				}

				if len(items) == 0 {
					msg.Text = "No items available."
				} else {
					msg.Text = "Your items:\n" + strings.Join(items, "\n")
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
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}

		log.Printf("Updates channel closed, reconnecting...")
		time.Sleep(time.Second) // Wait before reconnecting
	}
}
