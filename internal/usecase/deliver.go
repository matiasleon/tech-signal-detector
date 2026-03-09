package usecase

import (
	"context"
	"log"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// Notifier sends a notification for a given feed item.
type Notifier interface {
	Send(ctx context.Context, title, url, sourceName string, publishedAt time.Time) error
}

// DeliverUseCase orchestrates delivering unsent signals via a Notifier.
type DeliverUseCase struct {
	signals  domain.SignalRepository
	rawFeeds domain.RawFeedRepository
	sources  domain.SourceRepository
	notifier Notifier
}

// NewDeliverUseCase creates a new DeliverUseCase.
func NewDeliverUseCase(
	signals domain.SignalRepository,
	rawFeeds domain.RawFeedRepository,
	sources domain.SourceRepository,
	notifier Notifier,
) *DeliverUseCase {
	return &DeliverUseCase{
		signals:  signals,
		rawFeeds: rawFeeds,
		sources:  sources,
		notifier: notifier,
	}
}

// Execute delivers each signal by fetching its associated RawFeed, notifying, and marking it as sent.
// It continues on individual failures so that one error does not block remaining signals.
// It returns the number of signals successfully sent.
func (uc *DeliverUseCase) Execute(ctx context.Context, signals []domain.Signal) (int, error) {
	log.Printf("[deliver] sending %d signals", len(signals))

	sourceCache := make(map[string]string) // sourceID → source.Name

	sent := 0
	for _, signal := range signals {
		feed, err := uc.rawFeeds.GetByID(ctx, signal.RawFeedID)
		if err != nil {
			log.Printf("[deliver] ERROR get raw feed for signal %s: %v", signal.ID, err)
			continue
		}

		sourceName, ok := sourceCache[feed.SourceID]
		if !ok {
			if src, err := uc.sources.GetByID(ctx, feed.SourceID); err == nil {
				sourceName = src.Name
			} else {
				sourceName = feed.SourceID
			}
			sourceCache[feed.SourceID] = sourceName
		}

		log.Printf("[deliver] sending: %s", feed.Title)
		if err := uc.notifier.Send(ctx, feed.Title, feed.URL, sourceName, signal.PublishedAt); err != nil {
			log.Printf("[deliver] ERROR send signal %s: %v", signal.ID, err)
			continue
		}

		if err := uc.signals.MarkAsSent(ctx, signal.ID); err != nil {
			log.Printf("[deliver] ERROR mark signal %s as sent: %v", signal.ID, err)
			continue
		}

		sent++
	}

	return sent, nil
}
