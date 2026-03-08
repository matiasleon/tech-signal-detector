package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

const defaultHNURL = "https://hn.algolia.com/api/v1/search?tags=story&hitsPerPage=30&numericFilters=points>0"

type hnResponse struct {
	Hits []hnHit `json:"hits"`
}

type hnHit struct {
	ObjectID  string `json:"objectID"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Points    int    `json:"points"`
	CreatedAt string `json:"created_at"`
}

// HackerNews is a Collector implementation that fetches stories from HackerNews via the Algolia API.
type HackerNews struct{}

// NewHackerNews returns a new HackerNews collector.
func NewHackerNews() *HackerNews {
	return &HackerNews{}
}

// Collect fetches HackerNews stories and maps them to domain.RawFeed values.
func (h *HackerNews) Collect(ctx context.Context, source domain.Source) ([]domain.RawFeed, error) {
	url := source.URL
	if url == "" {
		url = defaultHNURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("hackernews: create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hackernews: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hackernews: unexpected status %d", resp.StatusCode)
	}

	var result hnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("hackernews: decode response: %w", err)
	}

	feeds := make([]domain.RawFeed, 0, len(result.Hits))
	for _, hit := range result.Hits {
		publishedAt, err := time.Parse(time.RFC3339, hit.CreatedAt)
		if err != nil {
			publishedAt = time.Time{}
		}

		itemURL := hit.URL
		if itemURL == "" {
			itemURL = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", hit.ObjectID)
		}

		feeds = append(feeds, domain.RawFeed{
			SourceID:    source.ID,
			ExternalID:  hit.ObjectID,
			Title:       hit.Title,
			URL:         itemURL,
			Score:       float64(hit.Points),
			PublishedAt: publishedAt,
			FetchedAt:   time.Now(),
		})
	}

	return feeds, nil
}
