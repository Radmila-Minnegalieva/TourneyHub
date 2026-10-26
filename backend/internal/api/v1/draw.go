package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (a *API) GetBracket(c echo.Context, id uuid.UUID) error {
	result, err := a.competition.GetBracket(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelBracket(result))
}

func (a *API) PreviewDraw(c echo.Context, id uuid.UUID) error {
	var req contract.SeedingRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.PreviewDraw(c.Request().Context(), id, toModelSeedingRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(201, fromModelDrawPreview(result))
}

func (a *API) ConfirmDraw(c echo.Context, id uuid.UUID) error {
	var req contract.ConfirmDrawRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.ConfirmDraw(c.Request().Context(), id, toModelConfirmDrawRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelBracket(result))
}
