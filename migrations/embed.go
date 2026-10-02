// Package migrations embeds the versioned goose SQL migrations in both binaries.
package migrations

import "embed"

// Files contains SQL only; goose ignores the Go source in this directory.
//
//go:embed *.sql
var Files embed.FS
