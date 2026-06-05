package queue

import (
	"fmt"
	"lead_scrapper_be/internal/config"
	"lead_scrapper_be/internal/model"
	"lead_scrapper_be/pkg/logger"
	"time"

	"gorm.io/gorm"
)

func PollPendingJobs(db *gorm.DB, queues []JobQueue, cfg *config.Config, log logger.Logger) {
	var pendingJobs []model.LeadScrapingJob

	// Find Pending jobs
	if err := db.Where("status = ?", "PENDING").Find(&pendingJobs).Error; err != nil {
		log.Error(fmt.Sprintf("Error fetching pending jobs: %v", err))
		return
	}
	if len(pendingJobs) > 0 {
		log.Info(fmt.Sprintf("[Outbox Poller] Found %d pending jobs in database. Enqueuing them now...", len(pendingJobs)))
	}

	// Traverse through the pending jobs and enqueue each of them based on the source queue
	for _, job := range pendingJobs {

		// Find matching queue
		enqueued := false
		for i := range queues {
			if queues[i].Source == job.Source {
				queues[i].Enqueue(job.ID, job.Source, job.IndustryType, job.Location, job.TargetRequested, job.LeadsCollected)
				enqueued = true
				break
			}
		}

		if !enqueued {
			log.Info(fmt.Sprintf("No queue found for source: %s", job.Source))
		}
	}
}

// StartOutboxPoller runs in the background and pulls PENDING jobs into the active queue
func StartOutboxPoller(db *gorm.DB, queues []JobQueue, cfg *config.Config, log logger.Logger) {
	interval := cfg.OutboxPollingInterval
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	log.Info(fmt.Sprintf("[Outbox Poller] Initializing with an interval of %v", interval))

	ticker := time.NewTicker(interval)
	go func() {
		// Initial fetch on server start
		log.Info("[Outbox Poller] Performing initial startup check for pending jobs...")
		PollPendingJobs(db, queues, cfg, log)

		// Keep fetching at the configured interval
		for range ticker.C {
			log.Info("[Outbox Poller] Interval tick reached. Running scheduled check for pending jobs...")
			PollPendingJobs(db, queues, cfg, log)
		}
	}()
}
