package api

import (
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (a *API) Register(c echo.Context) error {
	var req contract.RegisterRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.identity.Register(c.Request().Context(), toModelRegisterRequest(req)); err != nil {
		return err
	}
	return c.NoContent(202)
}

func (a *API) VerifyEmail(c echo.Context) error {
	var req contract.TokenRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.identity.VerifyEmail(c.Request().Context(), toModelTokenRequest(req)); err != nil {
		return err
	}
	return c.NoContent(204)
}

func (a *API) Login(c echo.Context) error {
	var req contract.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.identity.Login(c.Request().Context(), toModelLoginRequest(req), c.RealIP())
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelTokens(result))
}

func (a *API) RefreshTokens(c echo.Context) error {
	var req contract.TokenRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.identity.RefreshTokens(c.Request().Context(), toModelTokenRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelTokens(result))
}

func (a *API) Logout(c echo.Context) error {
	var req contract.TokenRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.identity.Logout(c.Request().Context(), toModelTokenRequest(req)); err != nil {
		return err
	}
	return c.NoContent(204)
}

func (a *API) RequestPasswordReset(c echo.Context) error {
	var req contract.EmailRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.identity.RequestPasswordReset(c.Request().Context(), toModelEmailRequest(req)); err != nil {
		return err
	}
	return c.NoContent(202)
}

func (a *API) ResetPassword(c echo.Context) error {
	var req contract.ResetPasswordRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	if err := a.identity.ResetPassword(c.Request().Context(), toModelResetPasswordRequest(req)); err != nil {
		return err
	}
	return c.NoContent(204)
}

func (a *API) UpdateProfile(c echo.Context) error {
	var req contract.UpdateProfileRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	result, err := a.identity.UpdateProfile(c.Request().Context(), toModelUpdateProfileRequest(req))
	if err != nil {
		return err
	}
	return c.JSON(200, fromModelProfile(result))
}

func (a *API) GetProfile(c echo.Context) error {
	result, err := a.identity.GetProfile(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelProfile(result))
}

func (a *API) SearchUsers(c echo.Context, params contract.SearchUsersParams) error {
	result, err := a.identity.SearchUsers(c.Request().Context(), toModelSearchUsersParams(params))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fromModelPublicUserPage(result))
}
