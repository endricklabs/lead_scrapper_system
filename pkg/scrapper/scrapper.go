package scrapper

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"gorm.io/gorm"
)

type Lead struct {
	Name string
	URL  string
}

func ScrapeLeads(ctx context.Context, industry, location string, target int) ([]Lead, error) {

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
		return nil, err
	}

	seen := map[string]bool{}

	for len(results) < target {

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
			return nil, err
		}

		// dedupe
		for _, b := range batch {

			if len(results) >= target {
				break
			}

			if b.Name == "" || seen[b.URL] {
				continue
			}

			seen[b.URL] = true

			results = append(results, b)

			fmt.Println("Collected:", len(results))
		}
	}

	return results, nil
}

func ScrapGoogleMaps(db *gorm.DB, industryType, location string, numberOfRequest int) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	leads, err := ScrapeLeads(ctx, industryType, location, numberOfRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n📍 FINAL LEADS:\n")

	for i, l := range leads {
		fmt.Printf("%d.\nName: %s\nMaps URL: %s\n\n",
			i+1,
			l.Name,
			l.URL,
		)
	}

}

func ScrapLinkedIn(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[LinkedIn Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder: LinkedIn requires auth - implement later
}

func ScrapFacebook(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[Facebook Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
}

func ScrapInstagram(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[Instagram Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
}
