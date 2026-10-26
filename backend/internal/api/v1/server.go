package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/authctx"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"
)

const BaseURL = "/api/v1"

func init() {
	openapi3.DefineStringFormatValidator("uuid", openapi3.NewCallbackValidator(uuid.Validate))
}

type Database interface {
	Ping(context.Context) error
}

func NewServer(db Database, backend service.Backend, cfg config.HTTP) (*echo.Echo, error) {
	spec, err := contract.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("load contract: %w", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validate contract: %w", err)
	}
	// Валидатор получает полный HTTP-путь; сгенерированные маршруты регистрируются в группе.
	spec.Servers = nil
	paths := openapi3.NewPaths()
	for path, item := range spec.Paths.Map() {
		paths.Set(BaseURL+path, item)
	}
	spec.Paths = paths
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.IPExtractor = echo.ExtractIPDirect()
	server.HTTPErrorHandler = handleError
	server.Use(middleware.RequestID(), middleware.Recover(), middleware.BodyLimit(cfg.BodyLimit))
	if len(cfg.CORSOrigins) > 0 {
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: cfg.CORSOrigins, AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID, "Last-Event-ID"}, ExposeHeaders: []string{echo.HeaderXRequestID, "Content-Disposition"}}))
	}
	server.GET("/health/live", func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]string{"status": "ok"}) })
	server.GET("/health/ready", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), cfg.ReadHeaderTimeout)
		defer cancel()
		if err := db.Ping(ctx); err != nil {
			return echo.NewHTTPError(http.StatusServiceUnavailable, "database unavailable").SetInternal(err)
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	server.GET("/openapi.yaml", func(c echo.Context) error { return c.Blob(http.StatusOK, "application/yaml", contract.Specification) })
	group := server.Group(BaseURL)
	api := New(backend, cfg)
	group.Use(api.authentication)
	group.Use(oapimiddleware.OapiRequestValidatorWithOptions(spec, &oapimiddleware.Options{
		Options: openapi3filter.Options{AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
			if authctx.FromContext(input.RequestValidationInput.Request.Context()).ID == uuid.Nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}
			return nil
		}},
	}))
	group.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !api.available {
				return echo.NewHTTPError(503, "database unavailable")
			}
			return next(c)
		}
	})
	contract.RegisterHandlers(group, api)
	return server, nil
}

func (a *API) authentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		header := c.Request().Header.Get(echo.HeaderAuthorization)
		if header != "" {
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !a.available {
				return echo.NewHTTPError(401, "invalid access token")
			}
			p, err := a.identity.Authenticate(c.Request().Context(), parts[1])
			if err != nil {
				return err
			}
			c.SetRequest(c.Request().WithContext(authctx.WithPrincipal(c.Request().Context(), p)))
		}
		return next(c)
	}
}

func handleError(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	message := "internal server error"
	var applicationError *apperror.Error
	if errors.As(err, &applicationError) {
		status = map[apperror.Kind]int{apperror.BadRequest: 400, apperror.Unauthorized: 401, apperror.Forbidden: 403, apperror.NotFound: 404, apperror.Conflict: 409, apperror.Validation: 422, apperror.RateLimited: 429, apperror.Unavailable: 503}[applicationError.Kind]
		if status == 0 {
			status = 500
		} else {
			message = applicationError.Message
		}
	}
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
