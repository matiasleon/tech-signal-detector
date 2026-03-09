package usecase

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// RelevanceEvaluator determines whether a raw feed is relevant.
// It is used exclusively for arXiv feeds, where relevance cannot be
// determined by a numeric score alone.
type RelevanceEvaluator interface {
	Evaluate(ctx context.Context, feed domain.RawFeed) (bool, error)
}

// FilterUseCase evaluates raw feeds against per-source filtering rules,
// persists the ones that pass as domain.Signal records, and returns them.
type FilterUseCase struct {
	sources    domain.SourceRepository
	signals    domain.SignalRepository
	evaluator  RelevanceEvaluator
	maxSignals int
}

// NewFilterUseCase creates a new FilterUseCase.
func NewFilterUseCase(
	sources domain.SourceRepository,
	signals domain.SignalRepository,
	evaluator RelevanceEvaluator,
	maxSignals int,
) *FilterUseCase {
	return &FilterUseCase{
		sources:    sources,
		signals:    signals,
		evaluator:  evaluator,
		maxSignals: maxSignals,
	}
}

// Execute filters the provided feeds according to per-source rules,
// saves passing feeds as signals, and returns the created signals.
func (uc *FilterUseCase) Execute(ctx context.Context, feeds []domain.RawFeed) ([]domain.Signal, error) {
	// Cache sources looked up during this execution to avoid redundant DB calls.
	sourceCache := make(map[string]*domain.Source)

	var created []domain.Signal

	log.Printf("[filter] evaluating %d feeds (max signals: %d)", len(feeds), uc.maxSignals)

	for _, feed := range feeds {
		source, err := uc.resolveSource(ctx, feed.SourceID, sourceCache)
		if err != nil {
			return nil, fmt.Errorf("filter: resolve source for feed %s: %w", feed.ID, err)
		}

		passes, err := uc.passes(ctx, feed, source)
		if err != nil {
			return nil, fmt.Errorf("filter: evaluate feed %s: %w", feed.ID, err)
		}

		if !passes {
			log.Printf("[filter] SKIP  [%s] %s (score=%.0f threshold=%.0f)", source.Name, feed.Title, feed.Score, source.ScoreThreshold)
			continue
		}

		log.Printf("[filter] PASS  [%s] %s (score=%.0f)", source.Name, feed.Title, feed.Score)

		signal := domain.Signal{
			ID:             uuid.NewString(),
			RawFeedID:      feed.ID,
			RelevanceScore: feed.Score,
			CreatedAt:      time.Now(),
			PublishedAt:    feed.PublishedAt,
		}

		if err := uc.signals.Save(ctx, signal); err != nil {
			return nil, fmt.Errorf("filter: save signal for feed %s: %w", feed.ID, err)
		}

		created = append(created, signal)
	}

	// Sort all passing signals by PublishedAt descending and cap to maxSignals.
	sort.Slice(created, func(i, j int) bool {
		return created[i].PublishedAt.After(created[j].PublishedAt)
	})
	if uc.maxSignals > 0 && len(created) > uc.maxSignals {
		log.Printf("[filter] limiting %d signals to %d (sorted by published_at desc)", len(created), uc.maxSignals)
		created = created[:uc.maxSignals]
	}

	log.Printf("[filter] %d signals created", len(created))
	return created, nil
}

// resolveSource returns the Source for the given id, consulting the cache first.
func (uc *FilterUseCase) resolveSource(
	ctx context.Context,
	sourceID string,
	cache map[string]*domain.Source,
) (*domain.Source, error) {
	if s, ok := cache[sourceID]; ok {
		return s, nil
	}

	s, err := uc.sources.GetByID(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	cache[sourceID] = s
	return s, nil
}

// passes reports whether a feed should produce a signal, according to the
// filtering rule defined for its source type.
func (uc *FilterUseCase) passes(ctx context.Context, feed domain.RawFeed, source *domain.Source) (bool, error) {
	switch source.Type {
	case domain.SourceTypeArXiv:
		return uc.evaluator.Evaluate(ctx, feed)

	case domain.SourceTypeTechCrunch:
		// Editorial source — all items are considered relevant.
		return true, nil

	default:
		return feed.Score >= source.ScoreThreshold, nil
	}
}
