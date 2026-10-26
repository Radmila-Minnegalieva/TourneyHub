// Package service реализует сценарии TourneyHub независимо от HTTP и PostgreSQL.
package service

import (
	"context"
	"errors"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/authctx"
	paging "github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/pagination"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

type Mailer interface {
	Send(context.Context, model.MailMessage) error
}
type Service struct {
	repo   repository.Store
	cfg    config.Config
	mailer Mailer
}

func New(repo repository.Store, cfg config.Config, mailer Mailer) *Service {
	return &Service{repo: repo, cfg: cfg, mailer: mailer}
}
func (s *Service) withTx(ctx context.Context, fn func(repository.Store) error) error {
	if s.repo == nil {
		return unavailable("database unavailable")
	}
	return s.repo.WithinTx(ctx, fn)
}
func currentUser(ctx context.Context) uuid.UUID             { return authctx.CurrentUser(ctx) }
func hasRole(ctx context.Context, roles ...model.Role) bool { return authctx.HasRole(ctx, roles...) }
func requireRole(ctx context.Context, roles ...model.Role) error {
	if currentUser(ctx) == uuid.Nil {
		return unauthorized("authentication required")
	}
	if !hasRole(ctx, roles...) {
		return forbidden("insufficient role")
	}
	return nil
}
func invalid(message string) error      { return apperror.New(apperror.Validation, message) }
func conflict(message string) error     { return apperror.New(apperror.Conflict, message) }
func notFound(message string) error     { return apperror.New(apperror.NotFound, message) }
func forbidden(message string) error    { return apperror.New(apperror.Forbidden, message) }
func unauthorized(message string) error { return apperror.New(apperror.Unauthorized, message) }
func rateLimited(message string) error  { return apperror.New(apperror.RateLimited, message) }
func unavailable(message string) error  { return apperror.New(apperror.Unavailable, message) }

func ignoreNoRows(err error) bool              { return errors.Is(err, repository.ErrNotFound) }
func pagination(limit, offset *int) (int, int) { return paging.Values(limit, offset) }

func notify(ctx context.Context, q repository.Store, userID uuid.UUID, kind, message string, id uuid.UUID) error {
	if err := q.Notify(ctx, userID, kind, message, id); err != nil {
		return err
	}
	recipient, err := q.NotificationRecipient(ctx, userID)
	if err != nil {
		return err
	}
	if !recipient.Enabled {
		return nil
	}
	return q.Outbox(ctx, "email", id, model.MailMessage{To: recipient.Email, Subject: "TourneyHub: " + kind, Text: message})
}
