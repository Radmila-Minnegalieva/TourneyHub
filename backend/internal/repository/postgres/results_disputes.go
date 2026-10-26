package postgres

import (
	"context"
	"encoding/json"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

const disputeJSON = `jsonb_build_object('id',d.id,'match_id',d.match_id,'report_id',d.report_id,'opened_by',d.opened_by,'reason',d.reason,'status',d.status,'resolution',d.resolution,'resolved_by',d.resolved_by,'created_at',d.created_at,'evidence_file_ids',coalesce((SELECT jsonb_agg(f.file_id ORDER BY f.file_id) FROM dispute_files f WHERE f.dispute_id=d.id),'[]'::jsonb))`

func (r *Repository) ReadDispute(ctx context.Context, matchID uuid.UUID) (model.Dispute, error) {
	var raw []byte
	var d model.Dispute
	if err := r.db.QueryRow(ctx, `SELECT `+disputeJSON+` FROM disputes d WHERE d.match_id=$1`, matchID).Scan(&raw); err != nil {
		return d, dbError(err)
	}
	err := json.Unmarshal(raw, &d)
	return d, err
}
func (r *Repository) InsertDispute(ctx context.Context, id, matchID, reportID, user uuid.UUID, reason string) error {
	_, err := r.db.Exec(ctx, `INSERT INTO disputes(id,match_id,report_id,opened_by,reason) VALUES($1,$2,$3,$4,$5)`, id, matchID, reportID, user, reason)
	return dbError(err)
}
func (r *Repository) InsertDisputeFiles(ctx context.Context, id uuid.UUID, files []uuid.UUID) error {
	for _, file := range files {
		if _, err := r.db.Exec(ctx, `INSERT INTO dispute_files(dispute_id,file_id) VALUES($1,$2)`, id, file); err != nil {
			return dbError(err)
		}
	}
	return nil
}
func (r *Repository) ResolveDispute(ctx context.Context, id, user uuid.UUID, reason string) error {
	_, err := r.db.Exec(ctx, `UPDATE disputes SET status='resolved',resolved_by=$2,resolution=$3,resolved_at=now() WHERE id=$1`, id, user, reason)
	return dbError(err)
}
func (r *Repository) ReadAssignedDisputePage(ctx context.Context, user uuid.UUID, admin bool, status *model.ListAssignedDisputesParamsStatus, limit, offset int) (model.DisputePage, error) {
	filter := ` FROM disputes d JOIN matches m ON m.id=d.match_id JOIN tournaments t ON t.id=m.tournament_id WHERE ($1 OR coalesce(m.referee_id,t.referee_id)=$2) AND ($3::text IS NULL OR d.status=$3)`
	args := []any{admin, user, status}
	result := model.DisputePage{Items: []model.Dispute{}, Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+filter, args...).Scan(&result.Total); err != nil {
		return result, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+disputeJSON+filter+` ORDER BY d.created_at DESC,d.id LIMIT $4 OFFSET $5`, append(args, limit, offset)...)
	if err != nil {
		return result, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var d model.Dispute
		if err = rows.Scan(&raw); err != nil {
			return result, dbError(err)
		}
		if err = json.Unmarshal(raw, &d); err != nil {
			return result, err
		}
		result.Items = append(result.Items, d)
	}
	return result, dbError(rows.Err())
}
func (r *Repository) ExpiredResultCandidates(ctx context.Context, limit int) ([]repository.AutoConfirmCandidate, error) {
	rows, err := r.db.Query(ctx, `SELECT r.id,r.match_id,m.tournament_id FROM result_reports r JOIN matches m ON m.id=r.match_id JOIN tournaments t ON t.id=m.tournament_id WHERE t.status='in_progress' AND t.deleted_at IS NULL AND r.status='pending' AND r.confirmation_deadline<=now() AND m.status='awaiting_confirmation' ORDER BY r.confirmation_deadline,r.id LIMIT $1`, limit)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	result := []repository.AutoConfirmCandidate{}
	for rows.Next() {
		var v repository.AutoConfirmCandidate
		if err = rows.Scan(&v.ReportID, &v.MatchID, &v.TournamentID); err != nil {
			return nil, dbError(err)
		}
		result = append(result, v)
	}
	return result, dbError(rows.Err())
}
