package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/security"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func captainTeam(ctx context.Context, q repository.Store, teamID, userID uuid.UUID) (model.Team, error) {
	team, err := q.Team(ctx, teamID, true)
	if err != nil {
		return team, err
	}
	if team.CaptainId != userID {
		return team, forbidden("only the team captain can perform this action")
	}
	return team, nil
}

func teamVersion(team model.Team, expected int) error {
	if team.Version != expected {
		return conflict("team version changed")
	}
	return nil
}

func validTeamLogo(ctx context.Context, q repository.Store, fileID *uuid.UUID, ownerID uuid.UUID) error {
	if fileID == nil {
		return nil
	}
	allowed, err := q.OwnsFile(ctx, *fileID, ownerID, "team_logo")
	if err != nil {
		return err
	}
	if !allowed {
		return invalid("team logo must be an owned team_logo file")
	}
	return nil
}

// ensurePersonalTeam создаёт команду из одного игрока для индивидуальной заявки.
// Блокировка пользователя сериализует параллельные первые заявки.
func ensurePersonalTeam(ctx context.Context, q repository.Store, userID uuid.UUID) (uuid.UUID, error) {
	nickname, err := q.LockUserNickname(ctx, userID)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := q.PersonalTeam(ctx, userID)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return uuid.Nil, err
	}
	id, err = q.CreateTeamRecord(ctx, nickname, nil, nil, userID, true)
	if err != nil {
		return uuid.Nil, err
	}
	_, err = q.AddTeamMember(ctx, id, userID)
	return id, err
}

func (s *Service) ListTeams(ctx context.Context, params model.ListTeamsParams) (model.TeamPage, error) {
	search := ""
	if params.Search != nil {
		search = *params.Search
	}
	limit, offset := pagination(params.Limit, params.Offset)
	return s.repo.TeamsPage(ctx, search, nil, limit, offset)
}

func (s *Service) ListMyTeams(ctx context.Context, params model.ListMyTeamsParams) (model.TeamPage, error) {
	limit, offset := pagination(params.Limit, params.Offset)
	owner := currentUser(ctx)
	return s.repo.TeamsPage(ctx, "", &owner, limit, offset)
}

func (s *Service) GetTeam(ctx context.Context, id uuid.UUID) (model.Team, error) {
	return s.repo.Team(ctx, id, false)
}

func (s *Service) CreateTeam(ctx context.Context, req model.CreateTeamRequest) (model.Team, error) {
	var team model.Team
	if err := requireRole(ctx, model.Participant); err != nil {
		return team, err
	}
	if strings.TrimSpace(req.Name) == "" {
		return team, invalid("team name cannot be blank")
	}
	err := s.withTx(ctx, func(q repository.Store) error {
		owner := currentUser(ctx)
		if err := validTeamLogo(ctx, q, req.LogoFileId, owner); err != nil {
			return err
		}
		id, err := q.CreateTeamRecord(ctx, req.Name, req.Description, req.LogoFileId, owner, false)
		if err != nil {
			return err
		}
		if _, err := q.AddTeamMember(ctx, id, owner); err != nil {
			return err
		}
		team, err = q.Team(ctx, id, false)
		if err != nil {
			return err
		}
		return q.Audit(ctx, owner, "team_created", "team", id, nil, team)
	})
	return team, err
}

func (s *Service) UpdateTeam(ctx context.Context, teamID uuid.UUID, req model.UpdateTeamRequest) (model.Team, error) {
	var team model.Team
	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		return team, invalid("team name cannot be blank")
	}
	err := s.withTx(ctx, func(q repository.Store) error {
		actor := currentUser(ctx)
		before, err := captainTeam(ctx, q, teamID, actor)
		if err != nil {
			return err
		}
		if err := teamVersion(before, req.Version); err != nil {
			return err
		}
		if err := validTeamLogo(ctx, q, req.LogoFileId, actor); err != nil {
			return err
		}
		if err := q.UpdateTeamRecord(ctx, teamID, req); err != nil {
			return err
		}
		team, err = q.Team(ctx, teamID, false)
		if err != nil {
			return err
		}
		return q.Audit(ctx, actor, "team_updated", "team", teamID, before, team)
	})
	return team, err
}

