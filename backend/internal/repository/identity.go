package repository

import (
	"context"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

// LoginAccount содержит только данные, необходимые для проверки входа.
type LoginAccount struct {
	ID            uuid.UUID
	PasswordHash  string
	EmailVerified bool
	Blocked       bool
}

type SessionIdentity struct {
	UserID    uuid.UUID
	SessionID uuid.UUID
}

type AuthenticatedUser struct {
	ID    uuid.UUID
	Roles []model.Role
}

// Identity описывает хранение аккаунтов и сессий без зависимости от драйвера БД.
type Identity interface {
	Profile(context.Context, uuid.UUID) (model.Profile, error)
	CreateUser(context.Context, model.RegisterRequest, string) (uuid.UUID, error)
	AddUserRole(context.Context, uuid.UUID, model.Role) error
	InvalidateAccountTokens(context.Context, uuid.UUID, string) error
	CreateAccountToken(context.Context, uuid.UUID, string, string, time.Time) error
	ConsumeAccountToken(context.Context, string, string) (uuid.UUID, error)
	VerifyUserEmail(context.Context, uuid.UUID) error
	RecordLoginAttempt(context.Context, string) (int, error)
	CreateRefreshSession(context.Context, uuid.UUID, string) (uuid.UUID, error)
	CreateAccessSession(context.Context, uuid.UUID, uuid.UUID, string) error
	LoginAccount(context.Context, string) (LoginAccount, error)
	LockLoginAccount(context.Context, uuid.UUID, string) (bool, error)
	AuthenticateAccess(context.Context, string) (AuthenticatedUser, error)
	LockRefreshSession(context.Context, string) (SessionIdentity, error)
	RevokeRefreshSession(context.Context, uuid.UUID) error
	RevokeSessionAccess(context.Context, uuid.UUID) error
	LogoutSession(context.Context, string) error
	PasswordResetAccount(context.Context, string) (uuid.UUID, error)
	SetPasswordHash(context.Context, uuid.UUID, string) error
	RevokeUserSessions(context.Context, uuid.UUID) error
	OwnsFile(context.Context, uuid.UUID, uuid.UUID, string) (bool, error)
	UpdateUserProfile(context.Context, uuid.UUID, model.UpdateProfileRequest) (bool, error)
	SearchPublicUsers(context.Context, string, int, int) (model.PublicUserPage, error)
}
