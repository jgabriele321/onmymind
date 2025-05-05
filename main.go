package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	ldMutex = &sync.RWMutex{} // Protects lastDeleted map
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
				case "export":
					// Prepare response message
					msg.Text = "📦 Preparing your data export..."
					sentMsg, _ := bot.Send(msg)

					// Fetch current items
					rows, err := db.Query("SELECT text, created_at FROM items ORDER BY created_at")
					if err != nil {
						msg.Text = "❌ Failed to fetch items"
						break
					}
					defer rows.Close()

					var backup struct {
						Items []struct {
							Text      string    `json:"text"`
							CreatedAt time.Time `json:"created_at"`
						} `json:"items"`
						DeletedItems []struct {
							Text      string    `json:"text"`
							DeletedAt time.Time `json:"deleted_at"`
						} `json:"deleted_items"`
						ExportedAt time.Time `json:"exported_at"`
					}

					backup.ExportedAt = time.Now()

					// Read current items
					for rows.Next() {
						var item struct {
							Text      string    `json:"text"`
							CreatedAt time.Time `json:"created_at"`
						}
						if err := rows.Scan(&item.Text, &item.CreatedAt); err != nil {
							continue
						}
						backup.Items = append(backup.Items, item)
					}

					// Fetch deleted items
					rows, err = db.Query("SELECT text, deleted_at FROM deleted ORDER BY deleted_at")
					if err != nil {
						msg.Text = "❌ Failed to fetch deleted items"
						break
					}
					defer rows.Close()

					// Read deleted items
					for rows.Next() {
						var item struct {
							Text      string    `json:"text"`
							DeletedAt time.Time `json:"deleted_at"`
						}
						if err := rows.Scan(&item.Text, &item.DeletedAt); err != nil {
							continue
						}
						backup.DeletedItems = append(backup.DeletedItems, item)
					}

					// Convert to pretty JSON
					jsonData, err := json.MarshalIndent(backup, "", "    ")
					if err != nil {
						msg.Text = "❌ Failed to create backup"
						break
					}

					// Create file with timestamp
					filename := fmt.Sprintf("mindbot-backup-%s.json",
						time.Now().Format("2006-01-02-150405"))

					// Send as document
					fileBytes := tgbot.FileBytes{
						Name:  filename,
						Bytes: jsonData,
					}

					doc := tgbot.NewDocument(update.Message.Chat.ID, fileBytes)
					doc.Caption = fmt.Sprintf("📦 Your MindBot Backup\n• %d items\n• %d deleted items",
						len(backup.Items), len(backup.DeletedItems))

					if _, err := bot.Send(doc); err != nil {
						msg.Text = "❌ Failed to send backup file"
						break
					}

					// Delete the "preparing" message
					deleteMsg := tgbot.NewDeleteMessage(update.Message.Chat.ID, sentMsg.MessageID)
					bot.Send(deleteMsg)
					continue

				case "undo":
					// Check if there's something to undo
					ldMutex.RLock()
					lastDel, exists := lastDeleted[update.Message.Chat.ID]
					ldMutex.RUnlock()

					if !exists {
						msg.Text = "❌ Nothing to undo"
						break
					}

					// Only allow undo within 1 hour of deletion
					if time.Since(lastDel.DeletedAt) > time.Hour {
						msg.Text = "❌ Can't undo deletions older than 1 hour"
						break
					}

					// Start transaction
					tx, err := db.Begin()
					if err != nil {
						log.Printf("Error starting transaction: %v", err)
						msg.Text = "❌ Failed to undo deletion"
						break
					}

					// Move item back to items table
					if _, err := tx.Exec("INSERT INTO items (text) VALUES (?)", lastDel.Text); err != nil {
						tx.Rollback()
						log.Printf("Error restoring item: %v", err)
						msg.Text = "❌ Failed to restore item"
						break
					}

					// Remove from deleted table
					if _, err := tx.Exec("DELETE FROM deleted WHERE text = ?", lastDel.Text); err != nil {
						tx.Rollback()
						log.Printf("Error removing from deleted: %v", err)
						msg.Text = "❌ Failed to update deleted items"
						break
					}

					if err := tx.Commit(); err != nil {
						log.Printf("Error committing transaction: %v", err)
						msg.Text = "❌ Failed to complete undo"
						break
					}

					// Clear the last deleted item
					ldMutex.Lock()
					delete(lastDeleted, update.Message.Chat.ID)
					ldMutex.Unlock()

					msg.Text = fmt.Sprintf("✅ Restored: %s", lastDel.Text)

				case "help":
					msg.Text = `I understand these commands:
/add <text> - Store new text
/pull - Get a random item
/delete - Delete the last pulled item
/list - Show all stored items
/deleted - Show deleted items
/export - Download a backup of all your data
/undo - Restore the last deleted item (within 1 hour)
/help - Show this help message`
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
