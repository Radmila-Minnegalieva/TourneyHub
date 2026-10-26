package service

import (
	"context"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func (s *Service) GetDispute(ctx context.Context, id uuid.UUID) (model.Dispute, error) {
	var result model.Dispute
	m, err := s.repo.ReadMatch(ctx, id, false)
	if err != nil {
		return result, err
	}
	tournament, err := s.repo.ReadTournament(ctx, m.TournamentId, false)
	if err != nil {
		return result, err
	}
	if err = matchReadAccess(ctx, s.repo, tournament, m); err != nil {
		return result, err
	}
	return s.repo.ReadDispute(ctx, id)
}

func (s *Service) DisputeResult(ctx context.Context, id, reportID uuid.UUID, req model.DisputeRequest) (model.Dispute, error) {
	var result model.Dispute
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		var err error
		result, err = openDispute(ctx, tx, t, m, reportID, req)
		return err
	})
	return result, err
}

func openDispute(ctx context.Context, q repository.Store, t model.Tournament, m model.Match, reportID uuid.UUID, req model.DisputeRequest) (model.Dispute, error) {
	var result model.Dispute
	report, err := disputableReport(ctx, q, t, m, reportID, req.Version)
	if err != nil {
		return result, err
	}
	if err = createDispute(ctx, q, currentUser(ctx), m.Id, report.Id, req); err != nil {
		return result, err
	}
	if err = recordDisputedMatch(ctx, q, t, m, report.Id); err != nil {
		return result, err
	}
	return q.ReadDispute(ctx, m.Id)
}

func disputableReport(ctx context.Context, q repository.Store, t model.Tournament, m model.Match, reportID uuid.UUID, version int) (model.ResultReport, error) {
	var report model.ResultReport
	if err := matchVersion(m, version); err != nil {
		return report, err
	}
	if t.Status != model.TournamentStatusInProgress && t.Status != model.TournamentStatusCompleted {
		return report, conflict("tournament cannot accept disputes")
	}
	report, err := q.ReadResultReport(ctx, reportID, m.Id)
	if err != nil {
		return report, err
	}
	if m.Status == model.MatchStatusCompleted {
		return report, requireCompletedDisputeCaptain(ctx, q, m, report)
	}
	return report, requirePendingDisputeCaptain(ctx, q, m, report)
}

func requireCompletedDisputeCaptain(ctx context.Context, q repository.Store, m model.Match, r model.ResultReport) error {
	side, err := matchCaptain(ctx, q, m, currentUser(ctx))
	if err != nil {
		return err
	}
	if side == 0 {
		return forbidden("participating captain only")
	}
	if r.Status != model.ResultReportStatusConfirmed {
		return conflict("only the final report can be disputed")
	}
	return nil
}

func requirePendingDisputeCaptain(ctx context.Context, q repository.Store, m model.Match, r model.ResultReport) error {
	if m.Status != model.MatchStatusAwaitingConfirmation || r.Status != model.ResultReportStatusPending {
		return conflict("report cannot be disputed")
	}
	if err := opposingCaptain(ctx, q, m, r); err != nil {
		return err
	}
	if r.ConfirmationDeadline == nil || !r.ConfirmationDeadline.After(time.Now()) {
		return conflict("confirmation deadline has expired")
	}
	return nil
}

func createDispute(ctx context.Context, q repository.Store, user, matchID, reportID uuid.UUID, req model.DisputeRequest) error {
	if err := validateEvidence(ctx, q, req.EvidenceFileIds, user); err != nil {
		return err
	}
	id := uuid.New()
	if err := q.InsertDispute(ctx, id, matchID, reportID, user, req.Reason); err != nil {
		return err
	}
	return saveDisputeFiles(ctx, q, id, req.EvidenceFileIds)
}

func saveDisputeFiles(ctx context.Context, q repository.Store, id uuid.UUID, files *[]uuid.UUID) error {
	if files == nil {
		return nil
	}
	return q.InsertDisputeFiles(ctx, id, *files)
}

func recordDisputedMatch(ctx context.Context, q repository.Store, t model.Tournament, m model.Match, reportID uuid.UUID) error {
	if err := q.SetReportStatus(ctx, reportID, model.ResultReportStatusDisputed); err != nil {
		return err
	}
	if err := q.SetMatchStatus(ctx, m.Id, model.MatchStatusDisputed); err != nil {
		return err
	}
	m.Status = model.MatchStatusDisputed
	m.Version++
	if err := matchChanged(ctx, q, currentUser(ctx), m, "result.disputed"); err != nil {
		return err
	}
	return notifyDisputeReferee(ctx, q, t, m)
}

