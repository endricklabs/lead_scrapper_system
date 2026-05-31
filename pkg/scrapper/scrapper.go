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
)

type Lead struct {
	Name string
	URL  string
}

// GENERIC LEAD SAVE FUNCTION THAT CAN BE USED BY ALL THE WORKERS FROM ALL THE SOURCES (SHOULD INVOLVE DEDEPLICATION)
func saveLead(db *gorm.DB, log logger.Logger, jobID uuid.UUID, tempLeads []Lead, industry, location, contextLog string) []Lead {
	if len(tempLeads) == 0 || jobID == uuid.Nil || db == nil {
		return tempLeads
	}

	batchSize := len(tempLeads)
	err := db.Transaction(func(tx *gorm.DB) error {
		// Update job status: accumulate leads_collected
		if err := tx.Exec("UPDATE lead_scraping_jobs SET leads_collected = leads_collected + ? WHERE id = ?", batchSize, jobID).Error; err != nil {
			return err
		}

		// Prepare batch for insertion
		var batchLeads []model.Lead
		for _, tl := range tempLeads {
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

		return nil
	})

	if err != nil {
		log.Error(fmt.Sprintf("[%s] Transaction failed for job %s: %v", contextLog, jobID, err))
	} else {
		log.Info(fmt.Sprintf("[%s] Updated job %s and inserted %d leads", contextLog, jobID, batchSize))
	}

	return []Lead{} // Clear after saving
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
			tempLeads = saveLead(db, log, jobID, tempLeads, industry, location, "Error Save")
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
				tempLeads = saveLead(db, log, jobID, tempLeads, industry, location, "Checkpoint")
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
