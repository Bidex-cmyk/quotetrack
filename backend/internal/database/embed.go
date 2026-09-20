package database

import "embed"

// Embed the migrations directory so the binary carries its own schema.
//
//go:embed migrations
var migrationsFS embed.FS
