package postgres

import (
	"context"
	"encoding/json"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

const reportJSON = `jsonb_build_object('id',r.id,'match_id',r.match_id,'author_id',r.author_id,'home_score',r.home_score,'away_score',r.away_score,'status',r.status,'confirmation_deadline',r.confirmation_deadline,'created_at',r.created_at,'games',coalesce((SELECT jsonb_agg(jsonb_build_object('number',g.number,'home_score',g.home_score,'away_score',g.away_score) ORDER BY g.number) FROM report_games g WHERE g.report_id=r.id),'[]'::jsonb),'evidence_file_ids',coalesce((SELECT jsonb_agg(f.file_id ORDER BY f.file_id) FROM report_files f WHERE f.report_id=r.id),'[]'::jsonb))`

func readReport(ctx context.Context, q DBTX, id, matchID uuid.UUID) (model.ResultReport, error) {
	var raw []byte
	var result model.ResultReport
	err := q.QueryRow(ctx, `SELECT `+reportJSON+` FROM result_reports r WHERE r.id=$1 AND r.match_id=$2`, id, matchID).Scan(&raw)
	if err != nil {
		return result, dbError(err)
	}
	err = json.Unmarshal(raw, &result)
	return result, err
}
func reportList(ctx context.Context, q DBTX, id uuid.UUID) ([]model.ResultReport, error) {
	rows, err := q.Query(ctx, `SELECT `+reportJSON+` FROM result_reports r WHERE r.match_id=$1 ORDER BY r.created_at,r.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.ResultReport{}
	for rows.Next() {
		var raw []byte
		var r model.ResultReport
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(raw, &r); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
func (r *Repository) ReadResultReport(ctx context.Context, id, matchID uuid.UUID) (model.ResultReport, error) {
	v, err := readReport(ctx, r.db, id, matchID)
	return v, dbError(err)
}
func (r *Repository) ReadResultReports(ctx context.Context, id uuid.UUID) ([]model.ResultReport, error) {
	v, err := reportList(ctx, r.db, id)
	return v, dbError(err)
}
func (r *Repository) ReadResultRules(ctx context.Context, tournamentID, stageID uuid.UUID) (repository.ResultValidationRules, error) {
	var v repository.ResultValidationRules
	err := r.db.QueryRow(ctx, `SELECT t.allows_draw,s.kind FROM tournaments t JOIN stages s ON s.tournament_id=t.id WHERE t.id=$1 AND s.id=$2`, tournamentID, stageID).Scan(&v.AllowsDraw, &v.StageKind)
	return v, dbError(err)
}
func (r *Repository) ValidMatchEvidence(ctx context.Context, file, user uuid.UUID) (bool, error) {
	var valid bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM files WHERE id=$1 AND owner_id=$2 AND purpose='match_protocol')`, file, user).Scan(&valid)
	return valid, dbError(err)
}
func (r *Repository) InsertResultReport(ctx context.Context, v model.ResultReport, source string, reportingSide *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `INSERT INTO result_reports(id,match_id,author_id,source,home_score,away_score,confirmation_deadline,reporting_application_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, v.Id, v.MatchId, v.AuthorId, source, v.HomeScore, v.AwayScore, v.ConfirmationDeadline, reportingSide)
	if err != nil {
		return dbError(err)
	}
	for _, g := range v.Games {
		if _, err = r.db.Exec(ctx, `INSERT INTO report_games(report_id,number,home_score,away_score) VALUES($1,$2,$3,$4)`, v.Id, g.Number, g.HomeScore, g.AwayScore); err != nil {
			return dbError(err)
		}
	}
	for _, id := range v.EvidenceFileIds {
		if _, err = r.db.Exec(ctx, `INSERT INTO report_files(report_id,file_id) VALUES($1,$2)`, v.Id, id); err != nil {
			return dbError(err)
		}
	}
	return nil
}
func (r *Repository) SetMatchStatus(ctx context.Context, id uuid.UUID, status model.MatchStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET status=$2,version=version+1,updated_at=now() WHERE id=$1`, id, status)
	return dbError(err)
}
func (r *Repository) ReadReportingApplication(ctx context.Context, id uuid.UUID) (*uuid.UUID, error) {
	var side *uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT reporting_application_id FROM result_reports WHERE id=$1`, id).Scan(&side)
	return side, dbError(err)
}
func (r *Repository) SupersedeResultReports(ctx context.Context, matchID uuid.UUID, exceptID *uuid.UUID, statuses []model.ResultReportStatus) error {
	values := make([]string, len(statuses))
	for i, s := range statuses {
		values[i] = string(s)
	}
	_, err := r.db.Exec(ctx, `UPDATE result_reports SET status='superseded' WHERE match_id=$1 AND ($2::uuid IS NULL OR id<>$2) AND status=ANY($3::text[])`, matchID, exceptID, values)
	return dbError(err)
}
func (r *Repository) SetReportStatus(ctx context.Context, id uuid.UUID, status model.ResultReportStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE result_reports SET status=$2 WHERE id=$1`, id, status)
	return dbError(err)
}
func (r *Repository) ConfirmResultReport(ctx context.Context, id uuid.UUID, actor *uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE result_reports SET status='confirmed',confirmed_by=$2,confirmed_at=coalesce(confirmed_at,clock_timestamp()) WHERE id=$1`, id, actor)
	return dbError(err)
}
func (r *Repository) SaveFinalResult(ctx context.Context, id uuid.UUID, home, away int, winner *uuid.UUID, reportID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET status='completed',home_score=$2,away_score=$3,winner_application_id=$4,final_report_id=$5,version=version+1,updated_at=now() WHERE id=$1`, id, home, away, winner, reportID)
	return dbError(err)
}
