package queue

import (
	"fmt"
	"lead_scrapper_be/pkg/scrapper"
	"sync"

	"gorm.io/gorm"
)

type Job struct {
	Source       string
	IndustryType string
	Location     string
}

type JobQueue struct {
	Id     int64
	Source string
	Jobs   chan Job
	wg     *sync.WaitGroup //Each queue has it's pool of workers managed by this WaitGroup
	DB     *gorm.DB
}

func NewJobQueue(size int64, id int64, db *gorm.DB) *JobQueue {
	return &JobQueue{
		Id:     id,
		Jobs:   make(chan Job, size),
		wg:     &sync.WaitGroup{},
		DB:     db,
	}
}

func (j *JobQueue) Enqueue(source string, industryType string, location string) {
	j.Jobs <- Job{
		Source:       source,
		IndustryType: industryType,
		Location:     location,
	}
}

func (j *JobQueue) Wait() {
	j.wg.Wait()
}

//

func worker(id int, jobs <-chan Job, wg *sync.WaitGroup, db *gorm.DB) {
	defer wg.Done()

	for {
		job, ok := <-jobs
		if !ok {
			fmt.Println("Returning from go routine because channel is closed")
			return
		}

		switch job.Source {
		case "google_maps":
			scrapper.ScrapGoogleMaps(db, job.IndustryType, job.Location)
		case "linked_in":
			scrapper.ScrapLinkedIn(db, job.IndustryType, job.Location)
		case "facebook":
			scrapper.ScrapFacebook(db, job.IndustryType, job.Location)
		case "instagram":
			scrapper.ScrapInstagram(db, job.IndustryType, job.Location)
		default:
			fmt.Printf("Worker %d processing source %s, industry %s, location %s\n", id, job.Source, job.IndustryType, job.Location)
		}
	}
}

func (j *JobQueue) StartWorkers(numberOfWorkers int) {
	for i := 1; i <= numberOfWorkers; i++ {
		j.wg.Add(1)
		go worker(i, j.Jobs, j.wg, j.DB)
	}
}

func InitQueue(db *gorm.DB) []JobQueue {
	var QueueList []JobQueue
	// Fixed sources for initialization
	sources := []string{"google_maps", "linked_in", "facebook", "instagram"}

	for i, source := range sources {
		q := NewJobQueue(10, int64(i+1), db)
		q.Source = source
		q.StartWorkers(50) // Start 50 workers for this queue
		QueueList = append(QueueList, *q)
	}

	return QueueList
}

