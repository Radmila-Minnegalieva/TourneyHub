package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/internal/migration"
	"github.com/Radmila-Minnegalieva/TourneyHub/migrations"
)

func run(ctx context.Context) error {
	command := "up"
	if len(os.Args) > 2 {
		return fmt.Errorf("usage: migrate [up|down|status]")
	}
	if len(os.Args) == 2 {
		command = os.Args[1]
	}
	if command != "up" && command != "down" && command != "status" {
		return fmt.Errorf("unknown migration command %q", command)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	provider, db, err := migration.Open(ctx, cfg.DatabaseURL, migrations.Files)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("close migration database", "error", err)
		}
	}()
	switch command {
	case "up":
		results, err := provider.Up(ctx)
		if err != nil {
			return err
		}
		for _, result := range results {
			slog.Info("migration applied", "migration", result.Source.Path)
		}
	case "down":
		_, err := provider.Down(ctx)
		return err
	case "status":
		results, err := provider.Status(ctx)
		if err != nil {
			return err
		}
		for _, result := range results {
			fmt.Printf("%03d %s %s\n", result.Source.Version, result.State, result.Source.Path)
		}
	}
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	if err := run(ctx); err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}
}
