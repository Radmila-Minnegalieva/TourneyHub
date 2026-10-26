package service

import (
	"context"
	"strings"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) ListAdminUsers(ctx context.Context, p model.ListAdminUsersParams) (model.ProfilePage, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.ProfilePage{}, err
	}
	return s.repo.AdminUserPage(ctx, p)
}

func protectLastAdmin(ctx context.Context, tx repository.Store, id uuid.UUID) error {
	isAdmin, err := tx.UserHasAdminRole(ctx, id)
	if err != nil {
		return err
	}
	if !isAdmin {
		return nil
	}
	count, err := tx.OtherActiveAdminCount(ctx, id)
	if err != nil {
		return err
	}
	if count == 0 {
		return conflict("cannot remove or block the last active administrator")
	}
	return nil
}

func lockAdmin(ctx context.Context, tx repository.Store, id uuid.UUID, version int) (model.Profile, error) {
	if err := tx.LockAdmin(ctx); err != nil {
		return model.Profile{}, err
	}
	actual, err := tx.LockUserVersion(ctx, id)
	if err != nil {
		return model.Profile{}, err
	}
	if actual != version {
		return model.Profile{}, conflict("stale user version")
	}
	return tx.Profile(ctx, id)
}

func (s *Service) BlockUser(ctx context.Context, id uuid.UUID, req model.BlockUserRequest) (model.Profile, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.Profile{}, err
	}
	if req.Blocked && (req.Reason == nil || strings.TrimSpace(*req.Reason) == "") {
		return model.Profile{}, invalid("block reason is required")
	}
	var result model.Profile
	err := s.withTx(ctx, func(tx repository.Store) error {
		before, err := lockAdmin(ctx, tx, id, req.Version)
		if err != nil {
			return err
		}
		if req.Blocked && !before.Blocked {
			if err := protectLastAdmin(ctx, tx, id); err != nil {
				return err
			}
		}
		if err := tx.SetUserBlocked(ctx, id, req.Blocked, req.Reason); err != nil {
			return err
		}
		if req.Blocked {
			if err := tx.RevokeUserSessions(ctx, id); err != nil {
				return err
			}
		}
		result, err = tx.Profile(ctx, id)
		if err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "user_blocked", "user", id, map[string]any{"blocked": before.Blocked}, map[string]any{"blocked": req.Blocked, "reason": req.Reason})
	})
	return result, err
}

func (s *Service) UpdateUserRoles(ctx context.Context, id uuid.UUID, req model.UpdateRolesRequest) (model.Profile, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.Profile{}, err
	}
	roles := make(map[model.Role]bool)
	for _, role := range req.Roles {
		roles[role] = true
	}
	if !roles[model.Participant] {
		return model.Profile{}, invalid("participant role is required")
	}
	var result model.Profile
	err := s.withTx(ctx, func(tx repository.Store) error {
		before, err := lockAdmin(ctx, tx, id, req.Version)
		if err != nil {
			return err
		}
		if !roles[model.Admin] && !before.Blocked {
			if err := protectLastAdmin(ctx, tx, id); err != nil {
				return err
			}
		}
		unique := make([]model.Role, 0, len(roles))
		for role := range roles {
			unique = append(unique, role)
		}
		if err := tx.ReplaceUserRoles(ctx, id, unique); err != nil {
			return err
		}
		result, err = tx.Profile(ctx, id)
		if err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "roles_updated", "user", id, map[string]any{"roles": before.Roles}, map[string]any{"roles": result.Roles})
	})
	return result, err
}

func (s *Service) ModerateTournament(ctx context.Context, id uuid.UUID, req model.ModerateTournamentRequest) (model.Tournament, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.Tournament{}, err
	}
	if strings.TrimSpace(req.Reason) == "" {
		return model.Tournament{}, invalid("moderation reason is required")
	}
	var result model.Tournament
	err := s.withTx(ctx, func(tx repository.Store) error {
		before, err := tx.ReadTournament(ctx, id, true)
		if err != nil {
			return err
		}
		if before.Version != req.Version {
			return conflict("stale tournament version")
		}
		if err := tx.SetTournamentVisibility(ctx, id, req.Hidden); err != nil {
			return err
		}
		if err := tx.Audit(ctx, currentUser(ctx), "tournament_moderated", "tournament", id, map[string]any{"hidden": before.Hidden}, map[string]any{"hidden": req.Hidden, "reason": req.Reason}); err != nil {
			return err
		}
		if err := tx.Event(ctx, id, "tournament_updated", map[string]any{"hidden": req.Hidden}); err != nil {
			return err
		}
		if err := notify(ctx, tx, before.Organizer.Id, "tournament_moderated", req.Reason, id); err != nil {
			return err
		}
		result, err = tx.ReadTournament(ctx, id, false)
		return err
	})
	return result, err
}

func (s *Service) DeleteTournament(ctx context.Context, id uuid.UUID, req model.ReasonRequest) error {
	if err := requireRole(ctx, model.Admin); err != nil {
		return err
	}
	if strings.TrimSpace(req.Reason) == "" {
		return invalid("delete reason is required")
	}
	return s.withTx(ctx, func(tx repository.Store) error {
		before, err := tx.ReadTournament(ctx, id, true)
		if err != nil {
			return err
		}
		if before.Version != req.Version {
			return conflict("stale tournament version")
		}
		if err := tx.SoftDeleteTournamentRecord(ctx, id); err != nil {
			return err
		}
		if err := tx.Event(ctx, id, "tournament_updated", map[string]any{"deleted": true}); err != nil {
			return err
		}
		if err := notify(ctx, tx, before.Organizer.Id, "tournament_deleted", req.Reason, id); err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "tournament_deleted", "tournament", id, nil, map[string]any{"reason": req.Reason})
	})
}

func (s *Service) ListAudit(ctx context.Context, p model.ListAuditParams) (model.AuditEntryPage, error) {
	if err := requireRole(ctx, model.Admin); err != nil {
		return model.AuditEntryPage{}, err
	}
	return s.repo.AuditPage(ctx, p)
}
