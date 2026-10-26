package service

import (
	"context"
	"strings"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

// Один матч резервирует час; интервалы стартов разных участников могут быть меньше.
const matchSlot = time.Hour

func scheduleAllowed(t model.Tournament, m model.Match, start time.Time, venue, url *string) error {
	if err := tournamentPlaying(t); err != nil {
		return err
	}
	if m.Status != model.MatchStatusScheduled {
		return conflict("only scheduled matches can be moved")
	}
	if (venue == nil || strings.TrimSpace(*venue) == "") && (url == nil || strings.TrimSpace(*url) == "") {
		return invalid("venue or online_url is required")
	}
	if t.StartsAt == nil || t.EndsAt == nil || start.Before(*t.StartsAt) || start.Add(matchSlot).After(*t.EndsAt) {
		return invalid("match must fit inside tournament dates")
	}
	return nil
}
func scheduleConflict(ctx context.Context, q repository.Store, t model.Tournament, m model.Match, start time.Time) error {
	if err := scheduleDependencyConflict(ctx, q, m, start); err != nil {
		return err
	}
	referee := m.RefereeId
	if referee == nil {
		referee = t.RefereeId
	}
	overlap, err := q.HasParticipantScheduleConflict(ctx, m, start, referee)
	if err != nil {
		return err
	}
	if overlap {
		return conflict("participant or referee already has a match in this one-hour slot")
	}
	return nil
}

func writeSchedule(ctx context.Context, q repository.Store, t model.Tournament, m model.Match, start time.Time, venue, url *string) (model.Match, error) {
	start = start.UTC()
	if err := scheduleAllowed(t, m, start, venue, url); err != nil {
		return m, err
	}
	if err := scheduleConflict(ctx, q, t, m, start); err != nil {
		return m, err
	}
	if err := q.SetMatchSchedule(ctx, m.Id, &start, venue, url); err != nil {
		return m, err
	}
	m.StartsAt = &start
	m.Venue = venue
	m.OnlineUrl = url
	m.Version++
	if err := matchChanged(ctx, q, currentUser(ctx), m, "match.scheduled"); err != nil {
		return m, err
	}
	return m, notifyMatch(ctx, q, m, "match_scheduled", "Назначено или изменено время матча")
}
func (s *Service) ScheduleMatch(ctx context.Context, id uuid.UUID, req model.ScheduleMatchRequest) (model.Match, error) {
	var result model.Match
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		if err := requireTournamentOwner(ctx, t); err != nil {
			return err
		}
		if err := matchVersion(m, req.Version); err != nil {
			return err
		}
		if err := tx.LockSchedule(ctx); err != nil {
			return err
		}
		var err error
		result, err = writeSchedule(ctx, tx, t, m, req.StartsAt, req.Venue, req.OnlineUrl)
		return err
	})
	return result, err
}
func (s *Service) ScheduleRound(ctx context.Context, id uuid.UUID, req model.ScheduleRoundRequest) ([]model.Match, error) {
	result := []model.Match{}
	err := s.withTx(ctx, func(tx repository.Store) error {
		t, err := tx.ReadTournament(ctx, id, true)
		if err != nil {
			return err
		}
		if err = requireTournamentOwner(ctx, t); err != nil {
			return err
		}
		if t.Version != req.Version {
			return conflict("tournament version is stale")
		}
		if err = tx.LockSchedule(ctx); err != nil {
			return err
		}
		matches, err := tx.RoundMatches(ctx, id, req.StageId, req.Round)
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return notFound("round has no matches")
		}
		// Освобождаем старые слоты всей пачки до проверки перестановки.
		if err = tx.ClearRoundSchedule(ctx, id, req.StageId, req.Round); err != nil {
			return err
		}
		for i, m := range matches {
			start := req.StartsAt.Add(time.Duration(i*req.IntervalMinutes) * time.Minute)
			updated, err := writeSchedule(ctx, tx, t, m, start, &req.Venue, nil)
			if err != nil {
				return err
			}
			result = append(result, updated)
		}
		return tx.BumpTournamentVersion(ctx, id)
	})
	return result, err
}
func validateReferee(ctx context.Context, q repository.Store, id uuid.UUID) error {
	valid, err := q.ValidReferee(ctx, id)
	if err != nil {
		return err
	}
	if !valid {
		return invalid("referee must be an active verified user with referee role")
	}
	return nil
}
func (s *Service) AssignMatchReferee(ctx context.Context, id uuid.UUID, req model.AssignRefereeRequest) (model.Match, error) {
	var result model.Match
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		if err := requireTournamentOwner(ctx, t); err != nil {
			return err
		}
		if err := matchVersion(m, req.Version); err != nil {
			return err
		}
		if m.Status == model.MatchStatusCompleted || m.Status == model.MatchStatusBye {
			return conflict("cannot reassign a finished match")
		}
		if err := tx.LockSchedule(ctx); err != nil {
			return err
		}
		if err := validateReferee(ctx, tx, req.RefereeId); err != nil {
			return err
		}
		m.RefereeId = &req.RefereeId
		if m.StartsAt != nil {
			if err := scheduleConflict(ctx, tx, t, m, *m.StartsAt); err != nil {
				return err
			}
		}
		if err := tx.SetMatchReferee(ctx, id, req.RefereeId); err != nil {
			return err
		}
		m.Version++
		result = m
		if err := matchChanged(ctx, tx, currentUser(ctx), m, "match.referee_assigned"); err != nil {
			return err
		}
		return notify(ctx, tx, req.RefereeId, "referee_assigned", "Вы назначены судьёй матча", id)
	})
	return result, err
}

// Связи сетки резервируют порядок слотов ещё до определения участников.
func scheduleDependencyConflict(ctx context.Context, q repository.Store, m model.Match, start time.Time) error {
	overlap, err := q.HasScheduleDependencyConflict(ctx, m, start)
	if err != nil {
		return err
	}
	if overlap {
		return conflict("a dependent match must start after the preceding one-hour slot")
	}
	return nil
}
