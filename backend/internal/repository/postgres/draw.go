package postgres

import (
	"context"
	"encoding/json"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

func (r *Repository) AcceptedApplicationsByRating(ctx context.Context, tournamentID, disciplineID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT a.id FROM applications a LEFT JOIN ratings r ON r.team_id=a.team_id AND r.discipline_id=$2 WHERE a.tournament_id=$1 AND a.status='accepted' ORDER BY coalesce(r.value,1500) DESC,a.created_at,a.id`, tournamentID, disciplineID)
	if err != nil {
		return nil, dbError(err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	return ids, dbError(err)
}
func (r *Repository) HasTournamentReports(ctx context.Context, id uuid.UUID) (bool, error) {
	var started bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM result_reports r JOIN matches m ON m.id=r.match_id WHERE m.tournament_id=$1)`, id).Scan(&started)
	return started, dbError(err)
}
func (r *Repository) InsertDrawPreview(ctx context.Context, p model.DrawPreview, actor uuid.UUID, method model.SeedingRequestMethod) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `INSERT INTO draw_previews(id,tournament_id,tournament_version,created_by,method,payload,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, p.Id, p.TournamentId, p.TournamentVersion, actor, method, raw, p.ExpiresAt)
	return dbError(err)
}
func (r *Repository) ReadDrawPreview(ctx context.Context, tournamentID, previewID uuid.UUID) (model.DrawPreview, bool, error) {
	var raw []byte
	var confirmed *time.Time
	var p model.DrawPreview
	err := r.db.QueryRow(ctx, `SELECT payload,confirmed_at FROM draw_previews WHERE id=$1 AND tournament_id=$2 FOR UPDATE`, previewID, tournamentID).Scan(&raw, &confirmed)
	if err != nil {
		return p, false, dbError(err)
	}
	err = json.Unmarshal(raw, &p)
	return p, confirmed != nil, err
}
func (r *Repository) ClearTournamentDraw(ctx context.Context, id uuid.UUID) error {
	for _, query := range []string{`DELETE FROM standings WHERE tournament_id=$1`, `DELETE FROM matches WHERE tournament_id=$1`, `DELETE FROM group_participants WHERE tournament_id=$1`, `DELETE FROM tournament_groups WHERE tournament_id=$1`, `DELETE FROM stages WHERE tournament_id=$1`} {
		if _, err := r.db.Exec(ctx, query, id); err != nil {
			return dbError(err)
		}
	}
	_, err := r.db.Exec(ctx, `UPDATE applications SET seed=NULL,final_place=NULL WHERE tournament_id=$1`, id)
	return dbError(err)
}
func (r *Repository) SetApplicationSeed(ctx context.Context, id uuid.UUID, position int) error {
	_, err := r.db.Exec(ctx, `UPDATE applications SET seed=$2 WHERE id=$1`, id, position)
	return dbError(err)
}
func (r *Repository) InsertStage(ctx context.Context, s model.Stage) error {
	_, err := r.db.Exec(ctx, `INSERT INTO stages(id,tournament_id,kind,position,status) VALUES($1,$2,$3,$4,$5)`, s.Id, s.TournamentId, s.Kind, s.Position, s.Status)
	return dbError(err)
}
func (r *Repository) InsertGroup(ctx context.Context, tournamentID uuid.UUID, g model.Group) error {
	if _, err := r.db.Exec(ctx, `INSERT INTO tournament_groups(id,stage_id,tournament_id,name) VALUES($1,$2,$3,$4)`, g.Id, g.StageId, tournamentID, g.Name); err != nil {
		return dbError(err)
	}
	for _, id := range g.ApplicationIds {
		if _, err := r.db.Exec(ctx, `INSERT INTO group_participants(group_id,stage_id,tournament_id,application_id) VALUES($1,$2,$3,$4)`, g.Id, g.StageId, tournamentID, id); err != nil {
			return dbError(err)
		}
	}
	return nil
}
func (r *Repository) InsertMatches(ctx context.Context, matches []model.Match) error {
	for _, m := range matches {
		_, err := r.db.Exec(ctx, `INSERT INTO matches(id,tournament_id,stage_id,group_id,round,position,home_application_id,away_application_id,winner_application_id,next_match_id,next_match_slot,loser_next_match_id,loser_next_match_slot,status) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, m.Id, m.TournamentId, m.StageId, m.GroupId, m.Round, m.Position, m.HomeApplicationId, m.AwayApplicationId, m.WinnerApplicationId, m.NextMatchId, m.NextMatchSlot, m.LoserNextMatchId, m.LoserNextMatchSlot, m.Status)
		if err != nil {
			return dbError(err)
		}
	}
	return nil
}
func (r *Repository) ConfirmDrawPreview(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE draw_previews SET confirmed_at=now() WHERE id=$1`, id)
	return dbError(err)
}
