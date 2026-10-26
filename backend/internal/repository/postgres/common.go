package postgres

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *Repository) Audit(ctx context.Context, actor uuid.UUID, action, entityType string, entityID uuid.UUID, before, after any) error {
	var actorID *uuid.UUID
	if actor != uuid.Nil {
		actorID = &actor
	}
	_, err := r.db.Exec(ctx, `INSERT INTO audit_entries(actor_id,action,entity_type,entity_id,before,after) VALUES($1,$2,$3,$4,$5,$6)`, actorID, action, entityType, entityID, before, after)
	return dbError(err)
}
func (r *Repository) Outbox(ctx context.Context, topic string, aggregateID uuid.UUID, payload any) error {
	_, err := r.db.Exec(ctx, `INSERT INTO outbox_events(topic,aggregate_id,payload) VALUES($1,$2,$3)`, topic, aggregateID, payload)
	return dbError(err)
}
func (r *Repository) Notify(ctx context.Context, userID uuid.UUID, kind, message string, resourceID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO notifications(user_id,kind,message,resource_id) VALUES($1,$2,$3,$4)`, userID, kind, message, resourceID)
	return dbError(err)
}
func (r *Repository) NotificationRecipient(ctx context.Context, userID uuid.UUID) (repository.NotificationRecipient, error) {
	var result repository.NotificationRecipient
	err := r.db.QueryRow(ctx, `SELECT email,email_notifications AND blocked_at IS NULL FROM users WHERE id=$1`, userID).Scan(&result.Email, &result.Enabled)
	return result, dbError(err)
}
func (r *Repository) Event(ctx context.Context, id uuid.UUID, kind string, payload any) error {
	_, err := r.db.Exec(ctx, `INSERT INTO tournament_events(tournament_id,kind,payload) VALUES($1,$2,$3)`, id, kind, payload)
	return dbError(err)
}
func (r *Repository) LockSchedule(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `SELECT pg_advisory_xact_lock(170221)`)
	return dbError(err)
}
func (r *Repository) LockRatings(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1::text,913))`, id.String())
	return dbError(err)
}
func (r *Repository) LockAdmin(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `SELECT pg_advisory_xact_lock(76123841)`)
	return dbError(err)
}
