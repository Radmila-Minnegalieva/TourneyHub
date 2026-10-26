package postgres

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

func (r *Repository) ReadRatedMatches(ctx context.Context, id uuid.UUID) ([]repository.RatedMatch, error) {
	rows, err := r.db.Query(ctx, `SELECT m.id,m.tournament_id,ha.team_id,aa.team_id,m.home_score,m.away_score FROM matches m JOIN tournaments t ON t.id=m.tournament_id JOIN applications ha ON ha.id=m.home_application_id JOIN applications aa ON aa.id=m.away_application_id WHERE t.discipline_id=$1 AND m.final_report_id IS NOT NULL ORDER BY (SELECT min(r.confirmed_at) FROM result_reports r WHERE r.match_id=m.id),m.id`, id)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	result := []repository.RatedMatch{}
	for rows.Next() {
		var v repository.RatedMatch
		if err = rows.Scan(&v.ID, &v.TournamentID, &v.HomeTeamID, &v.AwayTeamID, &v.HomeScore, &v.AwayScore); err != nil {
			return nil, dbError(err)
		}
		result = append(result, v)
	}
	return result, dbError(rows.Err())
}
func (r *Repository) MatchTeams(ctx context.Context, id uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	var home, away uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT ha.team_id,aa.team_id FROM matches m JOIN applications ha ON ha.id=m.home_application_id JOIN applications aa ON aa.id=m.away_application_id WHERE m.id=$1`, id).Scan(&home, &away)
	return home, away, dbError(err)
}
func (r *Repository) UpsertRating(ctx context.Context, team, discipline uuid.UUID, value float64, played int) error {
	_, err := r.db.Exec(ctx, `INSERT INTO ratings(team_id,discipline_id,value,played) VALUES($1,$2,$3,$4) ON CONFLICT(team_id,discipline_id) DO UPDATE SET value=excluded.value,played=excluded.played,version=ratings.version+1`, team, discipline, value, played)
	return dbError(err)
}
func (r *Repository) ReadLatestRatingRevision(ctx context.Context, match, team uuid.UUID) (repository.RatingRevision, error) {
	var v repository.RatingRevision
	err := r.db.QueryRow(ctx, `SELECT before,delta,revision FROM rating_changes WHERE match_id=$1 AND team_id=$2 ORDER BY revision DESC LIMIT 1`, match, team).Scan(&v.Before, &v.Delta, &v.Revision)
	return v, dbError(err)
}
func (r *Repository) InsertRatingRevision(ctx context.Context, discipline uuid.UUID, m repository.RatedMatch, team uuid.UUID, before, delta float64, revision int) error {
	_, err := r.db.Exec(ctx, `INSERT INTO rating_changes(match_id,tournament_id,team_id,discipline_id,before,delta,after,revision) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, m.ID, m.TournamentID, team, discipline, before, delta, before+delta, revision)
	return dbError(err)
}
func (r *Repository) ReadRating(ctx context.Context, team, discipline uuid.UUID) (repository.RatingSnapshot, error) {
	var v repository.RatingSnapshot
	err := r.db.QueryRow(ctx, `SELECT value,played FROM ratings WHERE team_id=$1 AND discipline_id=$2`, team, discipline).Scan(&v.Value, &v.Played)
	return v, dbError(err)
}
