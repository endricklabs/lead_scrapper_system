package lead_service

import (
	lead_dto "lead_scrapper_be/internal/dto/lead"
	"lead_scrapper_be/setup"
	"sync"

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

	var wg sync.WaitGroup

	// enqueue the jobs
	for _, reqSource := range leadScrapRequest.Source {
		for i := range s.app.QueueList {
			if s.app.QueueList[i].Source == string(reqSource.Source) {
				wg.Add(1)
				s.app.QueueList[i].Enqueue(string(reqSource.Source), leadScrapRequest.IndustryType, leadScrapRequest.Location, int(reqSource.NumberOfRequest), &wg)
			}
		}
	}

	// block until all scraping jobs complete
	wg.Wait()

	return nil
}
