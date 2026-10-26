package postgres

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

func readStandings(ctx context.Context, q DBTX, tid uuid.UUID) ([]model.Standings, error) {
	rows, err := q.Query(ctx, `SELECT s.stage_id,s.group_id,s.application_id,s.position,s.played,s.wins,s.draws,s.losses,s.score_for,s.score_against,s.points FROM standings s JOIN stages st ON st.id=s.stage_id LEFT JOIN tournament_groups g ON g.id=s.group_id WHERE s.tournament_id=$1 ORDER BY st.position,g.name NULLS FIRST,s.position,s.application_id`, tid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Standings{}
	for rows.Next() {
		var sid uuid.UUID
		var gid *uuid.UUID
		var r model.Standing
		if err = rows.Scan(&sid, &gid, &r.ApplicationId, &r.Position, &r.Played, &r.Wins, &r.Draws, &r.Losses, &r.ScoreFor, &r.ScoreAgainst, &r.Points); err != nil {
			return nil, err
		}
		r.ScoreDifference = r.ScoreFor - r.ScoreAgainst
		if len(result) == 0 || result[len(result)-1].StageId != sid || !optionalUUIDEqual(result[len(result)-1].GroupId, gid) {
			result = append(result, model.Standings{StageId: sid, GroupId: gid, Rows: []model.Standing{}})
		}
		result[len(result)-1].Rows = append(result[len(result)-1].Rows, r)
	}
	return result, rows.Err()
}
func (r *Repository) ReadStandings(ctx context.Context, id uuid.UUID) ([]model.Standings, error) {
	v, err := readStandings(ctx, r.db, id)
	return v, dbError(err)
}
func (r *Repository) ReadStandingParticipants(ctx context.Context, tournamentID, stageID, disciplineID uuid.UUID) ([]repository.StandingParticipant, error) {
	rows, err := r.db.Query(ctx, `SELECT a.id,gp.group_id,coalesce(r.value,1500) FROM stages s JOIN applications a ON a.tournament_id=s.tournament_id AND a.status='accepted' LEFT JOIN group_participants gp ON gp.stage_id=s.id AND gp.application_id=a.id LEFT JOIN ratings r ON r.team_id=a.team_id AND r.discipline_id=$3 WHERE s.id=$1 AND s.tournament_id=$2 AND (s.kind='round_robin' OR gp.application_id IS NOT NULL) ORDER BY a.seed,a.id`, stageID, tournamentID, disciplineID)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	result := []repository.StandingParticipant{}
	for rows.Next() {
		var v repository.StandingParticipant
		if err = rows.Scan(&v.ApplicationID, &v.GroupID, &v.Elo); err != nil {
			return nil, dbError(err)
		}
		result = append(result, v)
	}
	return result, dbError(rows.Err())
}
func (r *Repository) CompletedStageMatches(ctx context.Context, id uuid.UUID) ([]model.Match, error) {
	v, err := matchList(ctx, r.db, `SELECT `+matchColumns+` FROM matches m WHERE m.stage_id=$1 AND m.status='completed' ORDER BY m.round,m.position`, id)
	return v, dbError(err)
}
func (r *Repository) ClearStageStandings(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM standings WHERE stage_id=$1`, id)
	return dbError(err)
}
func (r *Repository) InsertStanding(ctx context.Context, tournamentID, stageID uuid.UUID, groupID *uuid.UUID, v model.Standing) error {
	_, err := r.db.Exec(ctx, `INSERT INTO standings(stage_id,tournament_id,group_id,application_id,position,played,wins,draws,losses,score_for,score_against,points) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, stageID, tournamentID, groupID, v.ApplicationId, v.Position, v.Played, v.Wins, v.Draws, v.Losses, v.ScoreFor, v.ScoreAgainst, v.Points)
	return dbError(err)
}

func optionalUUIDEqual(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
