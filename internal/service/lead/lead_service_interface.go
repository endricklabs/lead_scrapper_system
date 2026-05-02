package lead_service

import (
	lead_dto "lead_scrapper_be/internal/dto/lead"

	"github.com/labstack/echo/v4"
)

type LeadService interface {
	Scrap(c echo.Context, leadScrapRequest lead_dto.LeadScrapRequest) error
}
