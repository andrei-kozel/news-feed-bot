package bot

import (
	"context"

	"github.com/andrei-kozel/news-feed-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func ViewCmdListAllSources(storage *storage.SourcePostgresStorage) func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Here is the list of all available sources:")

		sources, err := storage.Sources(ctx)
		if err != nil {
			msg.Text = "Failed to get sources"
		}

		for _, source := range sources {
			msg.Text += "\n" + source.Name
		}

		_, err = bot.Send(msg) // Fix: Replace := with =
		return err
	}
}