func (s *Service) DeleteTeam(ctx context.Context, teamID uuid.UUID, req model.VersionRequest) error {
	return s.withTx(ctx, func(q repository.Store) error {
		actor := currentUser(ctx)
		team, err := captainTeam(ctx, q, teamID, actor)
		if err != nil {
			return err
		}
		if err := teamVersion(team, req.Version); err != nil {
			return err
		}
		active, err := q.TeamHasActiveApplications(ctx, teamID)
		if err != nil {
			return err
		}
		if active {
			return conflict("team has accepted applications in active tournaments")
		}
		if err := q.WithdrawPendingTeamApplications(ctx, teamID); err != nil {
			return err
		}
		if err := q.RevokeTeamInvitations(ctx, teamID); err != nil {
			return err
		}
		if err := q.SoftDeleteTeam(ctx, teamID); err != nil {
			return err
		}
		return q.Audit(ctx, actor, "team_deleted", "team", teamID, team, nil)
	})
}

func (s *Service) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]model.TeamMember, error) {
	if _, err := s.repo.Team(ctx, teamID, false); err != nil {
		return nil, err
	}
	return s.repo.TeamMembers(ctx, teamID)
}

func (s *Service) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID, req model.VersionRequest) error {
	return s.withTx(ctx, func(q repository.Store) error {
		actor := currentUser(ctx)
		team, err := q.Team(ctx, teamID, true)
		if err != nil {
			return err
		}
		if actor != team.CaptainId && actor != userID {
			return forbidden("only captain or the member can remove membership")
		}
		if team.CaptainId == userID {
			return conflict("captain must transfer captaincy before leaving")
		}
		if err := teamVersion(team, req.Version); err != nil {
			return err
		}
		deleted, err := q.DeleteTeamMember(ctx, teamID, userID)
		if err != nil {
			return err
		}
		if !deleted {
			return notFound("team member not found")
		}
		if err := q.BumpTeamVersion(ctx, teamID); err != nil {
			return err
		}
		return q.Audit(ctx, actor, "team_member_removed", "team", teamID, map[string]any{"user_id": userID}, nil)
	})
}

func (s *Service) TransferCaptain(ctx context.Context, teamID uuid.UUID, req model.TransferCaptainRequest) (model.Team, error) {
	var team model.Team
	err := s.withTx(ctx, func(q repository.Store) error {
		actor := currentUser(ctx)
		before, err := captainTeam(ctx, q, teamID, actor)
		if err != nil {
			return err
		}
		if before.Personal {
			return conflict("personal team cannot change captain")
		}
		if err := teamVersion(before, req.Version); err != nil {
			return err
		}
		allowed, err := q.ActiveTeamMember(ctx, teamID, req.UserId)
		if err != nil {
			return err
		}
		if !allowed {
			return invalid("new captain must be an unblocked team member")
		}
		if err := q.TransferTeamCaptain(ctx, teamID, req.UserId); err != nil {
			return err
		}
		team, err = q.Team(ctx, teamID, false)
		if err != nil {
			return err
		}
		return q.Audit(ctx, actor, "team_captain_transferred", "team", teamID, before, team)
	})
	return team, err
}

func invitationURL(webURL string, id uuid.UUID, token string) string {
	return strings.TrimRight(webURL, "/") + "/invitations/" + id.String() + "?token=" + url.QueryEscape(token)
}

func resolveInvitee(ctx context.Context, q repository.Store, teamID uuid.UUID, nickname *string) (*uuid.UUID, error) {
	if nickname == nil {
		return nil, nil
	}
	invitee, err := q.InviteeByNickname(ctx, *nickname)
	if err != nil {
		return nil, err
	}
	member, err := q.IsTeamMember(ctx, teamID, invitee)
	if err != nil {
		return nil, err
	}
	if member {
		return nil, conflict("player is already a team member")
	}
	return &invitee, nil
}

func (s *Service) CreateInvitation(ctx context.Context, teamID uuid.UUID, req model.CreateInvitationRequest) (model.InvitationLink, error) {
	var result model.InvitationLink
	token, err := security.NewToken()
	if err != nil {
		return result, err
	}
	err = s.withTx(ctx, func(q repository.Store) error {
		actor := currentUser(ctx)
		team, err := captainTeam(ctx, q, teamID, actor)
		if err != nil {
			return err
		}
		if team.Personal {
			return conflict("personal team can contain only its owner")
		}
		invitee, err := resolveInvitee(ctx, q, teamID, req.Nickname)
		if err != nil {
			return err
		}
		result.Invitation, err = q.CreateInvitationRecord(ctx, teamID, invitee, actor, security.TokenHash(token))
		if err != nil {
			return err
		}
		if invitee != nil {
			return notify(ctx, q, *invitee, "team_invitation", "Приглашение в команду "+team.Name, result.Invitation.Id)
		}
		return nil
	})
	result.Token, result.Url = token, invitationURL(s.cfg.Public.WebURL, result.Invitation.Id, token)
	return result, err
}

