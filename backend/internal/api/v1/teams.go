package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *API) ListTeams(c echo.Context, params contract.ListTeamsParams) error {
	result, err := a.teams.ListTeams(c.Request().Context(), toModelListTeamsParams(params))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTeamPage(result))
}

func (a *API) ListMyTeams(c echo.Context, params contract.ListMyTeamsParams) error {
	result, err := a.teams.ListMyTeams(c.Request().Context(), toModelListMyTeamsParams(params))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTeamPage(result))
}

func (a *API) ListMyInvitations(c echo.Context, params contract.ListMyInvitationsParams) error {
	result, err := a.teams.ListMyInvitations(c.Request().Context(), toModelListMyInvitationsParams(params))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelInvitationPage(result))
}

func (a *API) GetTeam(c echo.Context, id uuid.UUID) error {
	result, err := a.teams.GetTeam(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTeam(result))
}

func (a *API) ListTeamMembers(c echo.Context, id uuid.UUID) error {
	result, err := a.teams.ListTeamMembers(c.Request().Context(), id)
	if err != nil {
		return err
	}
	items := make([]contract.TeamMember, len(result))
	for i, item := range result {
		items[i] = fromModelTeamMember(item)
	}
	return c.JSON(http.StatusOK, items)
}

func (a *API) CreateTeam(c echo.Context) error {
	var req contract.CreateTeamRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.teams.CreateTeam(c.Request().Context(), toModelCreateTeamRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(201, fromModelTeam(result))
}

func (a *API) UpdateTeam(c echo.Context, teamID uuid.UUID) error {
	var req contract.UpdateTeamRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.teams.UpdateTeam(c.Request().Context(), teamID, toModelUpdateTeamRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelTeam(result))
}

func (a *API) DeleteTeam(c echo.Context, teamID uuid.UUID) error {
	var req contract.VersionRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.teams.DeleteTeam(c.Request().Context(), teamID, toModelVersionRequest(req)); err != nil {
		return err
	}
	return c.NoContent(204)
}

func (a *API) RemoveTeamMember(c echo.Context, teamID uuid.UUID, userID uuid.UUID) error {
	var req contract.VersionRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.teams.RemoveTeamMember(c.Request().Context(), teamID, userID, toModelVersionRequest(req)); err != nil {
		return err
	}
	return c.NoContent(204)
}

func (a *API) TransferCaptain(c echo.Context, teamID uuid.UUID) error {
	var req contract.TransferCaptainRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.teams.TransferCaptain(c.Request().Context(), teamID, toModelTransferCaptainRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelTeam(result))
}

func (a *API) CreateInvitation(c echo.Context, teamID uuid.UUID) error {
	var req contract.CreateInvitationRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.teams.CreateInvitation(c.Request().Context(), teamID, toModelCreateInvitationRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(201, fromModelInvitationLink(result))
}

func (a *API) DecideInvitation(c echo.Context, invitationID uuid.UUID) error {
	var req contract.InvitationDecisionRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.teams.DecideInvitation(c.Request().Context(), invitationID, toModelInvitationDecisionRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelInvitation(result))
}

func (a *API) RevokeInvitation(c echo.Context, teamID, invitationID uuid.UUID) error {
	if err := a.teams.RevokeInvitation(c.Request().Context(), teamID, invitationID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
