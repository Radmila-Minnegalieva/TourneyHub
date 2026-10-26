package postgres

import (
	"context"
	"encoding/json"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

const applicationJSON = `jsonb_build_object(
 'id',a.id,'tournament_id',a.tournament_id,'status',a.status,'version',a.version,
 'seed',a.seed,'final_place',a.final_place,'rejection_reason',a.rejection_reason,'created_at',a.created_at,
 'team',jsonb_build_object('id',tm.id,'name',tm.name,'description',tm.description,
 'captain_id',tm.captain_id,'personal',tm.personal,'version',tm.version,
 'logo_url',CASE WHEN tm.logo_file_id IS NOT NULL THEN '/api/v1/files/'||tm.logo_file_id END),
 'players',COALESCE((SELECT jsonb_agg(jsonb_build_object('id',u.id,'nickname',u.nickname,
 'name',u.name,'avatar_url',CASE WHEN u.avatar_file_id IS NOT NULL THEN '/api/v1/files/'||u.avatar_file_id END)
 ORDER BY u.nickname,u.id) FROM application_players ap JOIN users u ON u.id=ap.user_id
 WHERE ap.application_id=a.id),'[]'::jsonb))`

func (r *Repository) ReadApplication(ctx context.Context, id uuid.UUID) (model.Application, error) {
	var application model.Application
	var raw []byte
	err := r.db.QueryRow(ctx, `SELECT `+applicationJSON+` FROM applications a JOIN teams tm ON tm.id=a.team_id WHERE a.id=$1`, id).Scan(&raw)
	if err != nil {
		return application, dbError(err)
	}
	err = json.Unmarshal(raw, &application)
	return application, err
}

func (r *Repository) ListApplicationPage(ctx context.Context, filter repository.ApplicationFilter) (model.ApplicationPage, error) {
	page := model.ApplicationPage{Items: []model.Application{}, Limit: filter.Limit, Offset: filter.Offset}
	from := ` FROM applications a JOIN teams tm ON tm.id=a.team_id WHERE
 ($1::uuid IS NULL OR a.tournament_id=$1) AND ($2::text IS NULL OR a.status=$2)
 AND ($3::uuid IS NULL OR EXISTS(SELECT 1 FROM team_members m WHERE m.team_id=a.team_id AND m.user_id=$3)
 OR EXISTS(SELECT 1 FROM application_players ap WHERE ap.application_id=a.id AND ap.user_id=$3))`
	args := []any{filter.TournamentID, filter.Status, filter.UserID}
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+from, args...).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+applicationJSON+from+` ORDER BY a.created_at DESC,a.id ASC LIMIT $4 OFFSET $5`, append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return page, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var item model.Application
		if err := rows.Scan(&raw); err != nil {
			return page, dbError(err)
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return page, err
		}
		page.Items = append(page.Items, item)
	}
	return page, dbError(rows.Err())
}

func (r *Repository) LockTournamentApplication(ctx context.Context, applicationID, tournamentID uuid.UUID) error {
	var found uuid.UUID
	return dbError(r.db.QueryRow(ctx, `SELECT id FROM applications WHERE id=$1 AND tournament_id=$2 FOR UPDATE`, applicationID, tournamentID).Scan(&found))
}

func (r *Repository) LockApplicationTeam(ctx context.Context, applicationID uuid.UUID) error {
	var team uuid.UUID
	return dbError(r.db.QueryRow(ctx, `SELECT tm.id FROM teams tm JOIN applications a ON a.team_id=tm.id WHERE a.id=$1 FOR SHARE OF tm`, applicationID).Scan(&team))
}

func (r *Repository) ApplicationTeamCaptain(ctx context.Context, teamID uuid.UUID) (uuid.UUID, error) {
	var captain uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT captain_id FROM teams WHERE id=$1 AND deleted_at IS NULL FOR SHARE`, teamID).Scan(&captain)
	return captain, dbError(err)
}

func (r *Repository) EligibleApplicationPlayersCount(ctx context.Context, teamID uuid.UUID, players []uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM team_members m JOIN users u ON u.id=m.user_id WHERE m.team_id=$1 AND m.user_id=ANY($2::uuid[]) AND u.blocked_at IS NULL AND u.email_verified_at IS NOT NULL`, teamID, players).Scan(&count)
	return count, dbError(err)
}

func (r *Repository) CreateApplicationRecord(ctx context.Context, id, tournament, team, actor uuid.UUID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO applications(id,tournament_id,team_id,submitted_by) VALUES($1,$2,$3,$4)`, id, tournament, team, actor)
	return dbError(err)
}

func (r *Repository) ReserveApplicationPlayers(ctx context.Context, applicationID, tournamentID uuid.UUID, players []uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO application_players(application_id,tournament_id,user_id) SELECT $1,$2,unnest($3::uuid[])`, applicationID, tournamentID, players); err != nil {
		return dbError(err)
	}
	_, err := r.db.Exec(ctx, `INSERT INTO tournament_player_claims(tournament_id,user_id,application_id) SELECT $1,unnest($2::uuid[]),$3`, tournamentID, players, applicationID)
	return dbError(err)
}

func (r *Repository) ReleaseApplicationPlayers(ctx context.Context, applicationID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tournament_player_claims WHERE application_id=$1`, applicationID)
	return dbError(err)
}

func (r *Repository) SetApplicationDecision(ctx context.Context, id uuid.UUID, status model.ApplicationStatus, reason *string) error {
	_, err := r.db.Exec(ctx, `UPDATE applications SET status=$2,rejection_reason=CASE WHEN $2='rejected' THEN $3 ELSE NULL END WHERE id=$1`, id, status, reason)
	return dbError(err)
}

func (r *Repository) BumpApplicationVersion(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE applications SET version=version+1,updated_at=now() WHERE id=$1`, id)
	return dbError(err)
}
