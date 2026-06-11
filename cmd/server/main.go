package main

import (
	"net/http"
	"strings"

	"lead_scrapper_be/internal/route"
	"lead_scrapper_be/setup"

	"lead_scrapper_be/pkg/validator"

	_ "lead_scrapper_be/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// @title Lead Scrapper API
// @version 1.0
// @description This is a lead scrapping system API.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

func main() {
	e := echo.New()

	app := setup.NewApplication()

	if app.Config.FrontendURL != "" {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
		}))
	}

	e.Use(middleware.Secure())
	e.Use(middleware.Recover())

	e.Validator = validator.NewCustomValidator()
	route.SetupRoutes(e, app)

	addr := app.Config.ServerPort
	if !strings.HasPrefix(addr, ":") && !strings.Contains(addr, ":") {
		addr = ":" + addr
	}

	err := e.Start(addr)
	if err != nil {
		panic(err)
	}
}
