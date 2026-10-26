package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *API) ListDisciplines(c echo.Context, p contract.ListDisciplinesParams) error {
	result, err := a.catalog.ListDisciplines(c.Request().Context(), toModelListDisciplinesParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelDisciplinePage(result))
}

func (a *API) CreateDiscipline(c echo.Context) error {
	var request contract.DisciplineRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.catalog.CreateDiscipline(c.Request().Context(), toModelDisciplineRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, fromModelDiscipline(result))
}

func (a *API) UpdateDiscipline(c echo.Context, id uuid.UUID) error {
	var request contract.UpdateDisciplineRequest
	if err := bindJSON(c, &request); err != nil {
		return err
	}
	result, err := a.catalog.UpdateDiscipline(c.Request().Context(), id, toModelUpdateDisciplineRequest(request))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelDiscipline(result))
}

func (a *API) ListRatings(c echo.Context, p contract.ListRatingsParams) error {
	result, err := a.catalog.ListRatings(c.Request().Context(), toModelListRatingsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelRatingPage(result))
}

func (a *API) ListRatingHistory(c echo.Context, id uuid.UUID, p contract.ListRatingHistoryParams) error {
	result, err := a.catalog.ListRatingHistory(c.Request().Context(), id, toModelListRatingHistoryParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelRatingChangePage(result))
}

func (a *API) ListNotifications(c echo.Context, p contract.ListNotificationsParams) error {
	result, err := a.catalog.ListNotifications(c.Request().Context(), toModelListNotificationsParams(p))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelNotificationPage(result))
}

func (a *API) ReadNotification(c echo.Context, id uuid.UUID) error {
	if err := a.catalog.ReadNotification(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
