package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Keep one shared connection for SQLite to avoid lock churn across pooled conns.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, fmt.Errorf("set journal_mode WAL: %w", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}
	if _, err := db.Exec(`PRAGMA synchronous=NORMAL;`); err != nil {
		return nil, fmt.Errorf("set synchronous: %w", err)
	}

	return db, nil
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Activity:      &ActivityRepo{db: db},
		Actions:       &ActionsRepo{db: db},
		Settings:      &SettingsRepo{db: db},
		Backfill:      &BackfillRepo{db: db},
		ReactionRoles: &ReactionRolesRepo{db: db},
		Warnings:      &WarningsRepo{db: db},
		Scheduled:     &ScheduledMessagesRepo{db: db},
	}
}

type Repositories struct {
	Activity      *ActivityRepo
	Actions       *ActionsRepo
	Settings      *SettingsRepo
	Backfill      *BackfillRepo
	ReactionRoles *ReactionRolesRepo
	Warnings      *WarningsRepo
	Scheduled     *ScheduledMessagesRepo
}
