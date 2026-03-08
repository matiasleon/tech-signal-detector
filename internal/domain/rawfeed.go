package domain

import "time"

type RawFeed struct {
	ID         string
	SourceID   string
	ExternalID string
	Title      string
	URL        string
	Score      float64
	PublishedAt time.Time
	FetchedAt  time.Time
}
