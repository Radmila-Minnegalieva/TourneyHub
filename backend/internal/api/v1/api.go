// Package api implements the HTTP handlers for TourneyHub v1.
package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"net/http"
)

// API is the contract-first handler scaffold. Add service dependencies here as features are implemented.
type API struct{}

var _ contract.ServerInterface = (*API)(nil)

func New() *API { return &API{} }

func notImplemented(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "operation is not implemented")
}

func (a *API) ListAudit(ctx echo.Context, params contract.ListAuditParams) error {
	return notImplemented(ctx)
}

func (a *API) CreateDiscipline(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) UpdateDiscipline(ctx echo.Context, disciplineId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListAdminTournaments(ctx echo.Context, params contract.ListAdminTournamentsParams) error {
	return notImplemented(ctx)
}

func (a *API) DeleteTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ModerateTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListAdminUsers(ctx echo.Context, params contract.ListAdminUsersParams) error {
	return notImplemented(ctx)
}

func (a *API) BlockUser(ctx echo.Context, userId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) UpdateUserRoles(ctx echo.Context, userId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) Login(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) Logout(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) RequestPasswordReset(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) ResetPassword(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) RefreshTokens(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) Register(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) VerifyEmail(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) ListDisciplines(ctx echo.Context, params contract.ListDisciplinesParams) error {
	return notImplemented(ctx)
}

func (a *API) ListAssignedDisputes(ctx echo.Context, params contract.ListAssignedDisputesParams) error {
	return notImplemented(ctx)
}

func (a *API) UploadFile(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) DownloadFile(ctx echo.Context, fileId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) DecideInvitation(ctx echo.Context, invitationId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) GetMatch(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) GetDispute(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ResolveDispute(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) AssignMatchReferee(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) SetRefereeResult(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListResultReports(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ReportResult(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ConfirmResult(ctx echo.Context, matchId openapi_types.UUID, reportId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) DisputeResult(ctx echo.Context, matchId openapi_types.UUID, reportId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ScheduleMatch(ctx echo.Context, matchId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListRatings(ctx echo.Context, params contract.ListRatingsParams) error {
	return notImplemented(ctx)
}

func (a *API) ListTeams(ctx echo.Context, params contract.ListTeamsParams) error {
	return notImplemented(ctx)
}

func (a *API) CreateTeam(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) DeleteTeam(ctx echo.Context, teamId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) GetTeam(ctx echo.Context, teamId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) UpdateTeam(ctx echo.Context, teamId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) TransferCaptain(ctx echo.Context, teamId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) CreateInvitation(ctx echo.Context, teamId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) RevokeInvitation(ctx echo.Context, teamId openapi_types.UUID, invitationId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListTeamMembers(ctx echo.Context, teamId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) RemoveTeamMember(ctx echo.Context, teamId openapi_types.UUID, userId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListRatingHistory(ctx echo.Context, teamId openapi_types.UUID, params contract.ListRatingHistoryParams) error {
	return notImplemented(ctx)
}

func (a *API) ListTournaments(ctx echo.Context, params contract.ListTournamentsParams) error {
	return notImplemented(ctx)
}

func (a *API) CreateTournament(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) GetTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) UpdateTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListApplications(ctx echo.Context, tournamentId openapi_types.UUID, params contract.ListApplicationsParams) error {
	return notImplemented(ctx)
}

func (a *API) SubmitApplication(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ReviewApplication(ctx echo.Context, tournamentId openapi_types.UUID, applicationId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) WithdrawApplication(ctx echo.Context, tournamentId openapi_types.UUID, applicationId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) GetBracket(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) CancelTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) CompleteTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ConfirmDraw(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) PreviewDraw(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) StreamTournamentEvents(ctx echo.Context, tournamentId openapi_types.UUID, params contract.StreamTournamentEventsParams) error {
	return notImplemented(ctx)
}

func (a *API) ExportTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListMatches(ctx echo.Context, tournamentId openapi_types.UUID, params contract.ListMatchesParams) error {
	return notImplemented(ctx)
}

func (a *API) ListParticipants(ctx echo.Context, tournamentId openapi_types.UUID, params contract.ListParticipantsParams) error {
	return notImplemented(ctx)
}

func (a *API) PublishTournament(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) AssignTournamentReferee(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ScheduleRound(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) GetStandings(ctx echo.Context, tournamentId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) SearchUsers(ctx echo.Context, params contract.SearchUsersParams) error {
	return notImplemented(ctx)
}

func (a *API) GetProfile(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) UpdateProfile(ctx echo.Context) error {
	return notImplemented(ctx)
}

func (a *API) ListMyApplications(ctx echo.Context, params contract.ListMyApplicationsParams) error {
	return notImplemented(ctx)
}

func (a *API) ListMyInvitations(ctx echo.Context, params contract.ListMyInvitationsParams) error {
	return notImplemented(ctx)
}

func (a *API) ListNotifications(ctx echo.Context, params contract.ListNotificationsParams) error {
	return notImplemented(ctx)
}

func (a *API) ReadNotification(ctx echo.Context, notificationId openapi_types.UUID) error {
	return notImplemented(ctx)
}

func (a *API) ListMyTeams(ctx echo.Context, params contract.ListMyTeamsParams) error {
	return notImplemented(ctx)
}

func (a *API) ListMyTournaments(ctx echo.Context, params contract.ListMyTournamentsParams) error {
	return notImplemented(ctx)
}
