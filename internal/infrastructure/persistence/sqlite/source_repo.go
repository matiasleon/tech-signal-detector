package sqlite

import (
	"context"
	"database/sql"

	"github.com/matiasleonperalta/tech-signal-detectors/internal/domain"
)

// scannable abstracts *sql.Row and *sql.Rows so scan logic can be shared.
type scannable interface {
	Scan(dest ...any) error
}

// SourceRepository implements domain.SourceRepository backed by SQLite.
type SourceRepository struct {
	db *DB
}

// NewSourceRepository returns a new SourceRepository.
func NewSourceRepository(db *DB) *SourceRepository {
	return &SourceRepository{db: db}
}

// Save inserts or replaces a source.
func (r *SourceRepository) Save(ctx context.Context, s domain.Source) error {
	const q = `INSERT OR REPLACE INTO sources (id, name, type, url, enabled, score_threshold)
	            VALUES (?, ?, ?, ?, ?, ?)`
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	return r.db.exec(ctx, q, s.ID, s.Name, string(s.Type), s.URL, enabled, s.ScoreThreshold)
}

// GetEnabled returns all sources with enabled = 1.
func (r *SourceRepository) GetEnabled(ctx context.Context) ([]domain.Source, error) {
	const q = `SELECT id, name, type, url, enabled, score_threshold FROM sources WHERE enabled = 1`

	rows, err := r.db.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []domain.Source
	for rows.Next() {
		s, err := r.scan(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, *s)
	}
	return sources, rows.Err()
}

// GetByID returns the source with the given id, or sql.ErrNoRows if not found.
func (r *SourceRepository) GetByID(ctx context.Context, id string) (*domain.Source, error) {
	const q = `SELECT id, name, type, url, enabled, score_threshold FROM sources WHERE id = ?`
	return r.scan(r.db.queryRow(ctx, q, id))
}

// scan reads a Source from any scannable (row or rows).
func (r *SourceRepository) scan(row scannable) (*domain.Source, error) {
	var s domain.Source
	var enabled int
	err := row.Scan(&s.ID, &s.Name, (*string)(&s.Type), &s.URL, &enabled, &s.ScoreThreshold)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	s.Enabled = enabled != 0
	return &s, nil
}
