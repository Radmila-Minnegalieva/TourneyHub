package postgres

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

func (r *Repository) SetMatchParticipant(ctx context.Context, id uuid.UUID, slot model.MatchNextMatchSlot, participant *uuid.UUID) error {
	column := "home_application_id"
	if slot == "away" {
		column = "away_application_id"
	}
	_, err := r.db.Exec(ctx, `UPDATE matches SET `+column+`=$2,version=version+1,updated_at=now() WHERE id=$1`, id, participant)
	return dbError(err)
}
func (r *Repository) StageRemainingMatches(ctx context.Context, id uuid.UUID) (int, error) {
	var v int
	err := r.db.QueryRow(ctx, `SELECT count(*) FILTER(WHERE status NOT IN ('completed','bye')) FROM matches WHERE stage_id=$1`, id).Scan(&v)
	return v, dbError(err)
}
func (r *Repository) CompleteStage(ctx context.Context, id uuid.UUID) (model.StageKind, error) {
	var kind model.StageKind
	err := r.db.QueryRow(ctx, `UPDATE stages SET status='completed' WHERE id=$1 RETURNING kind`, id).Scan(&kind)
	return kind, dbError(err)
}
func (r *Repository) RemainingTournamentStages(ctx context.Context, id uuid.UUID) (int, error) {
	var v int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM stages WHERE tournament_id=$1 AND status<>'completed'`, id).Scan(&v)
	return v, dbError(err)
}
func (r *Repository) ReadPlayoffStage(ctx context.Context, id uuid.UUID) (model.Stage, error) {
	var s model.Stage
	err := r.db.QueryRow(ctx, `SELECT id,tournament_id,kind,position,status FROM stages WHERE tournament_id=$1 AND kind='playoffs' FOR UPDATE`, id).Scan(&s.Id, &s.TournamentId, &s.Kind, &s.Position, &s.Status)
	return s, dbError(err)
}
func (r *Repository) StageMatches(ctx context.Context, id uuid.UUID) ([]model.Match, error) {
	v, err := matchList(ctx, r.db, `SELECT `+matchColumns+` FROM matches m WHERE m.stage_id=$1 ORDER BY m.round,m.position`, id)
	return v, dbError(err)
}
func (r *Repository) DeleteStageMatches(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM matches WHERE stage_id=$1`, id)
	return dbError(err)
}
func (r *Repository) SetStageStatus(ctx context.Context, id uuid.UUID, status model.StageStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE stages SET status=$2 WHERE id=$1`, id, status)
	return dbError(err)
}
func (r *Repository) TournamentMatchCounts(ctx context.Context, id uuid.UUID) (int, int, error) {
	var total, unfinished int
	err := r.db.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE status NOT IN ('completed','bye')) FROM matches WHERE tournament_id=$1`, id).Scan(&total, &unfinished)
	return total, unfinished, dbError(err)
}
func (r *Repository) CompleteTournamentStages(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE stages SET status='completed' WHERE tournament_id=$1`, id)
	return dbError(err)
}
func (r *Repository) ClearFinalPlaces(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE applications SET final_place=NULL WHERE tournament_id=$1`, id)
	return dbError(err)
}
func (r *Repository) SetFinalPlace(ctx context.Context, id uuid.UUID, place int, onlyIfUnset bool) error {
	_, err := r.db.Exec(ctx, `UPDATE applications SET final_place=$2 WHERE id=$1 AND (NOT $3 OR final_place IS NULL)`, id, place, onlyIfUnset)
	return dbError(err)
}
func (r *Repository) UnplacedStandings(ctx context.Context, id uuid.UUID) ([]model.Standing, error) {
	rows, err := r.db.Query(ctx, `SELECT a.id,s.position FROM applications a JOIN standings s ON s.application_id=a.id WHERE a.tournament_id=$1 AND a.final_place IS NULL ORDER BY s.position,a.id`, id)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	result := []model.Standing{}
	for rows.Next() {
		var v model.Standing
		if err = rows.Scan(&v.ApplicationId, &v.Position); err != nil {
			return nil, dbError(err)
		}
		result = append(result, v)
	}
	return result, dbError(rows.Err())
}
func (r *Repository) PlacedApplicationsCount(ctx context.Context, id uuid.UUID) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM applications WHERE tournament_id=$1 AND final_place IS NOT NULL`, id).Scan(&n)
	return n, dbError(err)
}
func (r *Repository) PlayoffMatches(ctx context.Context, id uuid.UUID) ([]model.Match, error) {
	v, err := matchList(ctx, r.db, `SELECT `+matchColumns+` FROM matches m JOIN stages s ON s.id=m.stage_id WHERE m.tournament_id=$1 AND s.kind='playoffs' ORDER BY m.round DESC,m.position`, id)
	return v, dbError(err)
}
