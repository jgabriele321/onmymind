package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db         *sql.DB
	lastPulled = make(map[int64]string) // chatID → last pulled text
	lpMutex    = &sync.RWMutex{}        // Protects lastPulled map
)

func initDB() error {
	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Open SQLite database
	dbPath := filepath.Join("data", "mind.db")
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
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

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

	// Load .env file
	if err := loadEnv(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}
	log.Printf("Using token: %s...[last 5 chars hidden]", token[:len(token)-5])

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

			if update.Message.IsCommand() {
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
					} else {
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
							msg.Text = "No items stored."
						} else {
							msg.Text = "📝 Stored items:\n" + strings.Join(items, "\n")
						}
					}
				case "deleted":
					// List deleted items
					rows, err := db.Query("SELECT text FROM deleted ORDER BY deleted_at DESC")
					if err != nil {
						log.Printf("Error listing deleted items: %v", err)
						msg.Text = "Failed to list deleted items."
					} else {
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
							msg.Text = "No deleted items."
						} else {
							msg.Text = "🗑️ Deleted items:\n" + strings.Join(items, "\n")
						}
					}
				case "help":
					msg.Text = `I understand these commands:
/add <text> - Store new text
/pull - Get a random item
/delete - Delete the last pulled item
/list - Show all stored items
/deleted - Show deleted items`
				default:
					msg.Text = "I don't know that command. Try /help"
				}

				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending message: %v", err)
					time.Sleep(time.Second * 1)
					continue
				}
			}
		}

		log.Printf("Update channel closed, reconnecting in 3 seconds...")
		time.Sleep(time.Second * 3)
	}
}
