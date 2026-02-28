package blog

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func openDB(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// WAL mode for concurrent reads
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	if err := createTables(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			post_slug  TEXT NOT NULL,
			author     TEXT NOT NULL,
			body       TEXT NOT NULL,
			visible    BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_comments_post_slug ON comments(post_slug);

		CREATE TABLE IF NOT EXISTS subscribers (
			id                INTEGER PRIMARY KEY AUTOINCREMENT,
			email             TEXT NOT NULL UNIQUE,
			verified          BOOLEAN NOT NULL DEFAULT 0,
			verify_token      TEXT NOT NULL,
			unsubscribe_token TEXT NOT NULL,
			created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS notified_posts (
			slug        TEXT PRIMARY KEY,
			notified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE VIRTUAL TABLE IF NOT EXISTS search_index USING fts5(
			slug, title, tags, body, content_type
		);
	`)
	if err != nil {
		return fmt.Errorf("creating tables: %w", err)
	}
	return nil
}
