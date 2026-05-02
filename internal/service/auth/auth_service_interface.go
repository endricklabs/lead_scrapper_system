package auth_service

import (
	auth_dto "lead_scrapper_be/internal/dto/auth"

	"github.com/labstack/echo/v4"
)

type AuthService interface {
	Signup(signupRequest auth_dto.SignupRequest) error
	Login(c echo.Context, loginRequest auth_dto.LoginRequest) (*auth_dto.LoginResponse, error)
}
