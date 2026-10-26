package postgres

import (
	"context"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

const profileColumns = `u.id,u.email,u.nickname,u.name,u.timezone,u.email_notifications,
 u.email_verified_at IS NOT NULL,u.blocked_at IS NOT NULL,u.version,
 CASE WHEN u.avatar_file_id IS NULL THEN NULL ELSE '/api/v1/files/'||u.avatar_file_id END,
 ARRAY(SELECT role FROM user_roles WHERE user_id=u.id ORDER BY role)`

type identityScanner interface {
	Scan(...any) error
}

func scanProfile(row identityScanner) (model.Profile, error) {
	var p model.Profile
	var roles []string
	err := row.Scan(&p.Id, &p.Email, &p.Nickname, &p.Name, &p.Timezone,
		&p.EmailNotifications, &p.EmailVerified, &p.Blocked, &p.Version, &p.AvatarUrl, &roles)
	p.Roles = make([]model.Role, len(roles))
	for i, role := range roles {
		p.Roles[i] = model.Role(role)
	}
	return p, dbError(err)
}

func (r *Repository) Profile(ctx context.Context, id uuid.UUID) (model.Profile, error) {
	return scanProfile(r.db.QueryRow(ctx, `SELECT `+profileColumns+` FROM users u WHERE u.id=$1`, id))
}

func (r *Repository) CreateUser(ctx context.Context, req model.RegisterRequest, hash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `INSERT INTO users(email,password_hash,nickname,name,personal_data_consent_at) VALUES(lower($1),$2,$3,$4,now()) RETURNING id`, string(req.Email), hash, req.Nickname, req.Name).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) AddUserRole(ctx context.Context, id uuid.UUID, role model.Role) error {
	_, err := r.db.Exec(ctx, `INSERT INTO user_roles(user_id,role) VALUES($1,$2)`, id, string(role))
	return dbError(err)
}

func (r *Repository) InvalidateAccountTokens(ctx context.Context, id uuid.UUID, purpose string) error {
	_, err := r.db.Exec(ctx, `UPDATE user_tokens SET consumed_at=now() WHERE user_id=$1 AND purpose=$2 AND consumed_at IS NULL`, id, purpose)
	return dbError(err)
}

func (r *Repository) CreateAccountToken(ctx context.Context, id uuid.UUID, purpose, hash string, expires time.Time) error {
	_, err := r.db.Exec(ctx, `INSERT INTO user_tokens(user_id,purpose,token_hash,expires_at) VALUES($1,$2,$3,$4)`, id, purpose, hash, expires)
	return dbError(err)
}

func (r *Repository) ConsumeAccountToken(ctx context.Context, hash, purpose string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `UPDATE user_tokens SET consumed_at=now() WHERE token_hash=$1 AND purpose=$2 AND consumed_at IS NULL AND expires_at>now() RETURNING user_id`, hash, purpose).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) VerifyUserEmail(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET email_verified_at=COALESCE(email_verified_at,now()),updated_at=now(),version=version+1 WHERE id=$1`, id)
	return dbError(err)
}

func (r *Repository) RecordLoginAttempt(ctx context.Context, ip string) (int, error) {
	var attempts int
	err := r.db.QueryRow(ctx, `INSERT INTO auth_rate_limits(ip,window_start,attempts) VALUES($1,now(),1)
 ON CONFLICT(ip) DO UPDATE SET
 attempts=CASE WHEN auth_rate_limits.window_start<=now()-interval '1 minute' THEN 1 ELSE auth_rate_limits.attempts+1 END,
 window_start=CASE WHEN auth_rate_limits.window_start<=now()-interval '1 minute' THEN now() ELSE auth_rate_limits.window_start END
 RETURNING attempts`, ip).Scan(&attempts)
	return attempts, dbError(err)
}

func (r *Repository) CreateRefreshSession(ctx context.Context, userID uuid.UUID, hash string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `INSERT INTO refresh_sessions(user_id,token_hash,expires_at) VALUES($1,$2,now()+interval '30 days') RETURNING id`, userID, hash).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) CreateAccessSession(ctx context.Context, userID, refreshID uuid.UUID, hash string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO access_sessions(user_id,refresh_session_id,token_hash,expires_at) VALUES($1,$2,$3,now()+interval '15 minutes')`, userID, refreshID, hash)
	return dbError(err)
}

func (r *Repository) LoginAccount(ctx context.Context, email string) (repository.LoginAccount, error) {
	var account repository.LoginAccount
	err := r.db.QueryRow(ctx, `SELECT id,password_hash,email_verified_at IS NOT NULL,blocked_at IS NOT NULL FROM users WHERE lower(email)=lower($1)`, email).Scan(&account.ID, &account.PasswordHash, &account.EmailVerified, &account.Blocked)
	return account, dbError(err)
}

func (r *Repository) LockLoginAccount(ctx context.Context, id uuid.UUID, hash string) (bool, error) {
	var usable bool
	err := r.db.QueryRow(ctx, `SELECT blocked_at IS NULL AND email_verified_at IS NOT NULL AND password_hash=$2 FROM users WHERE id=$1 FOR UPDATE`, id, hash).Scan(&usable)
	return usable, dbError(err)
}

