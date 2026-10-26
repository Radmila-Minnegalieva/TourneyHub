package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *API) ListApplications(c echo.Context, id uuid.UUID, p contract.ListApplicationsParams) error {
	result, err := a.tournaments.ListApplications(c.Request().Context(), id, toModelListApplicationsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelApplicationPage(result))
}

func (a *API) ListParticipants(c echo.Context, id uuid.UUID, p contract.ListParticipantsParams) error {
	result, err := a.tournaments.ListParticipants(c.Request().Context(), id, toModelListParticipantsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelApplicationPage(result))
}

func (a *API) ListMyApplications(c echo.Context, p contract.ListMyApplicationsParams) error {
	result, err := a.tournaments.ListMyApplications(c.Request().Context(), toModelListMyApplicationsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelApplicationPage(result))
}

func (a *API) SubmitApplication(c echo.Context, id uuid.UUID) error {
	var request contract.ApplicationRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.SubmitApplication(c.Request().Context(), id, toModelApplicationRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, fromModelApplication(result))
}

func (a *API) ReviewApplication(c echo.Context, tournamentID, applicationID uuid.UUID) error {
	var request contract.ReviewApplicationRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.ReviewApplication(c.Request().Context(), tournamentID, applicationID, toModelReviewApplicationRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelApplication(result))
}

func (a *API) WithdrawApplication(c echo.Context, tournamentID, applicationID uuid.UUID) error {
	var request contract.VersionRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.tournaments.WithdrawApplication(c.Request().Context(), tournamentID, applicationID, toModelVersionRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelApplication(result))
}
