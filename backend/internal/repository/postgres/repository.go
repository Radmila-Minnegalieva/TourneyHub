// Package postgres реализует репозитории TourneyHub средствами PostgreSQL и pgx.
package postgres

import (
	"context"
	"errors"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX остаётся деталью реализации репозитория и не передаётся сервисам.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
type Repository struct {
	db   DBTX
	pool *pgxpool.Pool
}

var _ repository.Store = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository { return &Repository{db: pool, pool: pool} }
func (r *Repository) WithinTx(ctx context.Context, fn func(repository.Store) error) error {
	if r.pool == nil {
		return apperror.New(apperror.Unavailable, "database unavailable")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return dbError(err)
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	if err := fn(&Repository{db: tx}); err != nil {
		return dbError(err)
	}
	return dbError(tx.Commit(ctx))
}
func dbError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.Wrap(apperror.NotFound, "resource not found", repository.ErrNotFound)
	}
	var p *pgconn.PgError
	if errors.As(err, &p) {
		switch p.Code {
		case "23505":
			return apperror.Wrap(apperror.Conflict, "resource already exists or player is already registered", err)
		case "23503", "23514", "23502", "22P02":
			return apperror.Wrap(apperror.Validation, "invalid resource reference or value", err)
		case "40001", "40P01":
			return apperror.Wrap(apperror.Conflict, "concurrent modification; retry with current version", err)
		}
	}
	return err
}
func ignoreNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, repository.ErrNotFound)
}
