package scrapper

import (
	"context"
	"fmt"
	"lead_scrapper_be/internal/model"
	"time"

	"github.com/chromedp/chromedp"
	"gorm.io/gorm"
)

// semaphore limits the number of concurrent Chrome instances to prevent resource exhaustion
var sem = make(chan struct{}, 3)

func ScrapGoogleMaps(db *gorm.DB, industryType, location string) {
	// fmt.Printf("[Google Maps Scraper] Scraping leads for %s in %s...\n", industryType, location)

	// Acquire semaphore slot - blocks if 3 Chrome instances are already running
	sem <- struct{}{}
	defer func() { <-sem }() // Release slot when done

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var titles []string
	searchURL := fmt.Sprintf("https://www.google.com/maps/search/%s+in+%s", industryType, location)

	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		// Wait for at least one result to appear before extracting
		chromedp.WaitVisible(`div[role="feed"]`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		// Extract business names - try multiple selectors to be resilient
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('.qBF1Pd, .NrDZNb .fontHeadlineSmall, [jstcache] .qBF1Pd'))
				.map(e => e.innerText)
				.filter(t => t.trim() !== '')
		`, &titles),
	)

	if err != nil {
		fmt.Printf("[Google Maps] Error: %v\n", err)
		return
	}

	if len(titles) == 0 {
		fmt.Printf("[Google Maps] No results found for %s in %s\n", industryType, location)
		return
	}

	// --- Fast deduplication: fetch all existing leads for this source+location in ONE query ---
	var existingNames []string
	db.Model(&model.Lead{}).
		Where("source = ? AND location = ?", "google_maps", location).
		Pluck("name", &existingNames)

	// Load into a set (map) for O(1) lookups
	seen := make(map[string]struct{}, len(existingNames))
	for _, n := range existingNames {
		seen[n] = struct{}{}
	}
	// --------------------------------------------------------------------------------------

	var newLeads []model.Lead
	for _, title := range titles {
		if title == "" {
			continue
		}
		if _, exists := seen[title]; exists {
			// Already in DB or already queued in this batch - skip
			continue
		}
		// Mark as seen so other workers in the same batch don't double-insert
		seen[title] = struct{}{}
		newLeads = append(newLeads, model.Lead{
			Name:         title,
			IndustryType: industryType,
			Location:     location,
			Source:       "google_maps",
		})
	}

	if len(newLeads) == 0 {
		fmt.Printf("[Google Maps] All results already stored for %s in %s\n", industryType, location)
		return
	}

	// Bulk insert all new leads in one DB call
	db.Create(&newLeads)
	for _, l := range newLeads {
		fmt.Printf("[Google Maps] New lead saved: %s\n", l.Name)
	}
}

func ScrapLinkedIn(db *gorm.DB, industryType, location string) {
	fmt.Printf("[LinkedIn Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder: LinkedIn requires auth - implement later
}

func ScrapFacebook(db *gorm.DB, industryType, location string) {
	fmt.Printf("[Facebook Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
}

func ScrapInstagram(db *gorm.DB, industryType, location string) {
	fmt.Printf("[Instagram Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
}
