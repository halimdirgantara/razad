// Package database provides database connection management and migration support.
// It supports both PostgreSQL and SQLite backends through a common interface.
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/razad/razad/internal/config"
)

// Open opens a database connection based on the provided config.
func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	driver := cfg.Driver
	if driver == "" {
		driver = "sqlite3"
	}

	db, err := sql.Open(driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("database: failed to open %s: %w", driver, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	if cfg.ConnMaxLifetimeSeconds > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return db, nil
}
