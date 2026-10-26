package service

import (
	"context"
	"strings"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func requireTournamentOwner(ctx context.Context, t model.Tournament) error {
	if hasRole(ctx, model.Admin) || (hasRole(ctx, model.Organizer) && currentUser(ctx) == t.Organizer.Id) {
		return nil
	}
	return forbidden("only the tournament organizer can perform this operation")
}

func tournamentPublic(ctx context.Context, q repository.Store, id uuid.UUID) error {
	visible, err := q.TournamentVisible(ctx, id)
	if err != nil {
		return err
	}
	if !visible {
		return notFound("tournament not found")
	}
	return nil
}

func (s *Service) ListTournaments(ctx context.Context, p model.ListTournamentsParams) (model.TournamentPage, error) {
	return s.repo.ListTournamentPage(ctx, p, nil, false)
}

func (s *Service) ListMyTournaments(ctx context.Context, p model.ListMyTournamentsParams) (model.TournamentPage, error) {
	id := currentUser(ctx)
	return s.repo.ListTournamentPage(ctx, model.ListTournamentsParams(p), &id, true)
}

func (s *Service) ListAdminTournaments(ctx context.Context, p model.ListAdminTournamentsParams) (model.TournamentPage, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.TournamentPage{}, err
	}
	return s.repo.ListTournamentPage(ctx, model.ListTournamentsParams(p), nil, true)
}

func (s *Service) GetTournament(ctx context.Context, id uuid.UUID) (model.Tournament, error) {
	if err := tournamentPublic(ctx, s.repo, id); err != nil {
		return model.Tournament{}, err
	}
	return s.repo.ReadTournament(ctx, id, false)
}

func validateTournamentRequest(r model.CreateTournamentRequest, publishing bool) error {
	if strings.TrimSpace(r.Name) == "" || !r.Format.Valid() || r.MaxParticipants < 2 || r.MaxParticipants > 256 {
		return invalid("invalid tournament name, format or participant limit")
	}
	if r.Timezone == "" || r.Timezone == "Local" {
		return invalid("timezone must be an IANA identifier")
	}
	if _, err := time.LoadLocation(r.Timezone); err != nil {
		return invalid("timezone must be an IANA identifier")
	}
	if err := validateTournamentDates(r, publishing); err != nil {
		return err
	}
	return validateTournamentRules(r.Format, r.MaxParticipants, r.Rules)
}

func validateTournamentDates(r model.CreateTournamentRequest, required bool) error {
	dates := []*time.Time{r.RegistrationOpensAt, r.RegistrationClosesAt, r.StartsAt, r.EndsAt}
	for _, date := range dates {
		if required && date == nil {
			return invalid("publication requires registration and tournament dates")
		}
	}
	for i := 1; i < len(dates); i++ {
		previous, next := dates[i-1], dates[i]
		if previous == nil || next == nil {
			continue
		}
		if previous.After(*next) || (i != 2 && previous.Equal(*next)) {
			return invalid("dates must satisfy opens < closes <= starts < ends")
		}
	}
	return nil
}

func validateTournamentRules(format model.TournamentFormat, max int, r model.Rules) error {
	if !r.BestOf.Valid() || !r.RoundRobinLegs.Valid() || len(r.TieBreakers) == 0 || r.ConfirmationTimeoutHours < 1 || r.ConfirmationTimeoutHours > 168 {
		return invalid("invalid tournament rules")
	}
	for _, points := range []int{r.WinPoints, r.DrawPoints, r.LossPoints} {
		if points < 0 || points > 100 {
			return invalid("points must be between 0 and 100")
		}
	}
	seen := make(map[model.TieBreaker]bool)
	for _, tie := range r.TieBreakers {
		if !tie.Valid() || seen[tie] {
			return invalid("tie breakers must be valid and unique")
		}
		seen[tie] = true
	}
	if format == model.TournamentFormatRoundRobin && r.ThirdPlaceMatch {
		return invalid("a round robin has no third place match")
	}
	if err := validateGroupSettings(format, max, r.GroupSettings); err != nil {
		return err
	}
	if r.ThirdPlaceMatch {
		count := max
		if r.GroupSettings != nil {
			g := r.GroupSettings
			count = g.GroupCount*g.QualifiersPerGroup + g.WildcardCount
		}
		if count < 4 {
			return invalid("third-place match needs at least four playoff participants")
		}
	}
	return nil
}

