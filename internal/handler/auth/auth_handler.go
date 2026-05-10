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

// Signup
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth_dto.SignupRequest true "Signup Request"
// @Success 200 {object} api_dto.ApiResponse
// @Failure 400 {object} api_dto.ApiResponse
// @Failure 500 {object} api_dto.ApiResponse
// @Router /auth/signup [post]
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

// Login
// @Summary User login
// @Description Authenticate user and return access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth_dto.LoginRequest true "Login Request"
// @Success 200 {object} api_dto.ApiResponse{data=auth_dto.LoginResponse}
// @Failure 400 {object} api_dto.ApiResponse
// @Failure 401 {object} api_dto.ApiResponse
// @Router /auth/login [post]
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
