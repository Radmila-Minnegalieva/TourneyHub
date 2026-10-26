package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *API) ListTournaments(c echo.Context, p contract.ListTournamentsParams) error {
	result, err := a.tournaments.ListTournaments(c.Request().Context(), toModelListTournamentsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournamentPage(result))
}

func (a *API) ListMyTournaments(c echo.Context, p contract.ListMyTournamentsParams) error {
	result, err := a.tournaments.ListMyTournaments(c.Request().Context(), toModelListMyTournamentsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournamentPage(result))
}

func (a *API) GetTournament(c echo.Context, id uuid.UUID) error {
	result, err := a.tournaments.GetTournament(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}

func (a *API) CreateTournament(c echo.Context) error {
	var request contract.CreateTournamentRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.CreateTournament(c.Request().Context(), toModelCreateTournamentRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, fromModelTournament(result))
}

func (a *API) UpdateTournament(c echo.Context, id uuid.UUID) error {
	var request contract.UpdateTournamentRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.UpdateTournament(c.Request().Context(), id, toModelUpdateTournamentRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}

func (a *API) PublishTournament(c echo.Context, id uuid.UUID) error {
	var request contract.VersionRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.PublishTournament(c.Request().Context(), id, toModelVersionRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}

func (a *API) CancelTournament(c echo.Context, id uuid.UUID) error {
	var request contract.ReasonRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.CancelTournament(c.Request().Context(), id, toModelReasonRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}

func (a *API) CompleteTournament(c echo.Context, id uuid.UUID) error {
	var request contract.VersionRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.CompleteTournament(c.Request().Context(), id, toModelVersionRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}

func (a *API) AssignTournamentReferee(c echo.Context, id uuid.UUID) error {
	var request contract.AssignRefereeRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.AssignTournamentReferee(c.Request().Context(), id, toModelAssignRefereeRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelTournament(result))
}
