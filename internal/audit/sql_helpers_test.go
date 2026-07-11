package audit

import (
	"database/sql"
)

// sqlOpen is a tiny test helper so handler_test.go can open the same
// per-test db file that setupHandlerDB created. Kept separate so the
// handler tests don't have to import database/sql directly.
func sqlOpen(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", path)
}
