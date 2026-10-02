package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	contract "github.com/Radmila-Minnegalieva/TourneyHub/pkg/api/v1"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"
)

const BaseURL = "/api/v1"

type Database interface{ Ping(context.Context) error }

func NewServer(db Database, corsOrigins ...string) (*echo.Echo, error) {
	spec, err := contract.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("load contract: %w", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validate contract: %w", err)
	}
	// The validator receives the full HTTP path; generator registers on a group.
	spec.Servers = nil
	paths := openapi3.NewPaths()
	for path, item := range spec.Paths.Map() {
		paths.Set(BaseURL+path, item)
	}
	spec.Paths = paths
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.HTTPErrorHandler = handleError
	server.Use(middleware.RequestID(), middleware.Recover(), middleware.BodyLimit("6M"))
	if len(corsOrigins) > 0 {
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: corsOrigins, AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID, "Last-Event-ID"}, ExposeHeaders: []string{echo.HeaderXRequestID, "Content-Disposition"}}))
	}
	server.GET("/health/live", func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	server.GET("/health/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "database unavailable").SetInternal(err)
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	server.GET("/openapi.yaml", func(c echo.Context) error { return c.Blob(http.StatusOK, "application/yaml", contract.Specification) })
	group := server.Group(BaseURL)
	group.Use(oapimiddleware.OapiRequestValidatorWithOptions(spec, &oapimiddleware.Options{
		Options: openapi3filter.Options{AuthenticationFunc: func(context.Context, *openapi3filter.AuthenticationInput) error {
			// Fail closed until authentication is implemented; never accept arbitrary tokens.
			return echo.NewHTTPError(http.StatusNotImplemented, "authentication is not implemented")
		}},
	}))
	contract.RegisterHandlers(group, New())
	return server, nil
}

func handleError(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	message := "internal server error"
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		status = httpError.Code
		if status < 500 || status == http.StatusNotImplemented {
			message = fmt.Sprint(httpError.Message)
		}
	}
	code := map[int]string{400: "bad_request", 401: "unauthorized", 403: "forbidden", 404: "not_found", 405: "method_not_allowed", 409: "conflict", 413: "payload_too_large", 415: "unsupported_media_type", 422: "validation_failed", 429: "rate_limited", 501: "not_implemented", 503: "unavailable"}[status]
	if code == "" {
		code = "internal_error"
	}
	if status >= 500 && status != http.StatusNotImplemented {
		slog.Error("http request failed", "status", status, "error", err)
	}
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	if err := c.JSON(status, contract.Error{Code: code, Message: message, RequestId: requestID}); err != nil {
		slog.Error("write error response", "error", err)
	}
}
