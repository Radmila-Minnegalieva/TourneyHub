package migration

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func Open(ctx context.Context, dsn string, files fs.FS) (*goose.Provider, *sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open migration database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("ping migration database: %w", err)
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, files, goose.WithSessionLocker(locker))
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("create migration provider: %w", err)
	}
	return provider, db, nil
}
