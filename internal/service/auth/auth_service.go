package auth_service

import (
	"time"

	auth_dto "lead_scrapper_be/internal/dto/auth"
	"lead_scrapper_be/internal/model"
	"lead_scrapper_be/pkg/utils"
	"lead_scrapper_be/pkg/utils/api"
	auth_utils "lead_scrapper_be/pkg/utils/auth"
	"lead_scrapper_be/setup"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type authService struct {
	app *setup.Application
}

func NewAuthService(app *setup.Application) AuthService {
	return authService{
		app: app,
	}
}

func (s authService) Signup(signupRequest auth_dto.SignupRequest) error {

	hashedPassword, err := utils.HashPassword(signupRequest.Password)
	if err != nil {
		s.app.Logger.Error("Error while hashing password", "error", err)
		return err
	}

	databaseInput := model.User{
		Email:    signupRequest.Email,
		Password: hashedPassword,
	}

	err = s.app.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&databaseInput).Error; err != nil {
			s.app.Logger.Error("Error while creating user in database", "error", err)
			return err
		}

		var basicPackage model.SubscriptionPackage
		if err := tx.Where("slug = ?", "basic").First(&basicPackage).Error; err != nil {
			s.app.Logger.Error("Error while fetching basic subscription package", "error", err)
			return err
		}

		start := time.Now()
		userSubscription := model.UserSubscription{
			UserID:                databaseInput.ID,
			SubscriptionPackageID: basicPackage.ID,
			Status:                model.UserSubscriptionStatusActive,
			StartDate:             &start,
			EndDate:               nil,
		}

		if err := tx.Create(&userSubscription).Error; err != nil {
			s.app.Logger.Error("Error while creating user subscription", "error", err)
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.app.Logger.Info("User created successfully with basic subscription", "email", signupRequest.Email)
	return nil
}

func (s authService) Login(c echo.Context, loginRequest auth_dto.LoginRequest) (*auth_dto.LoginResponse, error) {

	var user model.User
	if err := s.app.DB.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		s.app.Logger.Error("Error while fetching user from database", "error", err)
		return nil, err
	}

	if !utils.CheckPasswordHash(loginRequest.Password, user.Password) {
		s.app.Logger.Error("Invalid password")
		return nil, api.NewError("Invalid email or password")
	}

	// Increment token version to invalidate old tokens
	user.TokenVersion++
	if err := s.app.DB.Model(&user).Update("token_version", user.TokenVersion).Error; err != nil {
		s.app.Logger.Error("Error while updating token version", "error", err)
		return nil, err
	}

	access_token, err := auth_utils.GenerateToken(user.ID.String(), user.Email, user.TokenVersion, []byte(s.app.Config.JWT.Secret))
	if err != nil {
		s.app.Logger.Error("Error while generating access token", "error", err)
		return nil, err
	}

	s.app.Logger.Info("User logged in successfully", "email", loginRequest.Email)
	return &auth_dto.LoginResponse{
		AccessToken: access_token,
	}, nil
}
