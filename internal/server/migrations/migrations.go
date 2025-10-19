// Package migrations embeds SQL migration files for use with database migration tools.
// The files are embedded using Go's embed.FS and can be used to run migrations at runtime
// without relying on external files.
package migrations

import (
	"embed"
)

//go:embed *.sql
var Migrations embed.FS
