package main

import (
	"net/http"
	"strings"

	"lead_scrapper_be/internal/route"
	"lead_scrapper_be/setup"

	"lead_scrapper_be/pkg/validator"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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
