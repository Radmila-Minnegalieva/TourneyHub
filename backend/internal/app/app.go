package app

import (
	"context"
	"errors"
	"net/http"

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
	workerCtx, stopWorkers := context.WithCancel(ctx)
	defer stopWorkers()
	workersDone := make(chan struct{})
	go func() { defer close(workersDone); deps.Service.Background(workerCtx) }()
	defer func() { stopWorkers(); <-workersDone }()
	server, err := handlers.NewServer(deps.Database, deps.Service, cfg.HTTP)
	if err != nil {
		return err
	}
	server.Server.ReadHeaderTimeout = cfg.HTTP.ReadHeaderTimeout
	server.Server.ReadTimeout = cfg.HTTP.ReadTimeout
	server.Server.IdleTimeout = cfg.HTTP.IdleTimeout
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
