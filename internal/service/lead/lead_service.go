package lead_service

import (
	lead_dto "lead_scrapper_be/internal/dto/lead"
	"lead_scrapper_be/pkg/queue"
	"lead_scrapper_be/pkg/worker"
	"lead_scrapper_be/setup"

	"github.com/labstack/echo/v4"
)

type leadService struct {
	app      *setup.Application
	jobQueue *queue.JobQueue
	worker   *worker.Worker
}

func NewLeadService(app *setup.Application) LeadService {
	return leadService{
		app: app,
	}
}

func (s leadService) Scrap(c echo.Context, leadScrapRequest lead_dto.LeadScrapRequest) error {

	//Prepare the job queues and worker
	s.jobQueue = queue.NewJobQueue(s.app.Config.LengthOfJobQueue)
	s.worker = worker.NewWorker()

	//set the sources
	sources := []string{"google_maps", "linked_in", "facebook", "instagram"}

	// enqueue the jobs
	for i := 1; i <= int(len(sources)); i++ {
		s.jobQueue.Enqueue(int64(i), sources[i-1], leadScrapRequest.IndustryType, leadScrapRequest.Location)
	}

	//Start the worker
	s.worker.Run(*s.jobQueue, int(s.app.Config.NumberOfWorkersPerRequest))

	//Enqueue the job

	return nil
}
