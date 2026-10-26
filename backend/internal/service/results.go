package service

import (
	"context"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func validateResult(ctx context.Context, q repository.Store, t model.Tournament, m model.Match, req model.ReportResultRequest) error {
	rules, err := q.ReadResultRules(ctx, t.Id, m.StageId)
	if err != nil {
		return err
	}
	games := make([]model.Game, len(req.Games))
	for i, g := range req.Games {
		games[i] = model.Game{Number: g.Number, Home: g.HomeScore, Away: g.AwayScore}
	}
	if err = model.ValidateSeries(int(t.Rules.BestOf), req.HomeScore, req.AwayScore, games, rules.AllowsDraw, rules.StageKind == model.StageKindPlayoffs); err != nil {
		return invalid(err.Error())
	}
	return nil
}

func validateEvidence(ctx context.Context, q repository.Store, ids *[]uuid.UUID, user uuid.UUID) error {
	if ids == nil {
		return nil
	}
	seen := map[uuid.UUID]bool{}
	for _, id := range *ids {
		if seen[id] {
			return invalid("duplicate evidence file")
		}
		seen[id] = true
		valid, err := q.ValidMatchEvidence(ctx, id, user)
		if err != nil {
			return err
		}
		if !valid {
			return invalid("evidence must be your own match_protocol file")
		}
	}
	return nil
}

func insertReport(ctx context.Context, q repository.Store, user uuid.UUID, t model.Tournament, m model.Match, req model.ReportResultRequest, source string, pending bool) (model.ResultReport, error) {
	report := model.ResultReport{Id: uuid.New(), MatchId: m.Id, AuthorId: user, HomeScore: req.HomeScore, AwayScore: req.AwayScore, Games: req.Games, EvidenceFileIds: []uuid.UUID{}, Status: model.ResultReportStatusPending, CreatedAt: time.Now().UTC()}
	if err := validateResult(ctx, q, t, m, req); err != nil {
		return report, err
	}
	if err := validateEvidence(ctx, q, req.EvidenceFileIds, user); err != nil {
		return report, err
	}
	if req.EvidenceFileIds != nil {
		report.EvidenceFileIds = *req.EvidenceFileIds
	}
	if pending {
		deadline := report.CreatedAt.Add(time.Duration(t.Rules.ConfirmationTimeoutHours) * time.Hour)
		report.ConfirmationDeadline = &deadline
	}
	var reportingSide *uuid.UUID
	if source == "captain" {
		side, err := matchCaptain(ctx, q, m, user)
		if err != nil {
			return report, err
		}
		if side == 1 {
			reportingSide = m.HomeApplicationId
		} else if side == 2 {
			reportingSide = m.AwayApplicationId
		}
	}
	return report, q.InsertResultReport(ctx, report, source, reportingSide)
}

func (s *Service) ListResultReports(ctx context.Context, id uuid.UUID) ([]model.ResultReport, error) {
	match, err := s.repo.ReadMatch(ctx, id, false)
	if err != nil {
		return nil, err
	}
	tournament, err := s.repo.ReadTournament(ctx, match.TournamentId, false)
	if err != nil {
		return nil, err
	}
	if err = matchReadAccess(ctx, s.repo, tournament, match); err != nil {
		return nil, err
	}
	return s.repo.ReadResultReports(ctx, id)
}

func (s *Service) ReportResult(ctx context.Context, id uuid.UUID, req model.ReportResultRequest) (model.ResultReport, error) {
	var result model.ResultReport
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		if err := tournamentPlaying(t); err != nil {
			return err
		}
		if err := matchVersion(m, req.Version); err != nil {
			return err
		}
		if err := requireScheduled(m); err != nil {
			return err
		}
		side, err := matchCaptain(ctx, tx, m, currentUser(ctx))
		if err != nil {
			return err
		}
		if side == 0 {
			return forbidden("only a participating captain can report results")
		}
		result, err = insertReport(ctx, tx, currentUser(ctx), t, m, req, "captain", true)
		if err != nil {
			return err
		}
		if err = tx.SetMatchStatus(ctx, id, model.MatchStatusAwaitingConfirmation); err != nil {
			return err
		}
		m.Status = model.MatchStatusAwaitingConfirmation
		m.Version++
		if err = matchChanged(ctx, tx, currentUser(ctx), m, "result.reported"); err != nil {
			return err
		}
		return notifyMatch(ctx, tx, m, "result_confirmation", "Результат матча ожидает подтверждения соперником")
	})
	return result, err
}
func opposingCaptain(ctx context.Context, q repository.Store, m model.Match, report model.ResultReport) error {
	side, err := matchCaptain(ctx, q, m, currentUser(ctx))
	if err != nil {
		return err
	}
	if side == 0 {
		return forbidden("only the opposing captain may act")
	}
	reportingSide, err := q.ReadReportingApplication(ctx, report.Id)
	if err != nil {
		return err
	}
	ownSide := m.HomeApplicationId
	if side == 2 {
		ownSide = m.AwayApplicationId
	}
	if optionalUUIDEqual(reportingSide, ownSide) || currentUser(ctx) == report.AuthorId {
		return forbidden("the reporting side cannot confirm or dispute its own report")
	}
	return nil
}