func validateGroupSettings(format model.TournamentFormat, max int, g *model.GroupSettings) error {
	if format != model.TournamentFormatGroupsPlayoffs {
		if g != nil {
			return invalid("group settings apply only to groups_playoffs")
		}
		return nil
	}
	if g == nil || g.GroupCount < 2 || g.GroupCount > 64 || g.QualifiersPerGroup < 1 || g.QualifiersPerGroup > 128 {
		return invalid("groups_playoffs requires valid group settings")
	}
	if g.GroupCount*2 > max || g.GroupCount*g.QualifiersPerGroup+g.WildcardCount > max || g.WildcardCount < 0 || g.WildcardCount > g.GroupCount {
		return invalid("group configuration exceeds the participant limit")
	}
	if g.WildcardCount > 0 && (g.WildcardPlace == nil || *g.WildcardPlace <= g.QualifiersPerGroup || *g.WildcardPlace > max/g.GroupCount) {
		return invalid("wildcard place must follow automatic qualifying places and exist in each group")
	}
	return nil
}

func tournamentRequest(t model.Tournament) model.CreateTournamentRequest {
	return model.CreateTournamentRequest{Name: t.Name, Description: t.Description, DisciplineId: t.DisciplineId, Format: t.Format, MaxParticipants: t.MaxParticipants, Timezone: t.Timezone, Rules: t.Rules, RegistrationOpensAt: t.RegistrationOpensAt, RegistrationClosesAt: t.RegistrationClosesAt, StartsAt: t.StartsAt, EndsAt: t.EndsAt}
}

func replacePresent[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

func replacePresentPointer[T any](dst **T, src *T) {
	if src != nil {
		*dst = src
	}
}

func mergeTournament(t model.Tournament, p model.UpdateTournamentRequest) model.CreateTournamentRequest {
	r := tournamentRequest(t)
	replacePresent(&r.Name, p.Name)
	replacePresent(&r.DisciplineId, p.DisciplineId)
	replacePresent(&r.Format, p.Format)
	replacePresent(&r.MaxParticipants, p.MaxParticipants)
	replacePresent(&r.Timezone, p.Timezone)
	replacePresent(&r.Rules, p.Rules)
	replacePresentPointer(&r.Description, p.Description)
	replacePresentPointer(&r.RegistrationOpensAt, p.RegistrationOpensAt)
	replacePresentPointer(&r.RegistrationClosesAt, p.RegistrationClosesAt)
	replacePresentPointer(&r.StartsAt, p.StartsAt)
	replacePresentPointer(&r.EndsAt, p.EndsAt)
	return r
}

func (s *Service) CreateTournament(ctx context.Context, r model.CreateTournamentRequest) (model.Tournament, error) {
	if err := requireRole(ctx, model.Organizer); err != nil {
		return model.Tournament{}, err
	}
	if err := validateTournamentRequest(r, false); err != nil {
		return model.Tournament{}, err
	}
	var result model.Tournament
	err := s.withTx(ctx, func(tx repository.Store) error {
		roster, err := tx.ReadDisciplineRosterRules(ctx, r.DisciplineId)
		if err != nil {
			return err
		}
		id := uuid.New()
		if err := tx.CreateTournamentRecord(ctx, id, currentUser(ctx), r, roster); err != nil {
			return err
		}
		result, err = tx.ReadTournament(ctx, id, false)
		if err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "tournament_created", "tournament", id, nil, result)
	})
	return result, err
}

func (s *Service) UpdateTournament(ctx context.Context, id uuid.UUID, p model.UpdateTournamentRequest) (model.Tournament, error) {
	return s.mutateTournament(ctx, id, p.Version, "tournament_updated", func(ctx context.Context, tx repository.Store, before model.Tournament) error {
		if before.Status != model.TournamentStatusDraft && before.Status != model.TournamentStatusRegistration {
			return conflict("only a draft or registration tournament can be edited")
		}
		r := mergeTournament(before, p)
		if err := validateTournamentRequest(r, before.Status == model.TournamentStatusRegistration); err != nil {
			return err
		}
		if r.MaxParticipants < before.AcceptedCount {
			return conflict("participant limit is below the accepted participant count")
		}
		return updateTournamentFields(ctx, tx, id, r, r.DisciplineId != before.DisciplineId)
	})
}

func updateTournamentFields(ctx context.Context, tx repository.Store, id uuid.UUID, r model.CreateTournamentRequest, refreshDiscipline bool) error {
	var roster repository.RosterRules
	var err error
	if refreshDiscipline {
		roster, err = tx.ReadDisciplineRosterRules(ctx, r.DisciplineId)
	} else {
		roster, err = tx.ReadTournamentRosterRules(ctx, id)
	}
	if err != nil {
		return err
	}
	compatible, err := tx.TournamentRosterCompatible(ctx, id, roster.Min, roster.Max)
	if err != nil {
		return err
	}
	if !compatible {
		return conflict("existing application rosters do not match the new discipline")
	}
	return tx.UpdateTournamentRecord(ctx, id, r, roster)
}

