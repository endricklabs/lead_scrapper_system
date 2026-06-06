package jwt_middleware

import (
	"net/http"
	"strings"

	"lead_scrapper_be/internal/model"
	"lead_scrapper_be/pkg/utils/api"
	"lead_scrapper_be/setup"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(app *setup.Application) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// 1. Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, api.NewError("Missing authorization header"))
			}

			// 2. Expect format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, api.NewError("Invalid authorization format"))
			}

			tokenString := parts[1]

			// 3. Parse and validate token
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Ensure signing method is correct
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, api.NewError("Invalid signing method")
				}
				return []byte(app.Config.JWT.Secret), nil
			})

			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, api.NewError("Invalid or expired token"))
			}

			// 4. Extract claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, api.NewError("Invalid token claims"))
			}

			// 5. Verify Token Version
			userID := claims["sub"].(string)
			tokenVersion := int(claims["version"].(float64)) // JWT numbers are float64

			var user model.User
			if err := app.DB.First(&user, "id = ?", userID).Error; err != nil {
				return c.JSON(http.StatusUnauthorized, api.NewError("User not found"))
			}

			if user.TokenVersion != tokenVersion {
				return c.JSON(http.StatusUnauthorized, api.NewError("Token has been invalidated"))
			}

			// 6. Store user info in context (for use in handlers)
			c.Set("user", &user)
			c.Set("user_id", claims["sub"])
			c.Set("email", claims["email"])

			// 7. Continue request
			return next(c)
		}
	}
}
