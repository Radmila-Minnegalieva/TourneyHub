package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/authctx"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/security"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) queueAccountToken(ctx context.Context, q repository.Store, id uuid.UUID, email, purpose string) error {
	token, err := security.NewToken()
	if err != nil {
		return err
	}
	if err := q.InvalidateAccountTokens(ctx, id, purpose); err != nil {
		return err
	}
	ttl := time.Hour
	subject := "TourneyHub: восстановление пароля"
	if purpose == "verify_email" {
		ttl = 24 * time.Hour
		subject = "TourneyHub: подтверждение e-mail"
	}
	if err := q.CreateAccountToken(ctx, id, purpose, security.TokenHash(token), time.Now().UTC().Add(ttl)); err != nil {
		return err
	}
	return q.Outbox(ctx, "email", id, map[string]any{"to": email, "subject": subject, "text": "Откройте одноразовую ссылку: " + accountTokenURL(s.cfg.Public.WebURL, purpose, token), "purpose": purpose, "token": token})
}

func accountTokenURL(webURL, purpose, token string) string {
	base := strings.TrimRight(webURL, "/")
	path := "/auth/reset-password"
	if purpose == "verify_email" {
		path = "/auth/verify-email"
	}
	return base + path + "?token=" + url.QueryEscape(token)
}

func (s *Service) Register(ctx context.Context, req model.RegisterRequest) error {
	if !bool(req.PersonalDataConsent) {
		return invalid("personal data consent is required")
	}
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return err
	}
	return s.withTx(ctx, func(q repository.Store) error {
		id, err := q.CreateUser(ctx, req, hash)
		if err != nil {
			return err
		}
		if err := q.AddUserRole(ctx, id, model.Participant); err != nil {
			return err
		}
		return s.queueAccountToken(ctx, q, id, string(req.Email), "verify_email")
	})
}

func (s *Service) VerifyEmail(ctx context.Context, req model.TokenRequest) error {
	return s.withTx(ctx, func(q repository.Store) error {
		id, err := q.ConsumeAccountToken(ctx, security.TokenHash(req.Token), "verify_email")
		if errors.Is(err, repository.ErrNotFound) {
			return invalid("verification token is invalid or expired")
		}
		if err != nil {
			return err
		}
		return q.VerifyUserEmail(ctx, id)
	})
}

func (s *Service) allowLogin(ctx context.Context, ip string) error {
	attempts, err := s.repo.RecordLoginAttempt(ctx, ip)
	if err != nil {
		return err
	}
	if attempts > 5 {
		return rateLimited("too many login attempts")
	}
	return nil
}

func issueSession(ctx context.Context, q repository.Store, userID uuid.UUID) (model.Tokens, error) {
	var tokens model.Tokens
	access, err := security.NewToken()
	if err != nil {
		return tokens, err
	}
	refresh, err := security.NewToken()
	if err != nil {
		return tokens, err
	}
	sessionID, err := q.CreateRefreshSession(ctx, userID, security.TokenHash(refresh))
	if err != nil {
		return tokens, err
	}
	if err := q.CreateAccessSession(ctx, userID, sessionID, security.TokenHash(access)); err != nil {
		return tokens, err
	}
	return model.Tokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: 900, TokenType: "Bearer"}, nil
}

func (s *Service) Login(ctx context.Context, req model.LoginRequest, ip string) (model.Tokens, error) {
	var tokens model.Tokens
	if err := s.allowLogin(ctx, ip); err != nil {
		return tokens, err
	}
	if len(req.Password) > 72 {
		return tokens, unauthorized("invalid credentials")
	}
	account, err := s.repo.LoginAccount(ctx, string(req.Email))
	if errors.Is(err, repository.ErrNotFound) {
		// Равная работа bcrypt не раскрывает наличие учётной записи по времени ответа.
		_ = security.CheckPassword("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", req.Password)
		return tokens, unauthorized("invalid credentials")
	}
	if err != nil {
		return tokens, err
	}
	if !security.CheckPassword(account.PasswordHash, req.Password) || account.Blocked || !account.EmailVerified {
		return tokens, unauthorized("invalid credentials or account unavailable")
	}
	err = s.withTx(ctx, func(q repository.Store) error {
		// Повторная проверка под блокировкой сериализует вход с блокировкой/сменой пароля.
		usable, err := q.LockLoginAccount(ctx, account.ID, account.PasswordHash)
		if err != nil {
			return err
		}
		if !usable {
			return unauthorized("account unavailable")
		}
		tokens, err = issueSession(ctx, q, account.ID)
		return err
	})
	return tokens, err
}