func (s *Service) mutateTournament(ctx context.Context, id uuid.UUID, version int, action string, mutate func(context.Context, repository.Store, model.Tournament) error) (model.Tournament, error) {
	var result model.Tournament
	err := s.withTx(ctx, func(tx repository.Store) error {
		before, err := tx.ReadTournament(ctx, id, true)
		if err != nil {
			return err
		}
		if err := requireTournamentOwner(ctx, before); err != nil {
			return err
		}
		if before.Version != version {
			return conflict("tournament version is stale")
		}
		if err := mutate(ctx, tx, before); err != nil {
			return err
		}
		if err := tx.BumpTournamentVersion(ctx, id); err != nil {
			return err
		}
		result, err = tx.ReadTournament(ctx, id, false)
		if err != nil {
			return err
		}
		if err := tx.Audit(ctx, currentUser(ctx), action, "tournament", id, before, result); err != nil {
			return err
		}
		return tx.Event(ctx, id, "tournament_updated", result)
	})
	return result, err
}

func (s *Service) PublishTournament(ctx context.Context, id uuid.UUID, r model.VersionRequest) (model.Tournament, error) {
	return s.mutateTournament(ctx, id, r.Version, "tournament_published", func(ctx context.Context, tx repository.Store, t model.Tournament) error {
		if t.Status != model.TournamentStatusDraft {
			return conflict("only a draft can be published")
		}
		if err := validateTournamentRequest(tournamentRequest(t), true); err != nil {
			return err
		}
		return tx.SetTournamentStatus(ctx, id, model.TournamentStatusRegistration)
	})
}

func (s *Service) CancelTournament(ctx context.Context, id uuid.UUID, r model.ReasonRequest) (model.Tournament, error) {
	if strings.TrimSpace(r.Reason) == "" {
		return model.Tournament{}, invalid("cancellation reason is required")
	}
	return s.mutateTournament(ctx, id, r.Version, "tournament_cancelled", func(ctx context.Context, tx repository.Store, t model.Tournament) error {
		if t.Status == model.TournamentStatusCompleted || t.Status == model.TournamentStatusCancelled {
			return conflict("the tournament cannot be cancelled in its current state")
		}
		if err := tx.SetTournamentStatus(ctx, id, model.TournamentStatusCancelled); err != nil {
			return err
		}
		return notifyTournamentPlayers(ctx, tx, id, "tournament_cancelled", r.Reason)
	})
}

func notifyTournamentPlayers(ctx context.Context, q repository.Store, id uuid.UUID, kind, message string) error {
	users, err := q.TournamentPlayerIDs(ctx, id)
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := notify(ctx, q, user, kind, message, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CompleteTournament(ctx context.Context, id uuid.UUID, r model.VersionRequest) (model.Tournament, error) {
	return s.mutateTournament(ctx, id, r.Version, "tournament_completed", func(ctx context.Context, tx repository.Store, t model.Tournament) error {
		if t.Status != model.TournamentStatusInProgress {
			return conflict("only an active tournament can be completed")
		}
		return finalizeTournament(ctx, tx, id)
	})
}

func (s *Service) AssignTournamentReferee(ctx context.Context, id uuid.UUID, r model.AssignRefereeRequest) (model.Tournament, error) {
	return s.mutateTournament(ctx, id, r.Version, "tournament_referee_assigned", func(ctx context.Context, tx repository.Store, t model.Tournament) error {
		if t.Status == model.TournamentStatusCancelled || t.Status == model.TournamentStatusCompleted {
			return conflict("the tournament is no longer active")
		}
		if err := tx.LockSchedule(ctx); err != nil {
			return err
		}
		if err := validateReferee(ctx, tx, r.RefereeId); err != nil {
			return err
		}
		if err := tx.SetTournamentReferee(ctx, id, r.RefereeId); err != nil {
			return err
		}
		t.RefereeId = &r.RefereeId
		if err := validateTournamentRefereeSchedule(ctx, tx, t); err != nil {
			return err
		}
		return notify(ctx, tx, r.RefereeId, "referee_assigned", "You have been assigned as tournament referee", id)
	})
}

func validateTournamentRefereeSchedule(ctx context.Context, tx repository.Store, t model.Tournament) error {
	matches, err := tx.MatchesUsingTournamentReferee(ctx, t.Id)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if err := scheduleConflict(ctx, tx, t, match, *match.StartsAt); err != nil {
			return err
		}
	}
	return nil
}