func (s *Service) ListMyInvitations(ctx context.Context, params model.ListMyInvitationsParams) (model.InvitationPage, error) {
	limit, offset := pagination(params.Limit, params.Offset)
	return s.repo.InvitationsPage(ctx, currentUser(ctx), limit, offset)
}

func (s *Service) DecideInvitation(ctx context.Context, invitationID uuid.UUID, req model.InvitationDecisionRequest) (model.Invitation, error) {
	var invitation model.Invitation
	err := s.withTx(ctx, func(q repository.Store) error {
		actor := currentUser(ctx)
		state, err := lockInvitationDecision(ctx, q, invitationID)
		if err != nil {
			return err
		}
		if err := state.validate(actor, req.Token, time.Now()); err != nil {
			return err
		}
		invitation, err = state.apply(ctx, q, actor, req.Decision)
		return err
	})
	return invitation, err
}

type lockedInvitation struct {
	invitation model.Invitation
	team       model.Team
	tokenHash  string
}

func lockInvitationDecision(ctx context.Context, q repository.Store, id uuid.UUID) (lockedInvitation, error) {
	var state lockedInvitation
	teamID, err := q.InvitationTeam(ctx, id)
	if err != nil {
		return state, err
	}
	// Единый порядок team → invitation сохраняет сериализацию с отзывом приглашения.
	state.team, err = q.Team(ctx, teamID, true)
	if err != nil {
		return state, err
	}
	record, err := q.LockInvitation(ctx, id)
	state.invitation, state.tokenHash = record.Invitation, record.TokenHash
	return state, err
}

func authorizeInvitation(invitation model.Invitation, hash string, actor uuid.UUID, token *string) error {
	if invitation.InviteeId != nil {
		if *invitation.InviteeId == actor {
			return nil
		}
		return forbidden("invitation belongs to another player")
	}
	if token == nil || subtle.ConstantTimeCompare([]byte(hash), []byte(security.TokenHash(*token))) != 1 {
		return forbidden("invalid invitation token")
	}
	return nil
}

func (state lockedInvitation) validate(actor uuid.UUID, token *string, now time.Time) error {
	if err := authorizeInvitation(state.invitation, state.tokenHash, actor, token); err != nil {
		return err
	}
	if state.invitation.Status != model.InvitationStatusPending || !state.invitation.ExpiresAt.After(now) {
		return conflict("invitation is already consumed or expired")
	}
	if state.team.Personal {
		return conflict("personal team cannot accept members")
	}
	return nil
}

func addInvitedMember(ctx context.Context, q repository.Store, teamID, actor uuid.UUID) error {
	inserted, err := q.AddTeamMember(ctx, teamID, actor)
	if err != nil {
		return err
	}
	if !inserted {
		return conflict("player is already a team member")
	}
	return q.BumpTeamVersion(ctx, teamID)
}

func (state lockedInvitation) apply(ctx context.Context, q repository.Store, actor uuid.UUID, decision model.InvitationDecisionRequestDecision) (model.Invitation, error) {
	invitation := state.invitation
	invitation.Status = model.InvitationStatusDeclined
	if decision == model.InvitationDecisionRequestDecisionAccept {
		if err := addInvitedMember(ctx, q, state.team.Id, actor); err != nil {
			return invitation, err
		}
		invitation.Status = model.InvitationStatusAccepted
	}
	if err := q.SetInvitationDecision(ctx, invitation.Id, invitation.Status, actor); err != nil {
		return invitation, err
	}
	err := notify(ctx, q, state.team.CaptainId, "invitation_decided", "Ответ на приглашение в команду "+state.team.Name, invitation.Id)
	return invitation, err
}

func (s *Service) RevokeInvitation(ctx context.Context, teamID, invitationID uuid.UUID) error {
	return s.withTx(ctx, func(q repository.Store) error {
		if _, err := captainTeam(ctx, q, teamID, currentUser(ctx)); err != nil {
			return err
		}
		status, err := q.LockTeamInvitationStatus(ctx, invitationID, teamID)
		if err != nil {
			return err
		}
		if status != "pending" && status != "revoked" {
			return conflict("invitation has already been consumed")
		}
		return q.RevokeInvitationRecord(ctx, invitationID)
	})
}
