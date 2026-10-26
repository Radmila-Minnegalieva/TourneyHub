package api

import (
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/slice"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (a *API) ScheduleMatch(c echo.Context, id uuid.UUID) error {
	var req contract.ScheduleMatchRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.ScheduleMatch(c.Request().Context(), id, toModelScheduleMatchRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatch(result))
}

func (a *API) ScheduleRound(c echo.Context, id uuid.UUID) error {
	var req contract.ScheduleRoundRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.ScheduleRound(c.Request().Context(), id, toModelScheduleRoundRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, slice.Map(result, fromModelMatch))
}

func (a *API) AssignMatchReferee(c echo.Context, id uuid.UUID) error {
	var req contract.AssignRefereeRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.AssignMatchReferee(c.Request().Context(), id, toModelAssignRefereeRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatch(result))
}
