package route

import (
	"lead_scrapper_be/setup"

	authHandler "lead_scrapper_be/internal/handler/auth"
	leadHandler "lead_scrapper_be/internal/handler/lead"
	jwt_middleware "lead_scrapper_be/internal/middleware/jwt"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, app *setup.Application) {

	api := e.Group("/api/v1")

	// Auth Route
	authApi := api.Group("/auth")
	authRoutes(authApi, app)

	// Lead Scrapper Route
	leadApi := api.Group("/lead")
	leadRoutes(leadApi, app)

}

func authRoutes(api *echo.Group, app *setup.Application) {
	authHandler := authHandler.NewAuthHandler(app)

	api.POST("/signup", authHandler.Signup)
	api.POST("/login", authHandler.Login)
}

func leadRoutes(api *echo.Group, app *setup.Application) {
	leadHandler := leadHandler.NewLeadHandler(app)

	api.Use(jwt_middleware.JWTMiddleware(app))
	api.POST("/scrap", leadHandler.Scrap)
}
