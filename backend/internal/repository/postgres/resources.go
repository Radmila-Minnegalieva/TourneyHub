package postgres

import (
	"context"
	"encoding/json"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
	"time"
)

func (r *Repository) CreateFile(ctx context.Context, id, userID uuid.UUID, upload model.FileUpload, contentType string) (model.File, error) {
	return jsonObject[model.File](ctx, r.db, `INSERT INTO files(id,owner_id,purpose,storage_key,filename,content_type,size_bytes,data) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING to_jsonb(files)-'data'`, id, userID, upload.Purpose, id.String(), upload.Filename, contentType, len(upload.Data), upload.Data)
}
func (r *Repository) File(ctx context.Context, id uuid.UUID) (repository.StoredFile, error) {
	var f repository.StoredFile
	err := r.db.QueryRow(ctx, `SELECT id,owner_id,purpose,filename,content_type,size_bytes,created_at,data FROM files WHERE id=$1`, id).Scan(&f.File.Id, &f.OwnerID, &f.File.Purpose, &f.File.Filename, &f.File.ContentType, &f.File.SizeBytes, &f.File.CreatedAt, &f.Data)
	return f, dbError(err)
}
func (r *Repository) CanReadProtocol(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	var allowed bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM matches m JOIN tournaments t ON t.id=m.tournament_id
 WHERE (EXISTS(SELECT 1 FROM report_files rf JOIN result_reports r ON r.id=rf.report_id WHERE rf.file_id=$1 AND r.match_id=m.id)
 OR EXISTS(SELECT 1 FROM dispute_files df JOIN disputes d ON d.id=df.dispute_id WHERE df.file_id=$1 AND d.match_id=m.id))
 AND (t.organizer_id=$2 OR t.referee_id=$2 OR m.referee_id=$2
 OR EXISTS(SELECT 1 FROM application_players ap WHERE ap.application_id IN (m.home_application_id,m.away_application_id) AND ap.user_id=$2)
 OR EXISTS(SELECT 1 FROM applications a JOIN teams tm ON tm.id=a.team_id WHERE a.id IN (m.home_application_id,m.away_application_id) AND tm.captain_id=$2)))`, id, userID).Scan(&allowed)
	return allowed, dbError(err)
}
func (r *Repository) ExportRows(ctx context.Context, id uuid.UUID) ([]repository.ExportRow, error) {
	rows, err := r.db.Query(ctx, `SELECT row_type,place,team,match_id,home_team,away_team,home_score,away_score,status FROM (
 SELECT 'standing'::text row_type,coalesce(a.final_place::text,'') place,t.name team,''::text match_id,''::text home_team,''::text away_team,''::text home_score,''::text away_score,''::text status,0 kind,coalesce(a.final_place,999) stage,0 round,0 position,a.id
 FROM applications a JOIN teams t ON t.id=a.team_id WHERE a.tournament_id=$1 AND a.status='accepted'
 UNION ALL
 SELECT 'match','', '',m.id::text,coalesce(h.name,''),coalesce(w.name,''),coalesce(m.home_score::text,''),coalesce(m.away_score::text,''),m.status,1,s.position,m.round,m.position,m.id
 FROM matches m JOIN stages s ON s.id=m.stage_id LEFT JOIN applications ha ON ha.id=m.home_application_id LEFT JOIN teams h ON h.id=ha.team_id LEFT JOIN applications wa ON wa.id=m.away_application_id LEFT JOIN teams w ON w.id=wa.team_id
 WHERE m.tournament_id=$1 AND m.status IN ('completed','bye')
 ) results ORDER BY kind,stage,round,position,id`, id)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	result := make([]repository.ExportRow, 0)
	for rows.Next() {
		var row repository.ExportRow
		if err := rows.Scan(&row.Kind, &row.Place, &row.Team, &row.MatchID, &row.HomeTeam, &row.AwayTeam, &row.HomeScore, &row.AwayScore, &row.Status); err != nil {
			return nil, dbError(err)
		}
		result = append(result, row)
	}
	return result, dbError(rows.Err())
}
func (r *Repository) EventBounds(ctx context.Context, id uuid.UUID) (int64, int64, error) {
	var min, max int64
	err := r.db.QueryRow(ctx, `SELECT coalesce(min(id),0),coalesce(max(id),0) FROM tournament_events WHERE tournament_id=$1`, id).Scan(&min, &max)
	return min, max, dbError(err)
}
func (r *Repository) EventsAfter(ctx context.Context, id uuid.UUID, after int64, limit int) ([]model.TournamentEvent, error) {
	rows, err := r.db.Query(ctx, `SELECT id,kind,payload FROM tournament_events WHERE tournament_id=$1 AND id>$2 ORDER BY id LIMIT $3`, id, after, limit)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	result := make([]model.TournamentEvent, 0)
	for rows.Next() {
		var event model.TournamentEvent
		if err := rows.Scan(&event.ID, &event.Kind, &event.Payload); err != nil {
			return nil, dbError(err)
		}
		result = append(result, event)
	}
	return result, dbError(rows.Err())
}
func (r *Repository) PendingMail(ctx context.Context) (repository.PendingMail, error) {
	var result repository.PendingMail
	var raw []byte
	err := r.db.QueryRow(ctx, `SELECT id,payload FROM outbox_events WHERE topic='email' AND delivered_at IS NULL AND available_at<=now() ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&result.ID, &raw)
	if err != nil {
		return result, dbError(err)
	}
	err = json.Unmarshal(raw, &result.Message)
	return result, err
}
func (r *Repository) CompleteMail(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE outbox_events SET delivered_at=now(),payload='{"redacted":true}'::jsonb,attempts=attempts+1 WHERE id=$1`, id)
	return dbError(err)
}
func (r *Repository) RetryMail(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE outbox_events SET attempts=attempts+1,available_at=now()+make_interval(secs=>least(3600,30*power(2,least(attempts,7)))::int) WHERE id=$1`, id)
	return dbError(err)
}
func (r *Repository) CleanupRateLimits(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM auth_rate_limits WHERE window_start<now()-interval '1 day'`)
	return dbError(err)
}
func (r *Repository) CleanupEvents(ctx context.Context, before time.Time) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tournament_events WHERE created_at<$1`, before)
	return dbError(err)
}
func (r *Repository) BootstrapState(ctx context.Context, email string) (repository.BootstrapState, error) {
	var result repository.BootstrapState
	err := r.db.QueryRow(ctx, `SELECT count(*),coalesce(bool_or(lower(u.email)=$1 AND u.blocked_at IS NULL AND u.email_verified_at IS NOT NULL),false) FROM user_roles r JOIN users u ON u.id=r.user_id WHERE r.role='admin'`, email).Scan(&result.AdminCount, &result.SameActiveAdmin)
	return result, dbError(err)
}
func (r *Repository) CreateBootstrapAdmin(ctx context.Context, id uuid.UUID, email, nickname, hash string) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO users(id,email,nickname,password_hash,email_verified_at,personal_data_consent_at) VALUES($1,$2,$3,$4,now(),now())`, id, email, nickname, hash); err != nil {
		return dbError(err)
	}
	for _, role := range []string{"participant", "organizer", "referee", "admin"} {
		if _, err := r.db.Exec(ctx, `INSERT INTO user_roles(user_id,role) VALUES($1,$2)`, id, role); err != nil {
			return dbError(err)
		}
	}
	return nil
}
