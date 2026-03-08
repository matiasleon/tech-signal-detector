package usecase

import (
	"context"
	"fmt"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// Notifier sends a notification for a given feed item.
type Notifier interface {
	Send(ctx context.Context, title, url string) error
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
func (uc *DeliverUseCase) Execute(ctx context.Context, signals []domain.Signal) error {
	for _, signal := range signals {
		feed, err := uc.rawFeeds.GetByID(ctx, signal.RawFeedID)
		if err != nil {
			fmt.Printf("deliver: fetch raw feed for signal %s: %v\n", signal.ID, fmt.Errorf("get raw feed: %w", err))
			continue
		}

		if err := uc.notifier.Send(ctx, feed.Title, feed.URL); err != nil {
			fmt.Printf("deliver: notify signal %s: %v\n", signal.ID, fmt.Errorf("send notification: %w", err))
			continue
		}

		if err := uc.signals.MarkAsSent(ctx, signal.ID); err != nil {
			fmt.Printf("deliver: mark signal %s as sent: %v\n", signal.ID, fmt.Errorf("mark as sent: %w", err))
			continue
		}
	}

	return nil
}
