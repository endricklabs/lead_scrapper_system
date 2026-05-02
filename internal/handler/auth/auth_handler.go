package auth

import (
	auth_dto "lead_scrapper_be/internal/dto/auth"
	auth_service "lead_scrapper_be/internal/service/auth"
	"lead_scrapper_be/pkg/utils/api"
	"lead_scrapper_be/setup"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService auth_service.AuthService
}

func NewAuthHandler(app *setup.Application) *AuthHandler {
	return &AuthHandler{
		authService: auth_service.NewAuthService(app),
	}
}

func (h *AuthHandler) Signup(c echo.Context) error {
	var signupRequest auth_dto.SignupRequest
	if err := c.Bind(&signupRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	// You might want to add validation here
	if err := c.Validate(&signupRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	if err := h.authService.Signup(signupRequest); err != nil {
		return api.InternalServerError(c, err.Error())
	}

	return api.SuccessfulResponse(c, "User registered successfully")
}

func (h *AuthHandler) Login(c echo.Context) error {
	var loginRequest auth_dto.LoginRequest
	if err := c.Bind(&loginRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	if err := c.Validate(&loginRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	loginResponse, err := h.authService.Login(c, loginRequest)
	if err != nil {
		return api.Unauthorized(c, "Invalid email or password")
	}

	return api.SuccessfulResponseWithData(c, "Login successful", loginResponse)
}
