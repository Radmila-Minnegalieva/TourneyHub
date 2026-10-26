package postgres

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
	"time"
)

func (r *Repository) RoundMatches(ctx context.Context, tournamentID, stageID uuid.UUID, round int) ([]model.Match, error) {
	matches, err := matchList(ctx, r.db, `SELECT `+matchColumns+` FROM matches m WHERE m.tournament_id=$1 AND m.stage_id=$2 AND m.round=$3 AND m.status<>'bye' ORDER BY m.position FOR UPDATE`, tournamentID, stageID, round)
	return matches, dbError(err)
}
func (r *Repository) ClearRoundSchedule(ctx context.Context, tournamentID, stageID uuid.UUID, round int) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET starts_at=NULL WHERE tournament_id=$1 AND stage_id=$2 AND round=$3 AND status='scheduled'`, tournamentID, stageID, round)
	return dbError(err)
}
func (r *Repository) SetMatchSchedule(ctx context.Context, id uuid.UUID, start *time.Time, venue, url *string) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET starts_at=$2,venue=$3,online_url=$4,version=version+1,updated_at=now() WHERE id=$1`, id, start, venue, url)
	return dbError(err)
}
func (r *Repository) HasParticipantScheduleConflict(ctx context.Context, m model.Match, start time.Time, referee *uuid.UUID) (bool, error) {
	var overlap bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM matches other JOIN tournaments ot ON ot.id=other.tournament_id WHERE other.id<>$1 AND ot.status='in_progress' AND other.status NOT IN ('completed','bye') AND other.starts_at > $2::timestamptz-interval '1 hour' AND other.starts_at < $2::timestamptz+interval '1 hour' AND (($3::uuid IS NOT NULL AND coalesce(other.referee_id,ot.referee_id)=$3) OR EXISTS(SELECT 1 FROM application_players mine JOIN application_players theirs ON theirs.user_id=mine.user_id WHERE mine.application_id IN ($4,$5) AND theirs.application_id IN (other.home_application_id,other.away_application_id))))`, m.Id, start, referee, m.HomeApplicationId, m.AwayApplicationId).Scan(&overlap)
	return overlap, dbError(err)
}
func (r *Repository) HasScheduleDependencyConflict(ctx context.Context, m model.Match, start time.Time) (bool, error) {
	var overlap bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM matches related WHERE related.starts_at IS NOT NULL AND (((related.next_match_id=$1 OR related.loser_next_match_id=$1) AND related.starts_at+interval '1 hour'>$2) OR (related.id IN ($3,$4) AND related.starts_at<$2::timestamptz+interval '1 hour')))`, m.Id, start, m.NextMatchId, m.LoserNextMatchId).Scan(&overlap)
	return overlap, dbError(err)
}
func (r *Repository) ValidReferee(ctx context.Context, id uuid.UUID) (bool, error) {
	var valid bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users u JOIN user_roles r ON r.user_id=u.id WHERE u.id=$1 AND u.blocked_at IS NULL AND u.email_verified_at IS NOT NULL AND r.role='referee')`, id).Scan(&valid)
	return valid, dbError(err)
}
func (r *Repository) SetMatchReferee(ctx context.Context, id, refereeID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET referee_id=$2,version=version+1,updated_at=now() WHERE id=$1`, id, refereeID)
	return dbError(err)
}
