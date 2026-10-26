package service

import (
	"context"
	"strings"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) ListApplications(ctx context.Context, id uuid.UUID, p model.ListApplicationsParams) (model.ApplicationPage, error) {
	t, err := s.repo.ReadTournament(ctx, id, false)
	if err != nil {
		return model.ApplicationPage{}, err
	}
	if err := requireTournamentOwner(ctx, t); err != nil {
		return model.ApplicationPage{}, err
	}
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.ListApplicationPage(ctx, repository.ApplicationFilter{Limit: limit, Offset: offset, TournamentID: &id, Status: p.Status})
}

func (s *Service) ListParticipants(ctx context.Context, id uuid.UUID, p model.ListParticipantsParams) (model.ApplicationPage, error) {
	if err := tournamentPublic(ctx, s.repo, id); err != nil {
		return model.ApplicationPage{}, err
	}
	limit, offset := pagination(p.Limit, p.Offset)
	status := model.ApplicationStatusAccepted
	return s.repo.ListApplicationPage(ctx, repository.ApplicationFilter{Limit: limit, Offset: offset, TournamentID: &id, Status: &status})
}

func (s *Service) ListMyApplications(ctx context.Context, p model.ListMyApplicationsParams) (model.ApplicationPage, error) {
	limit, offset := pagination(p.Limit, p.Offset)
	id := currentUser(ctx)
	return s.repo.ListApplicationPage(ctx, repository.ApplicationFilter{Limit: limit, Offset: offset, UserID: &id})
}

func registrationOpen(t model.Tournament, now time.Time) bool {
	return t.Status == model.TournamentStatusRegistration && !t.Hidden && t.RegistrationOpensAt != nil && t.RegistrationClosesAt != nil && !now.Before(*t.RegistrationOpensAt) && now.Before(*t.RegistrationClosesAt)
}

func (s *Service) SubmitApplication(ctx context.Context, id uuid.UUID, r model.ApplicationRequest) (model.Application, error) {
	var result model.Application
	err := s.withTx(ctx, func(tx repository.Store) error {
		t, err := tx.ReadTournament(ctx, id, true)
		if err != nil {
			return err
		}
		if !registrationOpen(t, time.Now()) {
			return conflict("registration is not open")
		}
		if t.AcceptedCount >= t.MaxParticipants {
			return conflict("the tournament has no available places")
		}
		team, err := applicationTeam(ctx, tx, currentUser(ctx), id, r)
		if err != nil {
			return err
		}
		applicationID := uuid.New()
		if err := tx.CreateApplicationRecord(ctx, applicationID, id, team, currentUser(ctx)); err != nil {
			return err
		}
		if err := tx.ReserveApplicationPlayers(ctx, applicationID, id, r.PlayerIds); err != nil {
			return err
		}
		result, err = tx.ReadApplication(ctx, applicationID)
		if err != nil {
			return err
		}
		if err := applicationChanged(ctx, tx, currentUser(ctx), id, applicationID, "application_submitted", nil, result); err != nil {
			return err
		}
		return notify(ctx, tx, t.Organizer.Id, "application_submitted", "A new tournament application needs review", applicationID)
	})
	return result, err
}

