package repository

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
	"time"
)

// StandingParticipant связывает участника таблицы с группой и текущим Эло.
type StandingParticipant struct {
	ApplicationID uuid.UUID
	GroupID       *uuid.UUID
	Elo           float64
}

// RatedMatch хранит данные завершённого матча для воспроизведения истории Эло.
type RatedMatch struct {
	ID           uuid.UUID
	TournamentID uuid.UUID
	HomeTeamID   uuid.UUID
	AwayTeamID   uuid.UUID
	HomeScore    int
	AwayScore    int
}

type RatingSnapshot struct {
	Value  float64
	Played int
}

type RatingRevision struct {
	Before   float64
	Delta    float64
	Revision int
}

type ResultValidationRules struct {
	AllowsDraw bool
	StageKind  model.StageKind
}

type AutoConfirmCandidate struct {
	ReportID     uuid.UUID
	MatchID      uuid.UUID
	TournamentID uuid.UUID
}

// Competition описывает операции хранения матчей, сеток, результатов и таблиц.
// Транзакцию предоставляет Store.WithinTx; бизнес-правила выполняет service.
type Competition interface {
	ReadMatch(context.Context, uuid.UUID, bool) (model.Match, error)
	MatchTournamentID(context.Context, uuid.UUID) (uuid.UUID, error)
	ReadMatchPage(context.Context, uuid.UUID, model.ListMatchesParams, int, int) (model.MatchPage, error)
	ReadBracket(context.Context, uuid.UUID) (model.Bracket, error)
	MatchCaptainSide(context.Context, *uuid.UUID, *uuid.UUID, uuid.UUID) (int, error)
	IsMatchRosterPlayer(context.Context, *uuid.UUID, *uuid.UUID, uuid.UUID) (bool, error)
	MatchNotificationRecipients(context.Context, *uuid.UUID, *uuid.UUID) ([]uuid.UUID, error)
	AcceptedApplicationsByRating(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error)
	HasTournamentReports(context.Context, uuid.UUID) (bool, error)
	InsertDrawPreview(context.Context, model.DrawPreview, uuid.UUID, model.SeedingRequestMethod) error
	ReadDrawPreview(context.Context, uuid.UUID, uuid.UUID) (model.DrawPreview, bool, error)
	ClearTournamentDraw(context.Context, uuid.UUID) error
	SetApplicationSeed(context.Context, uuid.UUID, int) error
	InsertStage(context.Context, model.Stage) error
	InsertGroup(context.Context, uuid.UUID, model.Group) error
	InsertMatches(context.Context, []model.Match) error
	ConfirmDrawPreview(context.Context, uuid.UUID) error
	RoundMatches(context.Context, uuid.UUID, uuid.UUID, int) ([]model.Match, error)
	ClearRoundSchedule(context.Context, uuid.UUID, uuid.UUID, int) error
	SetMatchSchedule(context.Context, uuid.UUID, *time.Time, *string, *string) error
	HasParticipantScheduleConflict(context.Context, model.Match, time.Time, *uuid.UUID) (bool, error)
	HasScheduleDependencyConflict(context.Context, model.Match, time.Time) (bool, error)
	ValidReferee(context.Context, uuid.UUID) (bool, error)
	SetMatchReferee(context.Context, uuid.UUID, uuid.UUID) error
	ReadResultReport(context.Context, uuid.UUID, uuid.UUID) (model.ResultReport, error)
	ReadResultReports(context.Context, uuid.UUID) ([]model.ResultReport, error)
	ReadResultRules(context.Context, uuid.UUID, uuid.UUID) (ResultValidationRules, error)
	ValidMatchEvidence(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	InsertResultReport(context.Context, model.ResultReport, string, *uuid.UUID) error
	SetMatchStatus(context.Context, uuid.UUID, model.MatchStatus) error
	ReadReportingApplication(context.Context, uuid.UUID) (*uuid.UUID, error)
	SupersedeResultReports(context.Context, uuid.UUID, *uuid.UUID, []model.ResultReportStatus) error
	SetReportStatus(context.Context, uuid.UUID, model.ResultReportStatus) error
	ConfirmResultReport(context.Context, uuid.UUID, *uuid.UUID) error
	SaveFinalResult(context.Context, uuid.UUID, int, int, *uuid.UUID, uuid.UUID) error
	ReadDispute(context.Context, uuid.UUID) (model.Dispute, error)
	InsertDispute(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, string) error
	InsertDisputeFiles(context.Context, uuid.UUID, []uuid.UUID) error
	ResolveDispute(context.Context, uuid.UUID, uuid.UUID, string) error
	ReadAssignedDisputePage(context.Context, uuid.UUID, bool, *model.ListAssignedDisputesParamsStatus, int, int) (model.DisputePage, error)
	ExpiredResultCandidates(context.Context, int) ([]AutoConfirmCandidate, error)
	ReadStandingParticipants(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]StandingParticipant, error)
	CompletedStageMatches(context.Context, uuid.UUID) ([]model.Match, error)
	ClearStageStandings(context.Context, uuid.UUID) error
	InsertStanding(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID, model.Standing) error
	ReadStandings(context.Context, uuid.UUID) ([]model.Standings, error)
	SetMatchParticipant(context.Context, uuid.UUID, model.MatchNextMatchSlot, *uuid.UUID) error
	StageRemainingMatches(context.Context, uuid.UUID) (int, error)
	CompleteStage(context.Context, uuid.UUID) (model.StageKind, error)
	RemainingTournamentStages(context.Context, uuid.UUID) (int, error)
	ReadPlayoffStage(context.Context, uuid.UUID) (model.Stage, error)
	StageMatches(context.Context, uuid.UUID) ([]model.Match, error)
	DeleteStageMatches(context.Context, uuid.UUID) error
	SetStageStatus(context.Context, uuid.UUID, model.StageStatus) error
	TournamentMatchCounts(context.Context, uuid.UUID) (int, int, error)
	CompleteTournamentStages(context.Context, uuid.UUID) error
	ClearFinalPlaces(context.Context, uuid.UUID) error
	SetFinalPlace(context.Context, uuid.UUID, int, bool) error
	UnplacedStandings(context.Context, uuid.UUID) ([]model.Standing, error)
	PlacedApplicationsCount(context.Context, uuid.UUID) (int, error)
	PlayoffMatches(context.Context, uuid.UUID) ([]model.Match, error)
	ReadRatedMatches(context.Context, uuid.UUID) ([]RatedMatch, error)
	MatchTeams(context.Context, uuid.UUID) (uuid.UUID, uuid.UUID, error)
	UpsertRating(context.Context, uuid.UUID, uuid.UUID, float64, int) error
	ReadLatestRatingRevision(context.Context, uuid.UUID, uuid.UUID) (RatingRevision, error)
	InsertRatingRevision(context.Context, uuid.UUID, RatedMatch, uuid.UUID, float64, float64, int) error
	ReadRating(context.Context, uuid.UUID, uuid.UUID) (RatingSnapshot, error)
}
