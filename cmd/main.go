package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/andrei-kozel/news-feed-bot/internal/bot"
	"github.com/andrei-kozel/news-feed-bot/internal/botkit"
	"github.com/andrei-kozel/news-feed-bot/internal/config"
	"github.com/andrei-kozel/news-feed-bot/internal/fetcher"
	"github.com/andrei-kozel/news-feed-bot/internal/notifier"
	"github.com/andrei-kozel/news-feed-bot/internal/storage"
	"github.com/andrei-kozel/news-feed-bot/internal/summary"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
)

func main() {
	botAPI, err := tgbotapi.NewBotAPI(config.Get().TelegramBotToken)
	if err != nil {
		log.Printf("Failed to connect to Telegram API: %v", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	fmt.Printf("DatabaseDSN: %v\n", config.Get().DatabaseDSN)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer db.Close()

	var (
		articlesStorage = storage.NewArticleStorage(db)
		sourceStorage   = storage.NewSourceStorage(db)
		fetcher         = fetcher.New(
			articlesStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
		notifier = notifier.New(
			articlesStorage,
			summary.NewOpenAISummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
			botAPI,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)
	newsBot := botkit.New(botAPI)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func(ctx context.Context) {
		if err := fetcher.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] Fetcher failed: %v", err)
				return
			}
			log.Printf("[ERROR] Fetcher stopped: %v", err)
		}
	}(ctx)

	go func(ctx context.Context) {
		if err := notifier.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[ERROR] Notifier failed: %v", err)
				return
			}
			log.Printf("[ERROR] Notifier stopped: %v", err)
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("[ERROR] Bot failed: %v", err)
			return
		}

		log.Printf("[ERROR] Bot stopped: %v", err)
	}
}
