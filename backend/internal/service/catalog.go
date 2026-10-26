package service

import (
	"context"
	"strings"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) ListDisciplines(ctx context.Context, p model.ListDisciplinesParams) (model.DisciplinePage, error) {
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.DisciplinePage(ctx, limit, offset)
}

func validateDiscipline(d model.DisciplineRequest) error {
	if strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.ScoreUnit) == "" {
		return invalid("name and score_unit must not be blank")
	}
	if d.MinRosterSize > d.MaxRosterSize {
		return invalid("min_roster_size exceeds max_roster_size")
	}
	return nil
}

func (s *Service) CreateDiscipline(ctx context.Context, req model.DisciplineRequest) (model.Discipline, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.Discipline{}, err
	}
	if err := validateDiscipline(req); err != nil {
		return model.Discipline{}, err
	}
	var result model.Discipline
	err := s.withTx(ctx, func(tx repository.Store) error {
		var err error
		result, err = tx.CreateDisciplineRecord(ctx, req)
		if err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "discipline_created", "discipline", result.Id, nil, result)
	})
	return result, err
}

func (s *Service) UpdateDiscipline(ctx context.Context, id uuid.UUID, req model.UpdateDisciplineRequest) (model.Discipline, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.Discipline{}, err
	}
	var result model.Discipline
	err := s.withTx(ctx, func(tx repository.Store) error {
		old, err := tx.Discipline(ctx, id, true)
		if err != nil {
			return err
		}
		if old.Version != req.Version {
			return conflict("stale discipline version")
		}
		result = mergeDiscipline(old, req)
		if err := validateDiscipline(model.DisciplineRequest{Name: result.Name, MinRosterSize: result.MinRosterSize, MaxRosterSize: result.MaxRosterSize, AllowsDraw: result.AllowsDraw, ScoreUnit: result.ScoreUnit}); err != nil {
			return err
		}
		result, err = tx.UpdateDisciplineRecord(ctx, id, result)
		if err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "discipline_updated", "discipline", id, old, result)
	})
	return result, err
}

// mergeDiscipline сохраняет значения полей, пропущенных в PATCH.
func mergeDiscipline(old model.Discipline, p model.UpdateDisciplineRequest) model.Discipline {
	replacePresent(&old.Name, p.Name)
	replacePresent(&old.MinRosterSize, p.MinRosterSize)
	replacePresent(&old.MaxRosterSize, p.MaxRosterSize)
	replacePresent(&old.AllowsDraw, p.AllowsDraw)
	replacePresent(&old.ScoreUnit, p.ScoreUnit)
	replacePresent(&old.Active, p.Active)
	return old
}

func (s *Service) ListRatings(ctx context.Context, p model.ListRatingsParams) (model.RatingPage, error) {
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.RatingPage(ctx, p.DisciplineId, limit, offset)
}

func (s *Service) ListRatingHistory(ctx context.Context, id uuid.UUID, p model.ListRatingHistoryParams) (model.RatingChangePage, error) {
	if _, err := s.repo.Team(ctx, id, false); err != nil {
		return model.RatingChangePage{}, err
	}
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.RatingHistoryPage(ctx, id, p.DisciplineId, limit, offset)
}

func (s *Service) ListNotifications(ctx context.Context, p model.ListNotificationsParams) (model.NotificationPage, error) {
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.NotificationPage(ctx, currentUser(ctx), p.Unread, limit, offset)
}

func (s *Service) ReadNotification(ctx context.Context, id uuid.UUID) error {
	updated, err := s.repo.MarkNotificationRead(ctx, id, currentUser(ctx))
	if err != nil {
		return err
	}
	if !updated {
		return notFound("notification not found")
	}
	return nil
}
