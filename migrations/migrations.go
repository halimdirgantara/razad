// Package migrations embeds the SQL migration files for Razad so they ship
// inside the daemon binary. Migration files live at the repository root under
// `migrations/` and are named NNNN_description.sql where NNNN is a
// zero-padded sequence number. The runner applies them in lexicographic order
// and records applied versions in the `schema_migrations` table.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
