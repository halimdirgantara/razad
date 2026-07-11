package database

import (
	"database/sql"
	"path/filepath"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// openTestDB returns a fresh SQLite database in a temp directory. Foreign
// keys are enabled so the FK constraints embedded in the migrations are
// actually enforced during tests.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db") + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
	return db
}

func TestMigrateAppliesAllEmbeddedMigrations(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}

	applied, err := AppliedVersions(db)
	if err != nil {
		t.Fatalf("AppliedVersions: %v", err)
	}
	if len(applied) < 2 {
		t.Fatalf("expected at least 2 migrations applied, got %v", applied)
	}
	// Versions must be in ascending order.
	if !sort.SliceIsSorted(applied, func(i, j int) bool { return applied[i] < applied[j] }) {
		t.Fatalf("applied versions not sorted: %v", applied)
	}
	// Every embedded .sql file in the migrations package must have a recorded version.
	embedded := listEmbeddedMigrations(t)
	if len(embedded) != len(applied) {
		t.Fatalf("embedded=%d applied=%d (embedded=%v applied=%v)", len(embedded), len(applied), embedded, applied)
	}
	for i, v := range embedded {
		if applied[i] != v {
			t.Fatalf("applied[%d]=%q want %q", i, applied[i], v)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openTestDB(t)

	if err := Migrate(db); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	first, err := AppliedVersions(db)
	if err != nil {
		t.Fatalf("AppliedVersions after first run: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	second, err := AppliedVersions(db)
	if err != nil {
		t.Fatalf("AppliedVersions after second run: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("second run re-applied migrations: first=%v second=%v", first, second)
	}
	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("applied versions diverged at %d: %q vs %q", i, first[i], second[i])
		}
	}
}

func TestMigrateCreatesSchemaMigrationsTable(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	var name string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("schema_migrations not found: %v", err)
	}
	if name != "schema_migrations" {
		t.Fatalf("expected schema_migrations, got %q", name)
	}
}

func TestMigrateCreatesExpectedBaselineTables(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	want := []string{
		"users", "organizations", "organization_members", "sessions",
		"projects", "apps", "app_deployments", "app_env_vars",
		"audit_events", "database_instances",
	}
	for _, table := range want {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %s missing after migrate: %v", table, err)
		}
	}
}

func TestMigrateCreatesParityTablesFrom0002(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	want := []string{
		"servers", "node_agents", "node_heartbeats", "health_snapshots",
		"system_services", "log_sources", "provisioning_jobs",
		"domains", "domain_bindings", "ssl_certificates",
		"ai_policies", "ai_action_templates", "ai_actions",
		"database_credentials", "database_backups", "app_log_streams",
	}
	for _, table := range want {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("parity table %s missing after migrate: %v", table, err)
		}
	}
}

func TestMigrateKeepsCredentialColumnsForBackwardsCompatibility(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	// Transitional design: database_instances still carries username/password/
	// connection_string so the existing repository INSERTs/SELECTs continue to
	// work. A dedicated database_credentials table (added in 0002) is the
	// long-term home for these fields; a future migration will repoint
	// reads/writes and drop the deprecated columns. This test guards the
	// transitional state so a careless future migration cannot silently break
	// the existing code path.
	cols := mustTableColumns(t, db, "database_instances")
	required := []string{"username", "password", "connection_string"}
	for _, r := range required {
		found := false
		for _, c := range cols {
			if c == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("database_instances is missing transitional column %q (got %v)", r, cols)
		}
	}
}

func TestMigrateAddsCredentialsTable(t *testing.T) {
	db := openTestDB(t)
	if err := Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	cols := mustTableColumns(t, db, "database_credentials")
	required := []string{"database_instance_id", "username", "password", "connection_string"}
	for _, r := range required {
		found := false
		for _, c := range cols {
			if c == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("database_credentials missing column %q (got %v)", r, cols)
		}
	}
}

func TestSplitVersionValidatesNames(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantOK  bool
	}{
		{"plain", "0001_initial.sql", "0001", true},
		{"with_underscore_filename", "0042_add_domains.sql", "0042", true},
		{"no_underscore", "foo.sql", "", false},
		{"non_digit_prefix", "abcd_foo.sql", "", false},
		{"empty", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := splitVersion(tc.in)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("splitVersion(%q)=(%q,%v) want (%q,%v)", tc.in, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func mustTableColumns(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var def *string
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &def, &pk); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		out = append(out, name)
	}
	return out
}

// listEmbeddedMigrations walks the embedded migrations FS so a test can
// assert the runner applied everything that shipped in the binary.
func listEmbeddedMigrations(t *testing.T) []string {
	t.Helper()
	migs, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	out := make([]string, len(migs))
	for i, m := range migs {
		out[i] = m.Version
	}
	return out
}
