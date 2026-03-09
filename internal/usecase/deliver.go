package usecase

import (
	"context"
	"log"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// Notifier sends a notification for a given feed item.
type Notifier interface {
	Send(ctx context.Context, title, url string, publishedAt time.Time) error
}

// DeliverUseCase orchestrates delivering unsent signals via a Notifier.
type DeliverUseCase struct {
	signals   domain.SignalRepository
	rawFeeds  domain.RawFeedRepository
	notifier  Notifier
}

// NewDeliverUseCase creates a new DeliverUseCase.
func NewDeliverUseCase(
	signals domain.SignalRepository,
	rawFeeds domain.RawFeedRepository,
	notifier Notifier,
) *DeliverUseCase {
	return &DeliverUseCase{
		signals:  signals,
		rawFeeds: rawFeeds,
		notifier: notifier,
	}
}

// Execute delivers each signal by fetching its associated RawFeed, notifying, and marking it as sent.
// It continues on individual failures so that one error does not block remaining signals.
// It returns the number of signals successfully sent.
func (uc *DeliverUseCase) Execute(ctx context.Context, signals []domain.Signal) (int, error) {
	log.Printf("[deliver] sending %d signals", len(signals))

	sent := 0
	for _, signal := range signals {
		feed, err := uc.rawFeeds.GetByID(ctx, signal.RawFeedID)
		if err != nil {
			log.Printf("[deliver] ERROR get raw feed for signal %s: %v", signal.ID, err)
			continue
		}

		log.Printf("[deliver] sending: %s", feed.Title)
		if err := uc.notifier.Send(ctx, feed.Title, feed.URL, signal.PublishedAt); err != nil {
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
