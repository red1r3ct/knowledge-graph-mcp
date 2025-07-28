package migrations

import "embed"

//go:embed sqlite/*.sql
var MigrationsFS embed.FS
