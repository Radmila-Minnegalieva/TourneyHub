package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (a *API) GetMatch(c echo.Context, id uuid.UUID) error {
	result, err := a.competition.GetMatch(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatch(result))
}

func (a *API) ListMatches(c echo.Context, id uuid.UUID, params contract.ListMatchesParams) error {
	result, err := a.competition.ListMatches(c.Request().Context(), id, toModelListMatchesParams(params))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatchPage(result))
}