func (s *Service) Authenticate(ctx context.Context, token string) (authctx.Principal, error) {
	var principal authctx.Principal
	if token == "" || len(token) > 512 {
		return principal, unauthorized("invalid access token")
	}
	account, err := s.repo.AuthenticateAccess(ctx, security.TokenHash(token))
	if errors.Is(err, repository.ErrNotFound) {
		return principal, unauthorized("invalid access token")
	}
	if err != nil {
		return principal, err
	}
	return authctx.Principal{ID: account.ID, Roles: account.Roles}, nil
}

func (s *Service) RefreshTokens(ctx context.Context, req model.TokenRequest) (model.Tokens, error) {
	var tokens model.Tokens
	err := s.withTx(ctx, func(q repository.Store) error {
		identity, err := q.LockRefreshSession(ctx, security.TokenHash(req.Token))
		if errors.Is(err, repository.ErrNotFound) {
			return unauthorized("invalid refresh token")
		}
		if err != nil {
			return err
		}
		if err := q.RevokeRefreshSession(ctx, identity.SessionID); err != nil {
			return err
		}
		if err := q.RevokeSessionAccess(ctx, identity.SessionID); err != nil {
			return err
		}
		tokens, err = issueSession(ctx, q, identity.UserID)
		return err
	})
	return tokens, err
}

func (s *Service) Logout(ctx context.Context, req model.TokenRequest) error {
	return s.withTx(ctx, func(q repository.Store) error { return q.LogoutSession(ctx, security.TokenHash(req.Token)) })
}

func (s *Service) RequestPasswordReset(ctx context.Context, req model.EmailRequest) error {
	return s.withTx(ctx, func(q repository.Store) error {
		id, err := q.PasswordResetAccount(ctx, string(req.Email))
		if errors.Is(err, repository.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return s.queueAccountToken(ctx, q, id, string(req.Email), "reset_password")
	})
}

func (s *Service) ResetPassword(ctx context.Context, req model.ResetPasswordRequest) error {
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return err
	}
	return s.withTx(ctx, func(q repository.Store) error {
		id, err := q.ConsumeAccountToken(ctx, security.TokenHash(req.Token), "reset_password")
		if errors.Is(err, repository.ErrNotFound) {
			return invalid("password reset token is invalid or expired")
		}
		if err != nil {
			return err
		}
		if err := q.SetPasswordHash(ctx, id, hash); err != nil {
			return err
		}
		if err := q.RevokeUserSessions(ctx, id); err != nil {
			return err
		}
		return q.InvalidateAccountTokens(ctx, id, "reset_password")
	})
}

func (s *Service) GetProfile(ctx context.Context) (model.Profile, error) {
	return s.repo.Profile(ctx, currentUser(ctx))
}

func validateProfileTimezone(timezone *string) error {
	if timezone == nil {
		return nil
	}
	if _, err := time.LoadLocation(*timezone); err != nil {
		return invalid("unknown IANA timezone")
	}
	return nil
}

func validateProfileAvatar(ctx context.Context, q repository.Store, fileID *uuid.UUID, ownerID uuid.UUID) error {
	if fileID == nil {
		return nil
	}
	allowed, err := q.OwnsFile(ctx, *fileID, ownerID, "avatar")
	if err != nil {
		return err
	}
	if !allowed {
		return invalid("avatar must be an owned avatar file")
	}
	return nil
}

func (s *Service) UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (model.Profile, error) {
	var profile model.Profile
	if err := validateProfileTimezone(req.Timezone); err != nil {
		return profile, err
	}
	err := s.withTx(ctx, func(q repository.Store) error {
		id := currentUser(ctx)
		if err := validateProfileAvatar(ctx, q, req.AvatarFileId, id); err != nil {
			return err
		}
		updated, err := q.UpdateUserProfile(ctx, id, req)
		if err != nil {
			return err
		}
		if !updated {
			return conflict("profile version changed")
		}
		profile, err = q.Profile(ctx, id)
		return err
	})
	return profile, err
}

func (s *Service) SearchUsers(ctx context.Context, params model.SearchUsersParams) (model.PublicUserPage, error) {
	limit, offset := pagination(params.Limit, params.Offset)
	return s.repo.SearchPublicUsers(ctx, strings.ToLower(params.Nickname), limit, offset)
}
