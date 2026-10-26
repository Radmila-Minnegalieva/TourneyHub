// Command bootstrap создаёт первого администратора через сервис приложения.
// Чтение и проверку ADMIN_* выполняет пакет config.
package main

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/container"
	"log/slog"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("bootstrap failed", "error", err)
		os.Exit(1)
	}
	slog.Info("administrator is ready")
}
func run() error {
	cfg, err := config.LoadBootstrap()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	deps, err := container.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer deps.Close()
	return deps.Service.BootstrapAdmin(ctx, cfg.Bootstrap.Email, cfg.Bootstrap.Nickname, cfg.Bootstrap.Password)
}
