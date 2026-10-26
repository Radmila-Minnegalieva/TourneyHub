package service

import (
	"context"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"

	"github.com/google/uuid"
)

// Identity задаёт сценарии учётных записей и проверку сессии.
type Identity interface {
	Register(ctx context.Context, req model.RegisterRequest) error
	VerifyEmail(ctx context.Context, req model.TokenRequest) error
	Login(ctx context.Context, req model.LoginRequest, ip string) (model.Tokens, error)
	Authenticate(ctx context.Context, token string) (model.Principal, error)
	RefreshTokens(ctx context.Context, req model.TokenRequest) (model.Tokens, error)
	Logout(ctx context.Context, req model.TokenRequest) error
	RequestPasswordReset(ctx context.Context, req model.EmailRequest) error
	ResetPassword(ctx context.Context, req model.ResetPasswordRequest) error
	GetProfile(ctx context.Context) (model.Profile, error)
	UpdateProfile(ctx context.Context, req model.UpdateProfileRequest) (model.Profile, error)
	SearchUsers(ctx context.Context, params model.SearchUsersParams) (model.PublicUserPage, error)
}

// Teams задаёт работу с командами, их составом и приглашениями.
type Teams interface {
	ListTeams(ctx context.Context, params model.ListTeamsParams) (model.TeamPage, error)
	ListMyTeams(ctx context.Context, params model.ListMyTeamsParams) (model.TeamPage, error)
	GetTeam(ctx context.Context, id uuid.UUID) (model.Team, error)
	CreateTeam(ctx context.Context, req model.CreateTeamRequest) (model.Team, error)
	UpdateTeam(ctx context.Context, teamID uuid.UUID, req model.UpdateTeamRequest) (model.Team, error)
	DeleteTeam(ctx context.Context, teamID uuid.UUID, req model.VersionRequest) error
	ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]model.TeamMember, error)
	RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID, req model.VersionRequest) error
	TransferCaptain(ctx context.Context, teamID uuid.UUID, req model.TransferCaptainRequest) (model.Team, error)
	CreateInvitation(ctx context.Context, teamID uuid.UUID, req model.CreateInvitationRequest) (model.InvitationLink, error)
	ListMyInvitations(ctx context.Context, params model.ListMyInvitationsParams) (model.InvitationPage, error)
	DecideInvitation(ctx context.Context, invitationID uuid.UUID, req model.InvitationDecisionRequest) (model.Invitation, error)
	RevokeInvitation(ctx context.Context, teamID, invitationID uuid.UUID) error
}

// Tournaments задаёт жизненный цикл турнира и рассмотрение заявок.
type Tournaments interface {
	ListTournaments(ctx context.Context, params model.ListTournamentsParams) (model.TournamentPage, error)
	ListMyTournaments(ctx context.Context, params model.ListMyTournamentsParams) (model.TournamentPage, error)
	GetTournament(ctx context.Context, id uuid.UUID) (model.Tournament, error)
	CreateTournament(ctx context.Context, req model.CreateTournamentRequest) (model.Tournament, error)
	UpdateTournament(ctx context.Context, id uuid.UUID, req model.UpdateTournamentRequest) (model.Tournament, error)
	PublishTournament(ctx context.Context, id uuid.UUID, req model.VersionRequest) (model.Tournament, error)
	CancelTournament(ctx context.Context, id uuid.UUID, req model.ReasonRequest) (model.Tournament, error)
	CompleteTournament(ctx context.Context, id uuid.UUID, req model.VersionRequest) (model.Tournament, error)
	AssignTournamentReferee(ctx context.Context, id uuid.UUID, req model.AssignRefereeRequest) (model.Tournament, error)
	ListApplications(ctx context.Context, id uuid.UUID, params model.ListApplicationsParams) (model.ApplicationPage, error)
	ListParticipants(ctx context.Context, id uuid.UUID, params model.ListParticipantsParams) (model.ApplicationPage, error)
	ListMyApplications(ctx context.Context, params model.ListMyApplicationsParams) (model.ApplicationPage, error)
	SubmitApplication(ctx context.Context, id uuid.UUID, req model.ApplicationRequest) (model.Application, error)
	ReviewApplication(ctx context.Context, tournamentID, applicationID uuid.UUID, req model.ReviewApplicationRequest) (model.Application, error)
	WithdrawApplication(ctx context.Context, tournamentID, applicationID uuid.UUID, req model.VersionRequest) (model.Application, error)
}

