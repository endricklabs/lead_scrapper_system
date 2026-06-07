package scrapper

import (
	"context"
	"fmt"
	"lead_scrapper_be/internal/config"
	"lead_scrapper_be/internal/model"
	"lead_scrapper_be/pkg/logger"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Lead struct {
	Name string
	URL  string
}

// deduplicateLeads removes duplicate leads from the batch based on existing records in the database.
// It checks for duplicates using website URL and name+location combination.
// Only truly new leads are returned.
func deduplicateLeads(tx *gorm.DB, log logger.Logger, tempLeads []Lead, industry, location, contextLog string) ([]Lead, error) {
	if len(tempLeads) == 0 {
		return tempLeads, nil
	}

	// Extract websites and names to check for existing leads
	var websites []string
	var names []string
	for _, lead := range tempLeads {
		if lead.URL != "" {
			websites = append(websites, lead.URL)
		}
		if lead.Name != "" {
			names = append(names, lead.Name)
		}
	}

	// Query for existing leads with matching websites or name+location+industry combinations
	var existingLeads []model.Lead
	query := tx.Where("1 = 0")

	if len(websites) > 0 {
		query = query.Or(tx.Where("website IN ?", websites))
	}

	if len(names) > 0 {
		query = query.Or(tx.Where("name IN ? AND industry_type = ? AND location = ?", names, industry, location))
	}

	if err := query.Find(&existingLeads).Error; err != nil {
		log.Error(fmt.Sprintf("[%s] Error querying existing leads for deduplication: %v", contextLog, err))
		return tempLeads, err
	}

	// Build maps of existing leads for quick lookup
	existingWebsites := make(map[string]bool)
	existingNameLocationIndustry := make(map[string]bool)

	for _, existing := range existingLeads {
		if existing.Website != "" {
			existingWebsites[existing.Website] = true
		}
		if existing.Name != "" {
			key := fmt.Sprintf("%s|%s|%s", existing.Name, existing.IndustryType, existing.Location)
			existingNameLocationIndustry[key] = true
		}
	}

	// Filter out duplicates: only keep leads that don't already exist
	var uniqueLeads []Lead
	for _, lead := range tempLeads {
		isDuplicate := false

		// Check if website already exists
		if lead.URL != "" && existingWebsites[lead.URL] {
			isDuplicate = true
		}

		// Check if name+location+industry combination already exists
		if !isDuplicate && lead.Name != "" {
			key := fmt.Sprintf("%s|%s|%s", lead.Name, industry, location)
			if existingNameLocationIndustry[key] {
				isDuplicate = true
			}
		}

		if !isDuplicate {
			uniqueLeads = append(uniqueLeads, lead)
		}
	}

	if len(uniqueLeads) < len(tempLeads) {
		log.Info(fmt.Sprintf("[%s] Deduplicated leads: filtered out %d duplicates, keeping %d unique leads", contextLog, len(tempLeads)-len(uniqueLeads), len(uniqueLeads)))
	}

	return uniqueLeads, nil
}

// GENERIC LEAD SAVE FUNCTION THAT CAN BE USED BY ALL THE WORKERS FROM ALL THE SOURCES (INVOLVES DEDUPLICATION)
// saveLead saves up to the allowed number of leads for the job.
// It returns the remaining tempLeads, a boolean flag `stop` indicating
// whether scraping should stop (quota reached), and an error if the
// database transaction failed.
func saveLead(db *gorm.DB, log logger.Logger, jobID uuid.UUID, tempLeads []Lead, industry, location, contextLog string) ([]Lead, bool, error) {
	if len(tempLeads) == 0 || jobID == uuid.Nil || db == nil {
		return tempLeads, false, nil
	}

	batchSize := len(tempLeads)
	var limitReachedError error
	stop := false
	var deduplicatedCount int

	err := db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch UserID for the Job
		var job model.LeadScrapingJob
		if err := tx.Select("user_id").First(&job, "id = ?", jobID).Error; err != nil {
			return err
		}

		// 2. Lock the active UserSubscription for this user to serialize concurrent saves
		var activeSub model.UserSubscription
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("SubscriptionPackage").
			Where("user_id = ? AND status = ?", job.UserID, model.UserSubscriptionStatusActive).
			First(&activeSub).Error; err != nil {
			return fmt.Errorf("active subscription not found: %w", err)
		}

		// 3. Limit Enforcement
		limitConfig := activeSub.SubscriptionPackage.MaxLeadsPerMonth
		if limitConfig > 0 {
			var currentMonthScraped int64
			if err := tx.Table("leads").
				Joins("JOIN lead_scraping_jobs ON leads.job_id = lead_scraping_jobs.id").
				Where("lead_scraping_jobs.user_id = ? AND leads.created_at >= ? AND leads.created_at <= ?",
					job.UserID, activeSub.StartDate, activeSub.EndDate).
				Count(&currentMonthScraped).Error; err != nil {
				return err
			}

			if int(currentMonthScraped)+batchSize > limitConfig {
				allowed := limitConfig - int(currentMonthScraped)
				if allowed <= 0 {
					// No capacity left. Mark job as completed and stop further scraping.
					limitReachedError = fmt.Errorf("monthly lead limit reached")
					batchSize = 0
					tempLeads = nil
					stop = true

					if err := tx.Model(&model.LeadScrapingJob{}).Where("id = ?", jobID).Update("status", model.JobStatusCompleted).Error; err != nil {
						return fmt.Errorf("failed to update job status to completed: %w", err)
					}
				} else {
					// Partial capacity available: limit to `allowed` and after saving mark completed.
					limitReachedError = fmt.Errorf("monthly lead limit reached")
					batchSize = allowed
					tempLeads = tempLeads[:allowed]
					// we'll mark job completed after inserting the allowed leads
				}
			}
		}

		if batchSize == 0 {
			return nil
		}

		// Deduplicate leads: filter out any that already exist in the database
		deduplicatedLeads, err := deduplicateLeads(tx, log, tempLeads, industry, location, contextLog)
		if err != nil {
			return fmt.Errorf("deduplication failed: %w", err)
		}

		// Update deduplicatedCount to reflect the deduplicated count
		deduplicatedCount = len(deduplicatedLeads)
		if deduplicatedCount == 0 {
			log.Info(fmt.Sprintf("[%s] All leads were duplicates, skipping insertion for job %s", contextLog, jobID))
			return nil
		}

		// Update job status: accumulate leads_collected with deduplicated count
		if err := tx.Exec("UPDATE lead_scraping_jobs SET leads_collected = leads_collected + ? WHERE id = ?", deduplicatedCount, jobID).Error; err != nil {
			return err
		}

		// Prepare batch for insertion from deduplicated leads
		var batchLeads []model.Lead
		for _, tl := range deduplicatedLeads {
			batchLeads = append(batchLeads, model.Lead{
				Name:         tl.Name,
				IndustryType: industry,
				Location:     location,
				Source:       "google_maps",
				Website:      tl.URL,
				JobID:        jobID,
			})
		}

		// Batch insert new leads
		if len(batchLeads) > 0 {
			if err := tx.Create(&batchLeads).Error; err != nil {
				return err
			}
		}

		// If we previously detected a partial capacity (allowed < requested), mark job completed
		if limitReachedError != nil && deduplicatedCount > 0 {
			if err := tx.Model(&model.LeadScrapingJob{}).Where("id = ?", jobID).Update("status", model.JobStatusCompleted).Error; err != nil {
				return fmt.Errorf("failed to update job status to completed: %w", err)
			}
			stop = true
		}

		return nil
	})

	if err != nil {
		log.Error(fmt.Sprintf("[%s] Transaction failed for job %s: %v", contextLog, jobID, err))
		return tempLeads, false, err
	} else if deduplicatedCount > 0 {
		log.Info(fmt.Sprintf("[%s] Updated job %s and inserted %d leads (deduplicated from %d)", contextLog, jobID, deduplicatedCount, batchSize))
	}

	if limitReachedError != nil {
		log.Error(fmt.Sprintf("[%s] Limit reached for job %s: %v", contextLog, jobID, limitReachedError))
		// If limit reached and we already marked job completed, signal stop without error.
		if stop {
			return []Lead{}, true, nil
		}
		return []Lead{}, true, nil
	}

	return []Lead{}, false, nil // Clear after saving; don't stop
}

