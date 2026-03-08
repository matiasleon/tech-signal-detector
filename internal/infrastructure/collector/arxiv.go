package collector

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
	"github.com/mmcdole/gofeed"
)

const defaultArXivURL = "https://export.arxiv.org/rss/cs.AI+cs.LG"

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// ArXiv is a Collector implementation that fetches papers from an arXiv RSS/Atom feed.
type ArXiv struct {
	parser *gofeed.Parser
}

// NewArXiv returns a new ArXiv collector.
func NewArXiv() *ArXiv {
	return &ArXiv{
		parser: gofeed.NewParser(),
	}
}

// Collect fetches papers from the arXiv feed and maps them to domain.RawFeed values.
func (a *ArXiv) Collect(ctx context.Context, source domain.Source) ([]domain.RawFeed, error) {
	url := source.URL
	if url == "" {
		url = defaultArXivURL
	}

	feed, err := a.parser.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("arxiv: parse feed: %w", err)
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
			Title:       cleanHTML(item.Title),
			URL:         item.Link,
			Score:       0,
			PublishedAt: publishedAt,
			FetchedAt:   time.Now(),
		})
	}

	return feeds, nil
}

// cleanHTML strips HTML tags and unescapes HTML entities from a string.
func cleanHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}
