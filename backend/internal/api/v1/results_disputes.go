package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (a *API) GetDispute(c echo.Context, id uuid.UUID) error {
	result, err := a.competition.GetDispute(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelDispute(result))
}

func (a *API) DisputeResult(c echo.Context, id, reportID uuid.UUID) error {
	var req contract.DisputeRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.DisputeResult(c.Request().Context(), id, reportID, toModelDisputeRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(201, fromModelDispute(result))
}

func (a *API) ResolveDispute(c echo.Context, id uuid.UUID) error {
	var req contract.ResolveDisputeRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.ResolveDispute(c.Request().Context(), id, toModelResolveDisputeRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatch(result))
}

func (a *API) ListAssignedDisputes(c echo.Context, params contract.ListAssignedDisputesParams) error {
	result, err := a.competition.ListAssignedDisputes(c.Request().Context(), toModelListAssignedDisputesParams(params))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelDisputePage(result))
}
