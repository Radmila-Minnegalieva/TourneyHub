package api

import (
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/slice"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (a *API) ListResultReports(c echo.Context, id uuid.UUID) error {
	result, err := a.competition.ListResultReports(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(200, slice.Map(result, fromModelResultReport))
}

func (a *API) ReportResult(c echo.Context, id uuid.UUID) error {
	var req contract.ReportResultRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.ReportResult(c.Request().Context(), id, toModelReportResultRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(201, fromModelResultReport(result))
}

func (a *API) ConfirmResult(c echo.Context, id, reportID uuid.UUID) error {
	var req contract.VersionRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.ConfirmResult(c.Request().Context(), id, reportID, toModelVersionRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatch(result))
}

func (a *API) SetRefereeResult(c echo.Context, id uuid.UUID) error {
	var req contract.ReportResultRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.competition.SetRefereeResult(c.Request().Context(), id, toModelReportResultRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelMatch(result))
}