func (r *Repository) AuthenticateAccess(ctx context.Context, hash string) (repository.AuthenticatedUser, error) {
	var result repository.AuthenticatedUser
	var roles []string
	err := r.db.QueryRow(ctx, `SELECT u.id,ARRAY(SELECT role FROM user_roles WHERE user_id=u.id ORDER BY role)
 FROM access_sessions s JOIN users u ON u.id=s.user_id JOIN refresh_sessions r ON r.id=s.refresh_session_id
 WHERE s.token_hash=$1 AND s.expires_at>now() AND s.revoked_at IS NULL AND r.revoked_at IS NULL
 AND r.expires_at>now() AND u.blocked_at IS NULL AND u.email_verified_at IS NOT NULL`, hash).Scan(&result.ID, &roles)
	result.Roles = make([]model.Role, len(roles))
	for i, role := range roles {
		result.Roles[i] = model.Role(role)
	}
	return result, dbError(err)
}

func (r *Repository) LockRefreshSession(ctx context.Context, hash string) (repository.SessionIdentity, error) {
	var result repository.SessionIdentity
	err := r.db.QueryRow(ctx, `SELECT r.user_id,r.id FROM refresh_sessions r JOIN users u ON u.id=r.user_id
 WHERE r.token_hash=$1 AND r.revoked_at IS NULL AND r.expires_at>now() AND u.blocked_at IS NULL AND u.email_verified_at IS NOT NULL FOR UPDATE OF r,u`, hash).Scan(&result.UserID, &result.SessionID)
	return result, dbError(err)
}

func (r *Repository) RevokeRefreshSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE refresh_sessions SET revoked_at=now() WHERE id=$1`, id)
	return dbError(err)
}

func (r *Repository) RevokeSessionAccess(ctx context.Context, refreshID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE access_sessions SET revoked_at=now() WHERE refresh_session_id=$1 AND revoked_at IS NULL`, refreshID)
	return dbError(err)
}

func (r *Repository) LogoutSession(ctx context.Context, hash string) error {
	_, err := r.db.Exec(ctx, `WITH revoked AS (UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,now()) WHERE token_hash=$1 RETURNING id)
 UPDATE access_sessions SET revoked_at=COALESCE(revoked_at,now()) WHERE refresh_session_id IN (SELECT id FROM revoked)`, hash)
	return dbError(err)
}

func (r *Repository) PasswordResetAccount(ctx context.Context, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE lower(email)=lower($1) AND blocked_at IS NULL FOR UPDATE`, email).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) SetPasswordHash(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET password_hash=$2,version=version+1,updated_at=now() WHERE id=$1`, id, hash)
	return dbError(err)
}

func (r *Repository) RevokeUserSessions(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `UPDATE refresh_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, id); err != nil {
		return dbError(err)
	}
	_, err := r.db.Exec(ctx, `UPDATE access_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, id)
	return dbError(err)
}

func (r *Repository) OwnsFile(ctx context.Context, fileID, ownerID uuid.UUID, purpose string) (bool, error) {
	var allowed bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM files WHERE id=$1 AND owner_id=$2 AND purpose=$3)`, fileID, ownerID, purpose).Scan(&allowed)
	return allowed, dbError(err)
}

func (r *Repository) UpdateUserProfile(ctx context.Context, id uuid.UUID, req model.UpdateProfileRequest) (bool, error) {
	tag, err := r.db.Exec(ctx, `UPDATE users SET nickname=COALESCE($3,nickname),name=COALESCE($4,name),avatar_file_id=COALESCE($5,avatar_file_id),timezone=COALESCE($6,timezone),email_notifications=COALESCE($7,email_notifications),version=version+1,updated_at=now() WHERE id=$1 AND version=$2`, id, req.Version, req.Nickname, req.Name, req.AvatarFileId, req.Timezone, req.EmailNotifications)
	return tag.RowsAffected() > 0, dbError(err)
}

func (r *Repository) SearchPublicUsers(ctx context.Context, nickname string, limit, offset int) (model.PublicUserPage, error) {
	page := model.PublicUserPage{Items: []model.PublicUser{}, Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM users WHERE blocked_at IS NULL AND strpos(lower(nickname),$1)>0`, nickname).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT id,nickname,name,CASE WHEN avatar_file_id IS NULL THEN NULL ELSE '/api/v1/files/'||avatar_file_id END FROM users WHERE blocked_at IS NULL AND strpos(lower(nickname),$1)>0 ORDER BY created_at DESC,id LIMIT $2 OFFSET $3`, nickname, limit, offset)
	if err != nil {
		return page, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var item model.PublicUser
		if err := rows.Scan(&item.Id, &item.Nickname, &item.Name, &item.AvatarUrl); err != nil {
			return page, dbError(err)
		}
		page.Items = append(page.Items, item)
	}
	return page, dbError(rows.Err())
}
