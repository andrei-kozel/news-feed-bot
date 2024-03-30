package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ViewCmdStart() func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I'm a news feed bot. You can use /help to see all available commands.")
		_, err := bot.Send(msg)
		return err
	}
}
