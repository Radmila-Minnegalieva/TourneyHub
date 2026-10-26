package service

import (
	"context"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) matchTransaction(ctx context.Context, id uuid.UUID, fn func(repository.Store, model.Tournament, model.Match) error) error {
	tournamentID, err := s.repo.MatchTournamentID(ctx, id)
	if err != nil {
		return err
	}
	return s.withTx(ctx, func(tx repository.Store) error {
		tournament, err := tx.ReadTournament(ctx, tournamentID, true)
		if err != nil {
			return err
		}
		match, err := tx.ReadMatch(ctx, id, true)
		if err != nil {
			return err
		}
		return fn(tx, tournament, match)
	})
}
func matchVersion(m model.Match, version int) error {
	if m.Version != version {
		return conflict("match version is stale")
	}
	return nil
}
func tournamentPlaying(t model.Tournament) error {
	if t.Status != model.TournamentStatusInProgress {
		return conflict("tournament is not in progress")
	}
	return nil
}
func (s *Service) GetMatch(ctx context.Context, id uuid.UUID) (model.Match, error) {
	m, err := s.repo.ReadMatch(ctx, id, false)
	if err != nil {
		return m, err
	}
	if err = tournamentPublic(ctx, s.repo, m.TournamentId); err != nil {
		return m, err
	}
	return m, nil
}
func (s *Service) ListMatches(ctx context.Context, id uuid.UUID, p model.ListMatchesParams) (model.MatchPage, error) {
	if err := tournamentPublic(ctx, s.repo, id); err != nil {
		return model.MatchPage{}, err
	}
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.ReadMatchPage(ctx, id, p, limit, offset)
}
func (s *Service) GetBracket(ctx context.Context, id uuid.UUID) (model.Bracket, error) {
	if err := tournamentPublic(ctx, s.repo, id); err != nil {
		return model.Bracket{}, err
	}
	return s.repo.ReadBracket(ctx, id)
}
func matchCaptain(ctx context.Context, q repository.Store, m model.Match, user uuid.UUID) (int, error) {
	return q.MatchCaptainSide(ctx, m.HomeApplicationId, m.AwayApplicationId, user)
}
func matchReadAccess(ctx context.Context, q repository.Store, t model.Tournament, m model.Match) error {
	user := currentUser(ctx)
	if hasRole(ctx, model.Admin) || t.Organizer.Id == user || assignedReferee(t, m, user) {
		return nil
	}
	side, err := matchCaptain(ctx, q, m, user)
	if err != nil {
		return err
	}
	if side > 0 {
		return nil
	}
	member, err := q.IsMatchRosterPlayer(ctx, m.HomeApplicationId, m.AwayApplicationId, user)
	if err != nil {
		return err
	}
	if !member {
		return forbidden("match participants only")
	}
	return nil
}
func assignedReferee(t model.Tournament, m model.Match, user uuid.UUID) bool {
	if m.RefereeId != nil {
		return *m.RefereeId == user
	}
	return t.RefereeId != nil && *t.RefereeId == user
}
func requireMatchReferee(ctx context.Context, t model.Tournament, m model.Match) error {
	if hasRole(ctx, model.Admin) {
		return nil
	}
	if !hasRole(ctx, model.Referee) || !assignedReferee(t, m, currentUser(ctx)) {
		return forbidden("assigned referee only")
	}
	return nil
}
func notifyMatch(ctx context.Context, q repository.Store, m model.Match, kind, message string) error {
	ids, err := q.MatchNotificationRecipients(ctx, m.HomeApplicationId, m.AwayApplicationId)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err = notify(ctx, q, id, kind, message, m.Id); err != nil {
			return err
		}
	}
	return nil
}
func matchChanged(ctx context.Context, q repository.Store, actor uuid.UUID, m model.Match, action string) error {
	if err := q.Audit(ctx, actor, action, "match", m.Id, nil, m); err != nil {
		return err
	}
	return q.Event(ctx, m.TournamentId, "match_updated", m)
}
func requireScheduled(m model.Match) error {
	if m.Status != model.MatchStatusScheduled || m.HomeApplicationId == nil || m.AwayApplicationId == nil {
		return conflict("match is not ready for a result")
	}
	if m.StartsAt == nil || m.StartsAt.After(time.Now()) {
		return conflict("match has not started")
	}
	return nil
}
func optionalUUIDEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