func (s *Service) ConfirmResult(ctx context.Context, id, reportID uuid.UUID, req model.VersionRequest) (model.Match, error) {
	var result model.Match
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		r, err := tx.ReadResultReport(ctx, reportID, id)
		if err != nil {
			return err
		}
		if err = opposingCaptain(ctx, tx, m, r); err != nil {
			return err
		}
		if m.Status == model.MatchStatusCompleted && r.Status == model.ResultReportStatusConfirmed {
			result = m
			return nil
		}
		if err = matchVersion(m, req.Version); err != nil {
			return err
		}
		if err = tournamentPlaying(t); err != nil {
			return err
		}
		if m.Status != model.MatchStatusAwaitingConfirmation || r.Status != model.ResultReportStatusPending {
			return conflict("report is no longer pending")
		}
		result, err = finalizeResult(ctx, tx, currentUser(ctx), t, m, r)
		return err
	})
	return result, err
}
func (s *Service) SetRefereeResult(ctx context.Context, id uuid.UUID, req model.ReportResultRequest) (model.Match, error) {
	var result model.Match
	err := s.matchTransaction(ctx, id, func(tx repository.Store, t model.Tournament, m model.Match) error {
		if err := requireMatchReferee(ctx, t, m); err != nil {
			return err
		}
		if err := matchVersion(m, req.Version); err != nil {
			return err
		}
		if err := tournamentPlaying(t); err != nil {
			return err
		}
		if m.Status == model.MatchStatusCompleted || m.Status == model.MatchStatusBye || m.Status == model.MatchStatusDisputed {
			return conflict("use dispute resolution for a disputed or completed match")
		}
		scheduled := m
		scheduled.Status = model.MatchStatusScheduled
		if err := requireScheduled(scheduled); err != nil {
			return err
		}

		if err := tx.SupersedeResultReports(ctx, id, nil, []model.ResultReportStatus{model.ResultReportStatusPending}); err != nil {
			return err
		}
		r, err := insertReport(ctx, tx, currentUser(ctx), t, m, req, "referee", false)
		if err != nil {
			return err
		}
		result, err = finalizeResult(ctx, tx, currentUser(ctx), t, m, r)
		return err
	})
	return result, err
}
func finalizeResult(ctx context.Context, tx repository.Store, actor uuid.UUID, t model.Tournament, m model.Match, r model.ResultReport) (model.Match, error) {
	if err := tx.LockRatings(ctx, t.DisciplineId); err != nil {
		return m, err
	}
	correction := m.HomeScore != nil
	winner := scoreWinner(m, r.HomeScore, r.AwayScore)
	if err := updateProgression(ctx, tx, m, winner); err != nil {
		return m, err
	}
	if err := tx.SupersedeResultReports(ctx, m.Id, &r.Id, []model.ResultReportStatus{model.ResultReportStatusPending, model.ResultReportStatusConfirmed, model.ResultReportStatusDisputed}); err != nil {
		return m, err
	}
	var by *uuid.UUID
	if actor != uuid.Nil {
		by = &actor
	}
	if err := tx.ConfirmResultReport(ctx, r.Id, by); err != nil {
		return m, err
	}
	if err := tx.SaveFinalResult(ctx, m.Id, r.HomeScore, r.AwayScore, winner, r.Id); err != nil {
		return m, err
	}
	m.Status = model.MatchStatusCompleted
	m.HomeScore = &r.HomeScore
	m.AwayScore = &r.AwayScore
	m.WinnerApplicationId = winner
	m.Version++
	if err := recordMatchRating(ctx, tx, t.Id, t.DisciplineId, ratingMatch{id: m.Id, homeScore: r.HomeScore, awayScore: r.AwayScore}, correction); err != nil {
		return m, err
	}
	if err := rebuildStandings(ctx, tx, t, m.StageId); err != nil {
		return m, err
	}
	if err := finishStage(ctx, tx, t, m.StageId); err != nil {
		return m, err
	}
	if err := matchChanged(ctx, tx, actor, m, "result.confirmed"); err != nil {
		return m, err
	}
	return m, notifyMatch(ctx, tx, m, "result_confirmed", "Зафиксирован итоговый результат матча")
}

func scoreWinner(m model.Match, home, away int) *uuid.UUID {
	if home > away {
		return m.HomeApplicationId
	}
	if away > home {
		return m.AwayApplicationId
	}
	return nil
}
