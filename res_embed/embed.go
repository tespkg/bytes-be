package res_embed

import "embed"

//go:embed migration/pg/*.sql
var PgMigrationFiles embed.FS
