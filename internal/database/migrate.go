package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/razad/razad/migrations"
)

// migrationsTable tracks which migration versions have been applied.
const migrationsTable = "schema_migrations"

// migration is one parsed SQL file from the migrations/ directory.
type migration struct {
	Version string
	Name    string
	SQL     string
}

// Migrate applies every pending migration in lexicographic order.
//
// Applied versions are recorded in `schema_migrations` so subsequent calls are
// idempotent. The migration files themselves use CREATE TABLE IF NOT EXISTS,
// so partial failures mid-batch leave the database in a recoverable state —
// the runner can be re-invoked after fixing the cause.
func Migrate(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database: cannot migrate: db is nil")
	}

	// SQLite needs this per-connection. Postgres ignores it.
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("database: enable foreign keys: %w", err)
	}

	if err := ensureMigrationsTable(db); err != nil {
		return err
	}

	items, err := loadMigrations()
	if err != nil {
		return err
	}

	applied, err := loadApplied(db)
	if err != nil {
		return err
	}

	for _, m := range items {
		if _, ok := applied[m.Version]; ok {
			continue
		}
		if err := applyOne(db, m); err != nil {
			return fmt.Errorf("database: apply migration %s: %w", m.Version, err)
		}
	}
	return nil
}

// ensureMigrationsTable creates the tracking table if it does not exist.
func ensureMigrationsTable(db *sql.DB) error {
	stmt := `CREATE TABLE IF NOT EXISTS ` + migrationsTable + ` (
		version TEXT PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("database: create %s: %w", migrationsTable, err)
	}
	return nil
}

// loadMigrations reads all *.sql files from the embedded migrations FS,
// parses them into migration structs, and returns them sorted by version.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("database: read embedded migrations: %w", err)
	}

	var items []migration
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		version, ok := splitVersion(name)
		if !ok {
			return nil, fmt.Errorf("database: invalid migration filename %q (expected NNNN_*.sql)", name)
		}
		raw, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return nil, fmt.Errorf("database: read %s: %w", name, err)
		}
		items = append(items, migration{
			Version: version,
			Name:    name,
			SQL:     string(raw),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Version < items[j].Version
	})
	return items, nil
}

// splitVersion extracts the leading numeric portion of a migration filename.
func splitVersion(name string) (string, bool) {
	idx := strings.Index(name, "_")
	if idx <= 0 {
		return "", false
	}
	candidate := name[:idx]
	for _, r := range candidate {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return candidate, true
}

// loadApplied returns the set of migration versions already recorded.
func loadApplied(db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.Query("SELECT version FROM " + migrationsTable)
	if err != nil {
		return nil, fmt.Errorf("database: select applied: %w", err)
	}
	defer rows.Close()

	applied := map[string]struct{}{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("database: scan version: %w", err)
		}
		applied[v] = struct{}{}
	}
	return applied, rows.Err()
}

// applyOne runs a single migration's SQL and records its version on success.
func applyOne(db *sql.DB, m migration) error {
	if _, err := db.Exec(m.SQL); err != nil {
		return err
	}
	if _, err := db.Exec(
		"INSERT INTO "+migrationsTable+" (version) VALUES (?)",
		m.Version,
	); err != nil {
		return fmt.Errorf("record version: %w", err)
	}
	return nil
}

// AppliedVersions returns the list of migration versions currently recorded
// in the database. It is intended for diagnostics and tests.
func AppliedVersions(db *sql.DB) ([]string, error) {
	if err := ensureMigrationsTable(db); err != nil {
		return nil, err
	}
	rows, err := db.Query("SELECT version FROM " + migrationsTable + " ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("database: select applied: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("database: scan version: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
