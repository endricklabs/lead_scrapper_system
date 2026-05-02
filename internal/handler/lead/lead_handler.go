package lead

import (
	lead_dto "lead_scrapper_be/internal/dto/lead"
	lead_service "lead_scrapper_be/internal/service/lead"

	"lead_scrapper_be/pkg/utils/api"
	"lead_scrapper_be/setup"

	"github.com/labstack/echo/v4"
)

type LeadHandler struct {
	leadService lead_service.LeadService
}

func NewLeadHandler(app *setup.Application) *LeadHandler {
	return &LeadHandler{
		leadService: lead_service.NewLeadService(app),
	}
}

func (h *LeadHandler) Scrap(c echo.Context) error {
	var leadScrapRequest lead_dto.LeadScrapRequest
	if err := c.Bind(&leadScrapRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	// You might want to add validation here
	if err := c.Validate(&leadScrapRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	if err := h.leadService.Scrap(c, leadScrapRequest); err != nil {
		return api.InternalServerError(c, err.Error())
	}

	return api.SuccessfulResponse(c, "Scrapped successfully")
}
