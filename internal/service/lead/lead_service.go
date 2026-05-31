package lead_service

import (
	lead_dto "lead_scrapper_be/internal/dto/lead"
	"lead_scrapper_be/internal/model"
	"lead_scrapper_be/pkg/queue"
	"lead_scrapper_be/pkg/utils/api"
	"lead_scrapper_be/setup"

	"github.com/labstack/echo/v4"
)

type leadService struct {
	app *setup.Application
}

func NewLeadService(app *setup.Application) LeadService {
	return leadService{
		app: app,
	}
}

func (s leadService) Scrap(c echo.Context, leadScrapRequest lead_dto.LeadScrapRequest) error {

	for _, reqSource := range leadScrapRequest.Source {

		// Not found, insert new
		job := model.LeadScrapingJob{
			Source:          string(reqSource.Source),
			IndustryType:    leadScrapRequest.IndustryType,
			Location:        leadScrapRequest.Location,
			TargetRequested: int(reqSource.NumberOfRequest),
			Status:          "PENDING",
		}
		s.app.DB.Create(&job)

	}

	// Trigger immediate background polling dynamically
	go queue.PollPendingJobs(s.app.DB, s.app.QueueList, s.app.Config, s.app.Logger)

	return api.SuccessfulResponse(c, "Lead Scrapping has been started.")
}
