package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
	"github.com/mmcdole/gofeed"
)

const defaultTechCrunchURL = "https://techcrunch.com/feed/"

// TechCrunch is a Collector implementation that fetches articles from the TechCrunch RSS feed.
type TechCrunch struct {
	parser *gofeed.Parser
}

// NewTechCrunch returns a new TechCrunch collector.
func NewTechCrunch() *TechCrunch {
	return &TechCrunch{
		parser: gofeed.NewParser(),
	}
}

// Collect fetches articles from the TechCrunch RSS feed and maps them to domain.RawFeed values.
func (t *TechCrunch) Collect(ctx context.Context, source domain.Source) ([]domain.RawFeed, error) {
	url := source.URL
	if url == "" {
		url = defaultTechCrunchURL
	}

	feed, err := t.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("techcrunch: parse feed: %w", err)
	}

	feeds := make([]domain.RawFeed, 0, len(feed.Items))
	for _, item := range feed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}

		feeds = append(feeds, domain.RawFeed{
			SourceID:    source.ID,
			ExternalID:  item.GUID,
			Title:       item.Title,
			URL:         item.Link,
			Score:       0,
			PublishedAt: publishedAt,
			FetchedAt:   time.Now(),
		})
	}

	return feeds, nil
}
