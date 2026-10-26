package postgres

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const matchColumns = `m.id,m.tournament_id,m.stage_id,m.group_id,m.round,m.position,m.home_application_id,m.away_application_id,m.winner_application_id,m.next_match_id,m.next_match_slot,m.loser_next_match_id,m.loser_next_match_slot,m.starts_at,m.venue,m.online_url,m.referee_id,m.status,m.home_score,m.away_score,m.version`

func scanMatch(row pgx.Row) (model.Match, error) {
	var m model.Match
	err := row.Scan(&m.Id, &m.TournamentId, &m.StageId, &m.GroupId, &m.Round, &m.Position, &m.HomeApplicationId, &m.AwayApplicationId, &m.WinnerApplicationId, &m.NextMatchId, &m.NextMatchSlot, &m.LoserNextMatchId, &m.LoserNextMatchSlot, &m.StartsAt, &m.Venue, &m.OnlineUrl, &m.RefereeId, &m.Status, &m.HomeScore, &m.AwayScore, &m.Version)
	return m, dbError(err)
}

func readMatch(ctx context.Context, q DBTX, id uuid.UUID, lock bool) (model.Match, error) {
	sql := `SELECT ` + matchColumns + ` FROM matches m WHERE m.id=$1`
	if lock {
		sql += ` FOR UPDATE`
	}
	return scanMatch(q.QueryRow(ctx, sql, id))
}

func matchList(ctx context.Context, q DBTX, query string, args ...any) ([]model.Match, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Match, 0)
	for rows.Next() {
		m, e := scanMatch(rows)
		if e != nil {
			return nil, e
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func readBracket(ctx context.Context, q DBTX, id uuid.UUID) (model.Bracket, error) {
	result := model.Bracket{Stages: []model.Stage{}, Groups: []model.Group{}, Matches: []model.Match{}}
	rows, err := q.Query(ctx, `SELECT id,tournament_id,kind,position,status FROM stages WHERE tournament_id=$1 ORDER BY position`, id)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var s model.Stage
		if err = rows.Scan(&s.Id, &s.TournamentId, &s.Kind, &s.Position, &s.Status); err != nil {
			rows.Close()
			return result, err
		}
		result.Stages = append(result.Stages, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	rows, err = q.Query(ctx, `SELECT g.id,g.stage_id,g.name,coalesce(array_agg(p.application_id ORDER BY a.seed) FILTER(WHERE p.application_id IS NOT NULL),'{}'::uuid[]) FROM tournament_groups g LEFT JOIN group_participants p ON p.group_id=g.id LEFT JOIN applications a ON a.id=p.application_id WHERE g.tournament_id=$1 GROUP BY g.id ORDER BY g.name,g.id`, id)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var g model.Group
		if err = rows.Scan(&g.Id, &g.StageId, &g.Name, &g.ApplicationIds); err != nil {
			rows.Close()
			return result, err
		}
		result.Groups = append(result.Groups, g)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return result, err
	}
	result.Matches, err = matchList(ctx, q, `SELECT `+matchColumns+` FROM matches m JOIN stages s ON s.id=m.stage_id WHERE m.tournament_id=$1 ORDER BY s.position,m.round,m.position`, id)
	return result, err
}
func (r *Repository) ReadMatch(ctx context.Context, id uuid.UUID, lock bool) (model.Match, error) {
	return readMatch(ctx, r.db, id, lock)
}
func (r *Repository) MatchTournamentID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var tid uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT tournament_id FROM matches WHERE id=$1`, id).Scan(&tid)
	return tid, dbError(err)
}
func (r *Repository) ReadMatchPage(ctx context.Context, id uuid.UUID, p model.ListMatchesParams, limit, offset int) (model.MatchPage, error) {
	result := model.MatchPage{Items: []model.Match{}, Limit: limit, Offset: offset}
	filter := ` FROM matches m WHERE m.tournament_id=$1 AND ($2::uuid IS NULL OR m.stage_id=$2) AND ($3::uuid IS NULL OR m.group_id=$3) AND ($4::int IS NULL OR m.round=$4) AND ($5::text IS NULL OR m.status=$5)`
	args := []any{id, p.StageId, p.GroupId, p.Round, p.Status}
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+filter, args...).Scan(&result.Total); err != nil {
		return result, dbError(err)
	}
	items, err := matchList(ctx, r.db, `SELECT `+matchColumns+filter+` ORDER BY m.starts_at NULLS LAST,m.stage_id,m.round,m.position,m.id LIMIT $6 OFFSET $7`, append(args, limit, offset)...)
	result.Items = items
	return result, dbError(err)
}
func (r *Repository) ReadBracket(ctx context.Context, id uuid.UUID) (model.Bracket, error) {
	v, err := readBracket(ctx, r.db, id)
	return v, dbError(err)
}
func (r *Repository) MatchCaptainSide(ctx context.Context, home, away *uuid.UUID, user uuid.UUID) (int, error) {
	var side int
	err := r.db.QueryRow(ctx, `SELECT CASE WHEN EXISTS(SELECT 1 FROM applications a JOIN teams t ON t.id=a.team_id WHERE a.id=$1 AND t.captain_id=$3) THEN 1 WHEN EXISTS(SELECT 1 FROM applications a JOIN teams t ON t.id=a.team_id WHERE a.id=$2 AND t.captain_id=$3) THEN 2 ELSE 0 END`, home, away, user).Scan(&side)
	return side, dbError(err)
}
func (r *Repository) IsMatchRosterPlayer(ctx context.Context, home, away *uuid.UUID, user uuid.UUID) (bool, error) {
	var member bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM application_players WHERE application_id IN ($1,$2) AND user_id=$3)`, home, away, user).Scan(&member)
	return member, dbError(err)
}
func (r *Repository) MatchNotificationRecipients(ctx context.Context, home, away *uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT user_id FROM (SELECT user_id FROM application_players WHERE application_id IN ($1,$2) UNION SELECT t.captain_id AS user_id FROM applications a JOIN teams t ON t.id=a.team_id WHERE a.id IN ($1,$2)) recipients ORDER BY user_id`, home, away)
	if err != nil {
		return nil, dbError(err)
	}
	result, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	return result, dbError(err)
}
