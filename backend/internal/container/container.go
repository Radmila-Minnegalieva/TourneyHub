package container

import (
	"context"
	"fmt"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Database *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return &Container{Database: pool}, nil
}

func (c *Container) Close() { c.Database.Close() }
