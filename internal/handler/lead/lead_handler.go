package lead

import (
	lead_dto "lead_scrapper_be/internal/dto/lead"
	"lead_scrapper_be/internal/model"
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

// Scrap
// @Summary Scrap leads from multiple sources
// @Description Start a lead scrapping job for the specified industry and location across multiple sources (Google Maps, LinkedIn, etc.)
// @Tags Lead
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body lead_dto.LeadScrapRequest true "Scrap Request"
// @Success 200 {object} api_dto.ApiResponse
// @Failure 400 {object} api_dto.ApiResponse
// @Failure 500 {object} api_dto.ApiResponse
// @Router /lead/scrap [post]
func (h *LeadHandler) Scrap(c echo.Context) error {
	var leadScrapRequest lead_dto.LeadScrapRequest
	if err := c.Bind(&leadScrapRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	user := c.Get("user").(*model.User)

	// You might want to add validation here
	if err := c.Validate(&leadScrapRequest); err != nil {
		return api.BadRequest(c, err.Error())
	}

	if err := h.leadService.Scrap(c, leadScrapRequest, *user); err != nil {
		return api.InternalServerError(c, err.Error())
	}

	return api.SuccessfulResponse(c, "Lead Scrapper triggered successfully")
}
