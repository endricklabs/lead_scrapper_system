package worker

import (
	"fmt"
	"lead_scrapper_be/pkg/queue"
	"sync"
)

type Worker struct {
	wg *sync.WaitGroup
}

func NewWorker() *Worker {
	return &Worker{
		wg: &sync.WaitGroup{},
	}
}

func worker(id int, jobs <-chan queue.Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		job, ok := <-jobs // <-- RECEIVE using <- operator

		if !ok {
			// channel closed, no more jobs
			fmt.Println("Returning from go routine because channel is closed")
			return
		}

		fmt.Println("Worker", id, "processing job", job.JobID, "Source", job.Source, "IndustryType", job.IndustryType, "Location", job.Location)
	}
}

func (w Worker) Run(jobQueue queue.JobQueue, numberOfWorkers int) {
	// start workers
	for i := 1; i <= numberOfWorkers; i++ {
		w.wg.Add(1)
		go worker(i, jobQueue.Jobs, w.wg)
	}
}
