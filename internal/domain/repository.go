package domain

import "context"

type SourceRepository interface {
	Save(ctx context.Context, source Source) error
	GetEnabled(ctx context.Context) ([]Source, error)
	GetByID(ctx context.Context, id string) (*Source, error)
}

type RawFeedRepository interface {
	Save(ctx context.Context, feed RawFeed) error
	ExistsByExternalID(ctx context.Context, sourceID, externalID string) (bool, error)
	GetByID(ctx context.Context, id string) (*RawFeed, error)
}

type SignalRepository interface {
	Save(ctx context.Context, signal Signal) error
	GetUnsent(ctx context.Context) ([]Signal, error)
	MarkAsSent(ctx context.Context, id string) error
}
