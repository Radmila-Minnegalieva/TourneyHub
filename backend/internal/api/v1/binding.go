package api

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"io"
)

func bindJSON(c echo.Context, dst any) error {
	decoder := json.NewDecoder(c.Request().Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return echo.NewHTTPError(400, "invalid JSON body").SetInternal(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return echo.NewHTTPError(400, "expected one JSON object")
	}
	return nil
}