func notifyDisputeReferee(ctx context.Context, q repository.Store, t model.Tournament, m model.Match) error {
	referee := m.RefereeId
	if referee == nil {
		referee = t.RefereeId
	}
	if referee == nil {
		return nil
	}
	return notify(ctx, q, *referee, "result_disputed", "Открыт спор по результату матча", m.Id)
}

func (s *Service) ResolveDispute(ctx context.Context, id uuid.UUID, req model.ResolveDisputeRequest) (model.Match, error) {
	var result model.Match
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		if err := requireMatchReferee(ctx, t, m); err != nil {
			return err
		}
		if t.Status != model.TournamentStatusInProgress && t.Status != model.TournamentStatusCompleted {
			return conflict("tournament cannot accept dispute resolutions")
		}
		if err := matchVersion(m, req.Version); err != nil {
			return err
		}
		if m.Status != model.MatchStatusDisputed {
			return conflict("match has no unresolved dispute")
		}

		d, err := tx.ReadDispute(ctx, id)
		if err != nil {
			return err
		}
		if d.Status != model.DisputeStatusOpen {
			return conflict("dispute already resolved")
		}
		if err = tx.SupersedeResultReports(ctx, id, nil, []model.ResultReportStatus{model.ResultReportStatusPending, model.ResultReportStatusDisputed}); err != nil {
			return err
		}
		report := model.ReportResultRequest{Version: req.Version, HomeScore: req.HomeScore, AwayScore: req.AwayScore, Games: req.Games, EvidenceFileIds: req.EvidenceFileIds}
		r, err := insertReport(ctx, tx, currentUser(ctx), t, m, report, "dispute", false)
		if err != nil {
			return err
		}
		result, err = finalizeResult(ctx, tx, currentUser(ctx), t, m, r)
		if err != nil {
			return err
		}
		if err = tx.ResolveDispute(ctx, d.Id, currentUser(ctx), req.Reason); err != nil {
			return err
		}
		return tx.Audit(ctx, currentUser(ctx), "dispute.resolved", "dispute", d.Id, d, map[string]any{"resolution": req.Reason, "report_id": r.Id})
	})
	return result, err
}
func (s *Service) ListAssignedDisputes(ctx context.Context, p model.ListAssignedDisputesParams) (model.DisputePage, error) {
	if err := requireRole(ctx, model.Referee); err != nil {
		return model.DisputePage{}, err
	}
	limit, offset := pagination(p.Limit, p.Offset)
	return s.repo.ReadAssignedDisputePage(ctx, currentUser(ctx), hasRole(ctx, model.Admin), p.Status, limit, offset)
}

// AutoConfirm атомарно фиксирует просроченные протоколы. Блокировки турнира и
// матча совпадают с пользовательским подтверждением и открытием спора.
func (s *Service) AutoConfirm(ctx context.Context) error {
	candidates, err := s.repo.ExpiredResultCandidates(ctx, 100)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if err = s.autoConfirmOne(ctx, candidate.ReportID, candidate.MatchID, candidate.TournamentID); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) autoConfirmOne(ctx context.Context, reportID, matchID, tournamentID uuid.UUID) error {
	return s.repo.WithinTx(ctx, func(tx repository.Store) error {
		tournament, err := tx.ReadTournament(ctx, tournamentID, true)
		if ignoreNoRows(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if tournament.Status != model.TournamentStatusInProgress {
			return nil
		}
		match, err := tx.ReadMatch(ctx, matchID, true)
		if err != nil {
			return err
		}
		report, err := tx.ReadResultReport(ctx, reportID, matchID)
		if err != nil {
			return err
		}
		if match.Status != model.MatchStatusAwaitingConfirmation || report.Status != model.ResultReportStatusPending || report.ConfirmationDeadline == nil || report.ConfirmationDeadline.After(time.Now()) {
			return nil
		}
		if err = tournamentPlaying(tournament); err != nil {
			return err
		}
		_, err = finalizeResult(ctx, tx, uuid.Nil, tournament, match, report)
		return err
	})
}
