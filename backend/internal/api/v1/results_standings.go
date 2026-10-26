package api

import (
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/slice"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (a *API) GetStandings(c echo.Context, id uuid.UUID) error {
	result, err := a.competition.GetStandings(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(200, slice.Map(result, fromModelStandings))
}
