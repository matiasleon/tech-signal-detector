package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
	"github.com/mmcdole/gofeed"
)

const defaultProductHuntURL = "https://www.producthunt.com/feed"

// ProductHunt is a Collector implementation that fetches products from the Product Hunt RSS feed.
type ProductHunt struct {
	parser *gofeed.Parser
}

// NewProductHunt returns a new ProductHunt collector.
func NewProductHunt() *ProductHunt {
	return &ProductHunt{
		parser: gofeed.NewParser(),
	}
}

// Collect fetches products from the Product Hunt RSS feed and maps them to domain.RawFeed values.
func (p *ProductHunt) Collect(ctx context.Context, source domain.Source) ([]domain.RawFeed, error) {
	url := source.URL
	if url == "" {
		url = defaultProductHuntURL
	}

	feed, err := p.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("producthunt: parse feed: %w", err)
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
