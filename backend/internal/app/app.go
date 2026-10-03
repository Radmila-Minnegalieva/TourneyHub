package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	handlers "github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/api/v1"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/container"
)

func Run(ctx context.Context, cfg config.Config) error {
	deps, err := container.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer deps.Close()
	server, err := handlers.NewServer(deps.Database, cfg.HTTP.CORSOrigins...)
	if err != nil {
		return err
	}
	server.Server.ReadHeaderTimeout = 5 * time.Second
	server.Server.ReadTimeout = 15 * time.Second
	server.Server.IdleTimeout = 60 * time.Second
	// WriteTimeout не задан: SSE-соединения остаются открытыми длительное время.
	errCh := make(chan error, 1)
	go func() { errCh <- server.Start(cfg.HTTP.Address) }()
	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
