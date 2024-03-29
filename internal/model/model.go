package model

import "time"

type Item struct {
	Title      string
	Category   []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string
}

type Source struct {
	ID        int64
	Name      string
	FeedURL   string
	CreatedAt time.Time
	Priority  int
}

type Article struct {
	ID          int64
	SourceID    int64
	Title       string
	Link        string
	Summary     string
	CreatedAt   time.Time
	PostedAt    time.Time
	PublishedAt time.Time
}
