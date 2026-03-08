package sqlite

import (
	"context"
	"time"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// SignalRepository implements domain.SignalRepository backed by SQLite.
type SignalRepository struct {
	db *DB
}

// NewSignalRepository returns a new SignalRepository.
func NewSignalRepository(db *DB) *SignalRepository {
	return &SignalRepository{db: db}
}

// Save inserts a new Signal row.
func (r *SignalRepository) Save(ctx context.Context, signal domain.Signal) error {
	const q = `
INSERT INTO signals (id, raw_feed_id, relevance_score, sent_at, created_at)
VALUES (?, ?, ?, ?, ?)`

	return r.db.exec(ctx, q,
		signal.ID,
		signal.RawFeedID,
		signal.RelevanceScore,
		signal.SentAt,
		signal.CreatedAt,
	)
}

// GetUnsent returns all signals where sent_at IS NULL.
func (r *SignalRepository) GetUnsent(ctx context.Context) ([]domain.Signal, error) {
	const q = `
SELECT id, raw_feed_id, relevance_score, sent_at, created_at
FROM signals
WHERE sent_at IS NULL`

	rows, err := r.db.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signals []domain.Signal
	for rows.Next() {
		var s domain.Signal
		var sentAt *time.Time
		if err := rows.Scan(&s.ID, &s.RawFeedID, &s.RelevanceScore, &sentAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.SentAt = sentAt
		signals = append(signals, s)
	}
	return signals, rows.Err()
}

// MarkAsSent sets sent_at to the current UTC time for the given signal id.
func (r *SignalRepository) MarkAsSent(ctx context.Context, id string) error {
	const q = `UPDATE signals SET sent_at = ? WHERE id = ?`
	return r.db.exec(ctx, q, time.Now().UTC(), id)
}