// ======================================================================================================
// ======================================================================================================
// ======================================================================================================

func ScrapeGoogleMapsLeads(ctx context.Context, db *gorm.DB, cfg *config.Config, log logger.Logger, jobID uuid.UUID, industry, location string, target int) error {

	var initialLeadsCollected int
	if db != nil && jobID != uuid.Nil {
		var job model.LeadScrapingJob
		if err := db.Select("leads_collected").First(&job, "id = ?", jobID).Error; err == nil {
			initialLeadsCollected = job.LeadsCollected
		} else {
			log.Error(fmt.Sprintf("Could not fetch job %s for initial leads count: %v", jobID, err))
		}
	}

	if initialLeadsCollected >= target {
		log.Info(fmt.Sprintf("Job %s already reached target (%d/%d leads). Skipping scraping.", jobID, initialLeadsCollected, target))
		return nil
	}

	query := industry + " in " + location
	results := []Lead{}

	err := chromedp.Run(ctx,

		chromedp.Navigate("https://www.google.com/maps?hl=en"),
		chromedp.Sleep(6*time.Second),

		// search
		chromedp.Click(`input[name="q"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="q"]`, query, chromedp.ByQuery),
		chromedp.KeyEvent("\r"),

		chromedp.Sleep(10*time.Second),
	)
	if err != nil {
		return err
	}

	tempLeads := []Lead{}

	for initialLeadsCollected+len(results) < target {

		var batch []Lead

		err := chromedp.Run(ctx,

			// wait feed
			chromedp.WaitVisible(`div[role="feed"]`, chromedp.ByQuery),

			// extract visible results
			chromedp.Evaluate(`
			(() => {
				const items = document.querySelectorAll('a.hfpxzc');

				return Array.from(items).map(el => ({
					Name: el.getAttribute('aria-label'),
					URL: el.href
				}));
			})()
			`, &batch),

			// scroll for more
			chromedp.Evaluate(`
			(() => {
				const feed = document.querySelector('div[role="feed"]');
				if (feed) feed.scrollBy(0, 1500);
			})()
			`, nil),

			chromedp.Sleep(2*time.Second),
		)

		if err != nil {
			var stop bool
			tempLeads, stop, _ = saveLead(db, log, jobID, tempLeads, industry, location, "Error Save")
			if stop {
				// Quota reached and job marked completed; stop gracefully
				return nil
			}
			return err
		}

		for _, b := range batch {

			if initialLeadsCollected+len(results) >= target {
				break
			}

			results = append(results, b)
			tempLeads = append(tempLeads, b)

			log.Info(fmt.Sprintf("[GoogleMaps] Session Collected: %d (Total job count: %d/%d)", len(results), initialLeadsCollected+len(results), target))

			// Check if we hit the checkpoint size EXACTLY, or if it's the last lead
			if int64(len(tempLeads)) == cfg.CheckpointNumberOfLeads || initialLeadsCollected+len(results) >= target {
				var saveErr error
				var stop bool
				tempLeads, stop, saveErr = saveLead(db, log, jobID, tempLeads, industry, location, "Checkpoint")
				if saveErr != nil {
					log.Info(fmt.Sprintf("[GoogleMaps] Stopping early due to save issue or limit: %v", saveErr))
					return saveErr
				}
				if stop {
					log.Info(fmt.Sprintf("[GoogleMaps] Stopping early due to quota reached"))
					return nil
				}
			}
		}
	}

	return nil
}

func ScrapGoogleMaps(db *gorm.DB, cfg *config.Config, log logger.Logger, jobID uuid.UUID, industryType, location string, numberOfRequest int) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	err := ScrapeGoogleMapsLeads(ctx, db, cfg, log, jobID, industryType, location, numberOfRequest)
	if err != nil {
		log.Error(fmt.Sprintf("[GoogleMaps] Scraping failed: %v", err))
		return err
	}

	return nil
}

func ScrapLinkedIn(db *gorm.DB, cfg *config.Config, log logger.Logger, jobID uuid.UUID, industryType, location string, numberOfRequest int) error {
	fmt.Printf("[LinkedIn Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
	return nil
}

func ScrapFacebook(db *gorm.DB, cfg *config.Config, log logger.Logger, jobID uuid.UUID, industryType, location string, numberOfRequest int) error {
	fmt.Printf("[Facebook Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
	return nil
}

func ScrapInstagram(db *gorm.DB, cfg *config.Config, log logger.Logger, jobID uuid.UUID, industryType, location string, numberOfRequest int) error {
	fmt.Printf("[Instagram Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
	return nil
}
