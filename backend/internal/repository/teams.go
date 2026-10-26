package repository

import (
	"context"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

type InvitationSecret struct {
	Invitation model.Invitation
	TokenHash  string
}

// Teams содержит типизированные операции хранения команд и приглашений.
type Teams interface {
	Team(context.Context, uuid.UUID, bool) (model.Team, error)
	LockUserNickname(context.Context, uuid.UUID) (string, error)
	PersonalTeam(context.Context, uuid.UUID) (uuid.UUID, error)
	CreateTeamRecord(context.Context, string, *string, *uuid.UUID, uuid.UUID, bool) (uuid.UUID, error)
	AddTeamMember(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	TeamsPage(context.Context, string, *uuid.UUID, int, int) (model.TeamPage, error)
	UpdateTeamRecord(context.Context, uuid.UUID, model.UpdateTeamRequest) error
	TeamHasActiveApplications(context.Context, uuid.UUID) (bool, error)
	WithdrawPendingTeamApplications(context.Context, uuid.UUID) error
	RevokeTeamInvitations(context.Context, uuid.UUID) error
	SoftDeleteTeam(context.Context, uuid.UUID) error
	TeamMembers(context.Context, uuid.UUID) ([]model.TeamMember, error)
	DeleteTeamMember(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	BumpTeamVersion(context.Context, uuid.UUID) error
	ActiveTeamMember(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	TransferTeamCaptain(context.Context, uuid.UUID, uuid.UUID) error
	InviteeByNickname(context.Context, string) (uuid.UUID, error)
	IsTeamMember(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	CreateInvitationRecord(context.Context, uuid.UUID, *uuid.UUID, uuid.UUID, string) (model.Invitation, error)
	InvitationsPage(context.Context, uuid.UUID, int, int) (model.InvitationPage, error)
	InvitationTeam(context.Context, uuid.UUID) (uuid.UUID, error)
	LockInvitation(context.Context, uuid.UUID) (InvitationSecret, error)
	SetInvitationDecision(context.Context, uuid.UUID, model.InvitationStatus, uuid.UUID) error
	LockTeamInvitationStatus(context.Context, uuid.UUID, uuid.UUID) (string, error)
	RevokeInvitationRecord(context.Context, uuid.UUID) error
}
