package sqlite

import (
	"context"
	"database/sql"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// RawFeedRepository implements domain.RawFeedRepository backed by SQLite.
type RawFeedRepository struct {
	db *DB
}

// NewRawFeedRepository returns a new RawFeedRepository.
func NewRawFeedRepository(db *DB) *RawFeedRepository {
	return &RawFeedRepository{db: db}
}

// Save inserts a RawFeed row, silently ignoring duplicates (same source_id + external_id).
func (r *RawFeedRepository) Save(ctx context.Context, feed domain.RawFeed) error {
	const q = `
INSERT OR IGNORE INTO raw_feeds
    (id, source_id, external_id, title, url, score, published_at, fetched_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	return r.db.exec(ctx, q,
		feed.ID,
		feed.SourceID,
		feed.ExternalID,
		feed.Title,
		feed.URL,
		feed.Score,
		feed.PublishedAt,
		feed.FetchedAt,
	)
}

// ExistsByExternalID reports whether a row with the given source_id and external_id exists.
func (r *RawFeedRepository) ExistsByExternalID(ctx context.Context, sourceID, externalID string) (bool, error) {
	const q = `SELECT COUNT(*) FROM raw_feeds WHERE source_id = ? AND external_id = ?`

	var count int
	err := r.db.queryRow(ctx, q, sourceID, externalID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetByID returns the RawFeed with the given id, or sql.ErrNoRows if not found.
func (r *RawFeedRepository) GetByID(ctx context.Context, id string) (*domain.RawFeed, error) {
	const q = `
SELECT id, source_id, external_id, title, url, score, published_at, fetched_at
FROM raw_feeds
WHERE id = ?`

	var f domain.RawFeed
	err := r.db.queryRow(ctx, q, id).Scan(
		&f.ID,
		&f.SourceID,
		&f.ExternalID,
		&f.Title,
		&f.URL,
		&f.Score,
		&f.PublishedAt,
		&f.FetchedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &f, nil
}
