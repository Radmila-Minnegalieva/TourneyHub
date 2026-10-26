package postgres

import (
	"context"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/pagination"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

func (r *Repository) AdminUserPage(ctx context.Context, p model.ListAdminUsersParams) (model.ProfilePage, error) {
	limit, offset := pagination.Values(p.Limit, p.Offset)
	page := model.ProfilePage{Items: []model.Profile{}, Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM users WHERE ($1::text IS NULL OR nickname ILIKE '%'||$1||'%' OR email ILIKE '%'||$1||'%')`, p.Search).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+profileColumns+` FROM users u WHERE ($1::text IS NULL OR nickname ILIKE '%'||$1||'%' OR email ILIKE '%'||$1||'%') ORDER BY created_at DESC,id LIMIT $2 OFFSET $3`, p.Search, limit, offset)
	if err != nil {
		return page, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanProfile(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, item)
	}
	return page, dbError(rows.Err())
}

func (r *Repository) UserHasAdminRole(ctx context.Context, id uuid.UUID) (bool, error) {
	var admin bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role='admin')`, id).Scan(&admin)
	return admin, dbError(err)
}

func (r *Repository) OtherActiveAdminCount(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM user_roles r JOIN users u ON u.id=r.user_id WHERE r.role='admin' AND u.blocked_at IS NULL AND u.email_verified_at IS NOT NULL AND u.id<>$1`, id).Scan(&count)
	return count, dbError(err)
}

func (r *Repository) LockUserVersion(ctx context.Context, id uuid.UUID) (int, error) {
	var version int
	err := r.db.QueryRow(ctx, `SELECT version FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&version)
	return version, dbError(err)
}

func (r *Repository) SetUserBlocked(ctx context.Context, id uuid.UUID, blocked bool, reason *string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET blocked_at=CASE WHEN $2 THEN now() END,block_reason=CASE WHEN $2 THEN $3 END,version=version+1,updated_at=now() WHERE id=$1`, id, blocked, reason)
	return dbError(err)
}

func (r *Repository) ReplaceUserRoles(ctx context.Context, id uuid.UUID, roles []model.Role) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, id); err != nil {
		return dbError(err)
	}
	for _, role := range roles {
		if _, err := r.db.Exec(ctx, `INSERT INTO user_roles(user_id,role) VALUES($1,$2)`, id, role); err != nil {
			return dbError(err)
		}
	}
	_, err := r.db.Exec(ctx, `UPDATE users SET version=version+1,updated_at=now() WHERE id=$1`, id)
	return dbError(err)
}

func (r *Repository) SetTournamentVisibility(ctx context.Context, id uuid.UUID, hidden bool) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET hidden=$2,version=version+1,updated_at=now() WHERE id=$1`, id, hidden)
	return dbError(err)
}

func (r *Repository) SoftDeleteTournamentRecord(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `UPDATE tournaments SET deleted_at=now(),hidden=true,version=version+1,updated_at=now() WHERE id=$1`, id); err != nil {
		return dbError(err)
	}
	_, err := r.db.Exec(ctx, `DELETE FROM tournament_player_claims WHERE tournament_id=$1`, id)
	return dbError(err)
}

func (r *Repository) AuditPage(ctx context.Context, p model.ListAuditParams) (model.AuditEntryPage, error) {
	limit, offset := pagination.Values(p.Limit, p.Offset)
	page := model.AuditEntryPage{Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM audit_entries WHERE ($1::uuid IS NULL OR entity_id=$1) AND ($2::uuid IS NULL OR actor_id=$2)`, p.EntityId, p.ActorId).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	items, err := jsonItems[model.AuditEntry](ctx, r.db, `SELECT to_jsonb(a) FROM audit_entries a WHERE ($1::uuid IS NULL OR entity_id=$1) AND ($2::uuid IS NULL OR actor_id=$2) ORDER BY created_at DESC,id LIMIT $3 OFFSET $4`, p.EntityId, p.ActorId, limit, offset)
	page.Items = items
	return page, err
}
