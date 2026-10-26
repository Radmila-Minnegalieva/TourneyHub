package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *API) ListAdminTournaments(c echo.Context, p contract.ListAdminTournamentsParams) error {
	result, err := a.admin.ListAdminTournaments(c.Request().Context(), toModelListAdminTournamentsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournamentPage(result))
}

func (a *API) ListAdminUsers(c echo.Context, p contract.ListAdminUsersParams) error {
	result, err := a.admin.ListAdminUsers(c.Request().Context(), toModelListAdminUsersParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelProfilePage(result))
}

func (a *API) BlockUser(c echo.Context, id uuid.UUID) error {
	var request contract.BlockUserRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.admin.BlockUser(c.Request().Context(), id, toModelBlockUserRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelProfile(result))
}

func (a *API) UpdateUserRoles(c echo.Context, id uuid.UUID) error {
	var request contract.UpdateRolesRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.admin.UpdateUserRoles(c.Request().Context(), id, toModelUpdateRolesRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelProfile(result))
}

func (a *API) ModerateTournament(c echo.Context, id uuid.UUID) error {
	var request contract.ModerateTournamentRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.admin.ModerateTournament(c.Request().Context(), id, toModelModerateTournamentRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}

func (a *API) DeleteTournament(c echo.Context, id uuid.UUID) error {
	var request contract.ReasonRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	if err := a.admin.DeleteTournament(c.Request().Context(), id, toModelReasonRequest(request)); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (a *API) ListAudit(c echo.Context, p contract.ListAuditParams) error {
	result, err := a.admin.ListAudit(c.Request().Context(), toModelListAuditParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelAuditEntryPage(result))
}
