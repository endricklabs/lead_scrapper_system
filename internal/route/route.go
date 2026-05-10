package route

import (
	"lead_scrapper_be/setup"

	authHandler "lead_scrapper_be/internal/handler/auth"
	leadHandler "lead_scrapper_be/internal/handler/lead"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo, app *setup.Application) {

	api := e.Group("/api/v1")

	// Auth Route
	authApi := api.Group("/auth")
	authRoutes(authApi, app)

	// Lead Scrapper Route
	leadApi := api.Group("/lead")
	leadRoutes(leadApi, app)

	// Swagger Route
	e.GET("/swagger/*", echoSwagger.WrapHandler)

}

func authRoutes(api *echo.Group, app *setup.Application) {
	authHandler := authHandler.NewAuthHandler(app)

	api.POST("/signup", authHandler.Signup)
	api.POST("/login", authHandler.Login)
}

func leadRoutes(api *echo.Group, app *setup.Application) {
	leadHandler := leadHandler.NewLeadHandler(app)

	api.POST("/scrap", leadHandler.Scrap)
}
