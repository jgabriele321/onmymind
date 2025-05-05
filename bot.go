package main
import (
	"fmt"
	"log"
	"os"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Simple version to test that the bot works
	bot, err := tgbot.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil && update.Message.IsCommand() {
			msg := tgbot.NewMessage(update.Message.Chat.ID, "")
			
			switch update.Message.Command() {
			case "add":
				text := update.Message.CommandArguments()
				if text == "" {
					msg.Text = "Usage: /add something"
				} else {
					msg.Text = fmt.Sprintf("Added: %s", text)
				}
			case "help":
				msg.Text = "I understand /add, /pull, /delete, /list, and /deleted commands."
			default:
				msg.Text = "I don't know that command"
			}
			
			bot.Send(msg)
		}
	}
}
