package store

import (
	"fmt"
	"sync"

	log "github.com/go-pkgz/lgr"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// DB implements Store interface using SQLite
type DB struct {
	db *sqlx.DB
	mu sync.RWMutex
}

// New creates a new DB store with the given database path
func New(dbPath string) (*DB, error) {
	db, err := connectSQLite(dbPath)
	if err != nil {
		return nil, err
	}

	store := &DB{db: db}

	if err := store.createSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	log.Printf("[DEBUG] initialized sqlite store at %s", dbPath)
	return store, nil
}

func connectSQLite(dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=1000",
		"PRAGMA foreign_keys=ON",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to set pragma %q: %w", pragma, err)
		}
	}

	db.SetMaxOpenConns(1)
	return db, nil
}

func (s *DB) createSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS services (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			description TEXT DEFAULT '',
			category TEXT DEFAULT 'Other',
			logo_url TEXT DEFAULT '',
			color TEXT DEFAULT '#0062ff',
			status TEXT NOT NULL DEFAULT 'active',
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			sort_order INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);
		CREATE INDEX IF NOT EXISTS idx_services_category ON services(category);
		CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
		CREATE INDEX IF NOT EXISTS idx_categories_sort ON categories(sort_order);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

// Close closes the database connection
func (s *DB) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