// Competition задаёт сетку, расписание, результаты и разрешение споров.
type Competition interface {
	GetMatch(ctx context.Context, id uuid.UUID) (model.Match, error)
	ListMatches(ctx context.Context, id uuid.UUID, params model.ListMatchesParams) (model.MatchPage, error)
	GetBracket(ctx context.Context, id uuid.UUID) (model.Bracket, error)
	GetStandings(ctx context.Context, id uuid.UUID) ([]model.Standings, error)
	PreviewDraw(ctx context.Context, id uuid.UUID, req model.SeedingRequest) (model.DrawPreview, error)
	ConfirmDraw(ctx context.Context, id uuid.UUID, req model.ConfirmDrawRequest) (model.Bracket, error)
	ScheduleMatch(ctx context.Context, id uuid.UUID, req model.ScheduleMatchRequest) (model.Match, error)
	ScheduleRound(ctx context.Context, id uuid.UUID, req model.ScheduleRoundRequest) ([]model.Match, error)
	AssignMatchReferee(ctx context.Context, id uuid.UUID, req model.AssignRefereeRequest) (model.Match, error)
	ListResultReports(ctx context.Context, id uuid.UUID) ([]model.ResultReport, error)
	ReportResult(ctx context.Context, id uuid.UUID, req model.ReportResultRequest) (model.ResultReport, error)
	ConfirmResult(ctx context.Context, id, reportID uuid.UUID, req model.VersionRequest) (model.Match, error)
	SetRefereeResult(ctx context.Context, id uuid.UUID, req model.ReportResultRequest) (model.Match, error)
	GetDispute(ctx context.Context, id uuid.UUID) (model.Dispute, error)
	DisputeResult(ctx context.Context, id, reportID uuid.UUID, req model.DisputeRequest) (model.Dispute, error)
	ResolveDispute(ctx context.Context, id uuid.UUID, req model.ResolveDisputeRequest) (model.Match, error)
	ListAssignedDisputes(ctx context.Context, params model.ListAssignedDisputesParams) (model.DisputePage, error)
}

// Admin задаёт управление ролями, модерацию и доступ к журналу действий.
type Admin interface {
	ListAdminTournaments(ctx context.Context, params model.ListAdminTournamentsParams) (model.TournamentPage, error)
	ListAdminUsers(ctx context.Context, params model.ListAdminUsersParams) (model.ProfilePage, error)
	BlockUser(ctx context.Context, id uuid.UUID, req model.BlockUserRequest) (model.Profile, error)
	UpdateUserRoles(ctx context.Context, id uuid.UUID, req model.UpdateRolesRequest) (model.Profile, error)
	ModerateTournament(ctx context.Context, id uuid.UUID, req model.ModerateTournamentRequest) (model.Tournament, error)
	DeleteTournament(ctx context.Context, id uuid.UUID, req model.ReasonRequest) error
	ListAudit(ctx context.Context, params model.ListAuditParams) (model.AuditEntryPage, error)
}

// Catalog задаёт справочник дисциплин, рейтинги и уведомления пользователя.
type Catalog interface {
	ListDisciplines(ctx context.Context, params model.ListDisciplinesParams) (model.DisciplinePage, error)
	CreateDiscipline(ctx context.Context, req model.DisciplineRequest) (model.Discipline, error)
	UpdateDiscipline(ctx context.Context, id uuid.UUID, req model.UpdateDisciplineRequest) (model.Discipline, error)
	ListRatings(ctx context.Context, params model.ListRatingsParams) (model.RatingPage, error)
	ListRatingHistory(ctx context.Context, id uuid.UUID, params model.ListRatingHistoryParams) (model.RatingChangePage, error)
	ListNotifications(ctx context.Context, params model.ListNotificationsParams) (model.NotificationPage, error)
	ReadNotification(ctx context.Context, id uuid.UUID) error
}

// Resources задаёт доступ к файлам, выгрузке и событиям турнира.
type Resources interface {
	UploadFile(ctx context.Context, upload model.FileUpload) (model.File, error)
	DownloadFile(ctx context.Context, id uuid.UUID) (model.FileContent, error)
	ExportTournament(ctx context.Context, id uuid.UUID) ([]byte, error)
	TournamentVisible(ctx context.Context, id uuid.UUID) error
	LatestEventID(ctx context.Context, id uuid.UUID) (int64, error)
	ReplayEvents(ctx context.Context, id uuid.UUID, after int64) (model.EventBatch, error)
}

// Backend используется при сборке зависимостей транспорта. Каждый обработчик
// хранит только интерфейс своей предметной области; фоновые задачи сюда не входят.
type Backend interface {
	Identity
	Teams
	Tournaments
	Competition
	Admin
	Catalog
	Resources
}

var _ Backend = (*Service)(nil)
