package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// Collector fetches raw feeds from a given source.
type Collector interface {
	Collect(ctx context.Context, source domain.Source) ([]domain.RawFeed, error)
}

// FetchUseCase orchestrates fetching raw feeds from all enabled sources.
type FetchUseCase struct {
	sources    domain.SourceRepository
	rawFeeds   domain.RawFeedRepository
	collectors map[domain.SourceType]Collector
}

// NewFetchUseCase creates a new FetchUseCase.
func NewFetchUseCase(
	sources domain.SourceRepository,
	rawFeeds domain.RawFeedRepository,
	collectors map[domain.SourceType]Collector,
) *FetchUseCase {
	return &FetchUseCase{
		sources:    sources,
		rawFeeds:   rawFeeds,
		collectors: collectors,
	}
}

// Execute fetches all enabled sources, deduplicates feeds, persists new ones, and returns them.
func (uc *FetchUseCase) Execute(ctx context.Context) ([]domain.RawFeed, error) {
	sources, err := uc.sources.GetEnabled(ctx)
	if err != nil {
		return nil, err
	}

	var results []domain.RawFeed

	for _, source := range sources {
		collector, ok := uc.collectors[source.Type]
		if !ok {
			continue
		}

		feeds, err := collector.Collect(ctx, source)
		if err != nil {
			return nil, err
		}

		for _, feed := range feeds {
			exists, err := uc.rawFeeds.ExistsByExternalID(ctx, feed.SourceID, feed.ExternalID)
			if err != nil {
				return nil, err
			}
			if exists {
				continue
			}

			feed.ID = uuid.NewString()
			feed.FetchedAt = time.Now()

			if err := uc.rawFeeds.Save(ctx, feed); err != nil {
				return nil, err
			}

			results = append(results, feed)
		}
	}

	return results, nil
}
