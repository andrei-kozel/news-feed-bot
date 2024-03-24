package fetcher

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/andrei-kozel/news-feed-bot/internal/model"
	"github.com/andrei-kozel/news-feed-bot/internal/source"
)

type ArticleStorage interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceStorage interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles ArticleStorage
	sources  SourceStorage

	fetchInterval  time.Duration
	filterKeywords []string
}

func New(
	articleStorage ArticleStorage,
	sourceStorage SourceStorage,
	fetchInterval time.Duration,
	filterKeywords []string) *Fetcher {
	return &Fetcher{
		articles:       articleStorage,
		sources:        sourceStorage,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

func (f *Fetcher) Run(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		log.Printf("[ERROR] Failed to fetch articles: %v", err)
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				log.Printf("[ERROR] Failed to fetch articles: %v", err)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)

		rssSource := source.NewRSSSourceFromModel(src)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("[ERROR] Failed to fetch items from source %s: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("[ERROR] Failed to process items from source %s: %v", source.Name(), err)
				return
			}
		}(rssSource)
	}

	wg.Wait()
	return nil
}

func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		// TODO: Check if the item contains any of the filter keywords

		article := model.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			CreatedAt:   time.Now(),
			PostedAt:    item.Date.UTC(),
			PublishedAt: item.Date.UTC(),
		}

		if err := f.articles.Store(ctx, article); err != nil {
			return err
		}
	}

	return nil
}
