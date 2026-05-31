package queue

import (
	"fmt"
	"lead_scrapper_be/internal/config"
	"lead_scrapper_be/internal/model"
	"lead_scrapper_be/pkg/logger"
	"lead_scrapper_be/pkg/scrapper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Job struct {
	JobID           uuid.UUID
	Source          string
	IndustryType    string
	Location        string
	NumberOfRequest int
	LeadsCollected  int
}

type JobQueue struct {
	Id     int64
	Source string
	Jobs   chan Job
	DB     *gorm.DB
	Config *config.Config
	Logger logger.Logger
}

func NewJobQueue(size int64, id int64, db *gorm.DB, cfg *config.Config, log logger.Logger) *JobQueue {
	return &JobQueue{
		Id:     id,
		Jobs:   make(chan Job, size),
		DB:     db,
		Config: cfg,
		Logger: log,
	}
}

func (j *JobQueue) Enqueue(jobID uuid.UUID, source string, industryType string, location string, numberOfRequest int, leadsCollected int) {
	j.Jobs <- Job{
		JobID:           jobID,
		Source:          source,
		IndustryType:    industryType,
		Location:        location,
		NumberOfRequest: numberOfRequest,
		LeadsCollected:  leadsCollected,
	}
}

//

func worker(id int, jobs <-chan Job, db *gorm.DB, cfg *config.Config, log logger.Logger) {
	for {
		job, ok := <-jobs
		if !ok {
			log.Info("Returning from go routine because channel is closed")
			return
		}

		var err error
		switch job.Source {
		case "google_maps":
			err = scrapper.ScrapGoogleMaps(db, cfg, log, job.JobID, job.IndustryType, job.Location, job.NumberOfRequest)
		case "linked_in":
			err = scrapper.ScrapLinkedIn(db, cfg, log, job.JobID, job.IndustryType, job.Location, job.NumberOfRequest)
		case "facebook":
			err = scrapper.ScrapFacebook(db, cfg, log, job.JobID, job.IndustryType, job.Location, job.NumberOfRequest)
		case "instagram":
			err = scrapper.ScrapInstagram(db, cfg, log, job.JobID, job.IndustryType, job.Location, job.NumberOfRequest)
		default:
			log.Info(fmt.Sprintf("Worker %d processing source %s, industry %s, location %s", id, job.Source, job.IndustryType, job.Location))
		}

		if err != nil {
			log.Error(fmt.Sprintf("Error processing job: %v", err))
		} else {
			db.Model(&model.LeadScrapingJob{}).Where("id = ?", job.JobID).Update("status", "COMPLETED")
		}

	}
}

func (j *JobQueue) StartWorkers(numberOfWorkers int) {
	for i := 1; i <= numberOfWorkers; i++ {
		go worker(i, j.Jobs, j.DB, j.Config, j.Logger)
	}
}

func InitQueue(db *gorm.DB, cfg *config.Config, log logger.Logger) []JobQueue {
	var QueueList []JobQueue
	// Fixed sources for initialization
	sources := []string{"google_maps", "linked_in", "facebook", "instagram"}

	for i, source := range sources {
		sz := cfg.LengthOfJobQueue
		if sz <= 0 {
			sz = 1000
		}
		workers := cfg.NumberOfWorkersPerRequest
		if workers <= 0 {
			workers = 10
		}

		q := NewJobQueue(sz, int64(i+1), db, cfg, log)
		q.Source = source
		q.StartWorkers(int(workers)) // Start workers for this queue
		QueueList = append(QueueList, *q)
	}

	return QueueList
}
