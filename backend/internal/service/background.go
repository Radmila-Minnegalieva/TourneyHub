package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
)

// Background выполняет технические задачи с периодом из конфигурации приложения.
func (s *Service) Background(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.Worker.Interval)
	defer ticker.Stop()
	for {
		s.runBackground(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *Service) runBackground(ctx context.Context) {
	if err := s.AutoConfirm(ctx); err != nil && ctx.Err() == nil {
		slog.Error("confirm expired results", "error", err)
	}
	if err := s.DeliverMail(ctx); err != nil && ctx.Err() == nil {
		slog.Error("deliver outbox", "error", err)
	}
	if err := s.Cleanup(ctx); err != nil && ctx.Err() == nil {
		slog.Error("cleanup runtime records", "error", err)
	}
}
func (s *Service) DeliverMail(ctx context.Context) error {
	if s.mailer == nil {
		return nil
	}
	for i := 0; i < 50; i++ {
		done, err := s.deliverNextMail(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
	}
	return nil
}

// deliverNextMail удерживает запись outbox до фиксации результата доставки.
// Повтор SMTP после обрыва между DATA и commit допустим.
func (s *Service) deliverNextMail(ctx context.Context) (bool, error) {
	done := false
	err := s.withTx(ctx, func(q repository.Store) error {
		mail, err := q.PendingMail(ctx)
		if ignoreNoRows(err) {
			done = true
			return nil
		}
		if err != nil {
			return err
		}
		if err := s.mailer.Send(ctx, mail.Message); err != nil {
			slog.Warn("SMTP delivery deferred", "outbox_id", mail.ID)
			return q.RetryMail(ctx, mail.ID)
		}
		return q.CompleteMail(ctx, mail.ID)
	})
	return done, err
}