func applicationTeam(ctx context.Context, tx repository.Store, userID, tournamentID uuid.UUID, r model.ApplicationRequest) (uuid.UUID, error) {
	roster, err := tx.ReadTournamentRosterRules(ctx, tournamentID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(r.PlayerIds) < roster.Min || len(r.PlayerIds) > roster.Max {
		return uuid.Nil, invalid("roster size does not match the discipline")
	}
	team := r.TeamId
	if team == nil {
		if roster.Min != 1 || roster.Max != 1 || len(r.PlayerIds) != 1 || r.PlayerIds[0] != userID {
			return uuid.Nil, invalid("an individual application must contain only the current player")
		}
		id, err := ensurePersonalTeam(ctx, tx, userID)
		if err != nil {
			return uuid.Nil, err
		}
		team = &id
	}
	captain, err := tx.ApplicationTeamCaptain(ctx, *team)
	if err != nil {
		return uuid.Nil, err
	}
	if captain != userID {
		return uuid.Nil, forbidden("only the team captain can submit an application")
	}
	if err := validateApplicationPlayers(ctx, tx, *team, r.PlayerIds); err != nil {
		return uuid.Nil, err
	}
	return *team, nil
}

func validateApplicationPlayers(ctx context.Context, q repository.Store, teamID uuid.UUID, players []uuid.UUID) error {
	seen := make(map[uuid.UUID]bool)
	for _, id := range players {
		if seen[id] {
			return invalid("players must be unique")
		}
		seen[id] = true
	}
	count, err := q.EligibleApplicationPlayersCount(ctx, teamID, players)
	if err != nil {
		return err
	}
	if count != len(players) {
		return invalid("every player must be an active verified member of the team")
	}
	return nil
}

func applicationChanged(ctx context.Context, q repository.Store, actor, tournamentID, applicationID uuid.UUID, action string, before, after any) error {
	if err := q.BumpTournamentVersion(ctx, tournamentID); err != nil {
		return err
	}
	if err := q.Audit(ctx, actor, action, "application", applicationID, before, after); err != nil {
		return err
	}
	return q.Event(ctx, tournamentID, "tournament_updated", map[string]any{"application_id": applicationID, "action": action})
}

func (s *Service) ReviewApplication(ctx context.Context, tournamentID, applicationID uuid.UUID, r model.ReviewApplicationRequest) (model.Application, error) {
	if !r.Decision.Valid() || (r.Decision == model.ReviewApplicationRequestDecisionReject && (r.Reason == nil || strings.TrimSpace(*r.Reason) == "")) {
		return model.Application{}, invalid("a valid decision and rejection reason are required")
	}
	return s.mutateApplication(ctx, tournamentID, applicationID, r.Version, "application_reviewed", func(ctx context.Context, tx repository.Store, t model.Tournament, before model.Application) error {
		if err := requireTournamentOwner(ctx, t); err != nil {
			return err
		}
		if t.Status != model.TournamentStatusRegistration || before.Status != model.ApplicationStatusPending {
			return conflict("only a pending application can be reviewed during registration")
		}
		status := model.ApplicationStatusRejected
		if r.Decision == model.ReviewApplicationRequestDecisionAccept {
			if t.AcceptedCount >= t.MaxParticipants {
				return conflict("the tournament has no available places")
			}
			status = model.ApplicationStatusAccepted
		}
		if status == model.ApplicationStatusRejected {
			if err := tx.ReleaseApplicationPlayers(ctx, applicationID); err != nil {
				return err
			}
		}
		if err := tx.SetApplicationDecision(ctx, applicationID, status, r.Reason); err != nil {
			return err
		}
		return notify(ctx, tx, before.Team.CaptainId, "application_reviewed", "Your tournament application was "+string(status), applicationID)
	})
}

func (s *Service) WithdrawApplication(ctx context.Context, tournamentID, applicationID uuid.UUID, r model.VersionRequest) (model.Application, error) {
	return s.mutateApplication(ctx, tournamentID, applicationID, r.Version, "application_withdrawn", func(ctx context.Context, tx repository.Store, t model.Tournament, before model.Application) error {
		if before.Team.CaptainId != currentUser(ctx) {
			return forbidden("only the team captain can withdraw an application")
		}
		if !registrationOpen(t, time.Now()) || (before.Status != model.ApplicationStatusPending && before.Status != model.ApplicationStatusAccepted) {
			return conflict("the application cannot be withdrawn after registration closes")
		}
		if err := tx.ReleaseApplicationPlayers(ctx, applicationID); err != nil {
			return err
		}
		if err := tx.SetApplicationDecision(ctx, applicationID, model.ApplicationStatusWithdrawn, nil); err != nil {
			return err
		}
		return notify(ctx, tx, t.Organizer.Id, "application_withdrawn", "A tournament application was withdrawn", applicationID)
	})
}

func (s *Service) mutateApplication(ctx context.Context, tournamentID, applicationID uuid.UUID, version int, action string, mutate func(context.Context, repository.Store, model.Tournament, model.Application) error) (model.Application, error) {
	var result model.Application
	err := s.withTx(ctx, func(tx repository.Store) error {
		t, err := tx.ReadTournament(ctx, tournamentID, true)
		if err != nil {
			return err
		}
		if err := tx.LockTournamentApplication(ctx, applicationID, tournamentID); err != nil {
			return err
		}
		if err := tx.LockApplicationTeam(ctx, applicationID); err != nil {
			return err
		}
		before, err := tx.ReadApplication(ctx, applicationID)
		if err != nil {
			return err
		}
		if before.Version != version {
			return conflict("application version is stale")
		}
		if err := mutate(ctx, tx, t, before); err != nil {
			return err
		}
		if err := tx.BumpApplicationVersion(ctx, applicationID); err != nil {
			return err
		}
		result, err = tx.ReadApplication(ctx, applicationID)
		if err != nil {
			return err
		}
		return applicationChanged(ctx, tx, currentUser(ctx), tournamentID, applicationID, action, before, result)
	})
	return result, err
}
