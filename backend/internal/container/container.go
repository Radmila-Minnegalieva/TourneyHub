package container

import (
	"context"
	"fmt"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/smtpclient"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository/postgres"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Container struct {
	Database   *pgxpool.Pool
	Repository repository.Store
	Service    *service.Service
}

func New(ctx context.Context, cfg config.Config) (*Container, error) {
	pool, err := pgxpool.New(ctx, cfg.Database.URL())
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect database: %w", err)
	}
	store := postgres.New(pool)
	var mailer service.Mailer
	if cfg.SMTP.Address != "" {
		mailer = smtpclient.New(cfg.SMTP)
	}
	return &Container{Database: pool, Repository: store, Service: service.New(store, cfg, mailer)}, nil
}

func (c *Container) Close() { c.Database.Close() }
