package sqlite

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

// DB wraps a *sql.DB for SQLite access.
type DB struct {
	db *sql.DB
}

// Open opens (or creates) the SQLite file at path and runs migrations.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	d := &DB{db: sqlDB}
	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return d, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

// migrate creates all tables if they do not already exist.
func (d *DB) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS sources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    score_threshold REAL NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS raw_feeds (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    external_id TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    score REAL NOT NULL DEFAULT 0,
    published_at DATETIME NOT NULL,
    fetched_at DATETIME NOT NULL,
    UNIQUE(source_id, external_id)
);

CREATE TABLE IF NOT EXISTS signals (
    id TEXT PRIMARY KEY,
    raw_feed_id TEXT NOT NULL,
    relevance_score REAL NOT NULL DEFAULT 0,
    sent_at DATETIME,
    created_at DATETIME NOT NULL
);
`
	if _, err := d.db.Exec(schema); err != nil {
		return err
	}

	// Add published_at column to signals if it does not already exist.
	// SQLite does not support IF NOT EXISTS for ALTER TABLE, so we ignore
	// "duplicate column" errors.
	_, alterErr := d.db.Exec(`ALTER TABLE signals ADD COLUMN published_at DATETIME NOT NULL DEFAULT ''`)
	if alterErr != nil && !isDuplicateColumn(alterErr) {
		return alterErr
	}

	return nil
}

// isDuplicateColumn reports whether err indicates a duplicate column in SQLite.
func isDuplicateColumn(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "duplicate column")
}

// contains is a simple substring check to avoid importing strings in the migration helper.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexStr(s, sub) >= 0)
}

func indexStr(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// queryRow is a thin wrapper around QueryRowContext.
func (d *DB) queryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// exec is a thin wrapper around ExecContext.
func (d *DB) exec(ctx context.Context, query string, args ...any) error {
	_, err := d.db.ExecContext(ctx, query, args...)
	return err
}
