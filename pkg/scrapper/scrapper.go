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

func ScrapeGoogleMapsLeads(ctx context.Context, industry, location string, target int) ([]Lead, error) {

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

	leads, err := ScrapeGoogleMapsLeads(ctx, industryType, location, numberOfRequest)
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

func ScrapeLinkedInLeads(ctx context.Context, industry, location string, target int) ([]Lead, error) {
	// LinkedIn public company search: no login required for basic results
	searchURL := fmt.Sprintf(
		"https://www.linkedin.com/search/results/companies/?keywords=%s%%20%s&origin=GLOBAL_SEARCH_HEADER",
		industry, location,
	)

	results := []Lead{}
	seen := map[string]bool{}

	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(6*time.Second),
	)
	if err != nil {
		return nil, err
	}

	for len(results) < target {
		var batch []Lead

		err := chromedp.Run(ctx,
			// wait for result cards
			chromedp.WaitVisible(`.search-results-container`, chromedp.ByQuery),

			// extract company name + URL from each result card
			chromedp.Evaluate(`
			(() => {
				const cards = document.querySelectorAll('.entity-result__item');
				return Array.from(cards).map(card => {
					const anchor = card.querySelector('.app-aware-link[href*="/company/"]');
					const nameEl = card.querySelector('.entity-result__title-text');
					if (!anchor || !nameEl) return null;
					return {
						Name: nameEl.innerText.trim(),
						URL:  anchor.href.split('?')[0],
					};
				}).filter(Boolean);
			})()
			`, &batch),

			// scroll down for more results
			chromedp.Evaluate(`window.scrollBy(0, 1500)`, nil),
			chromedp.Sleep(2*time.Second),
		)
		if err != nil {
			return nil, err
		}

		prevLen := len(results)
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

		// no new results after a scroll → we've hit the end of the page
		if len(results) == prevLen {
			fmt.Println("[LinkedIn] No new results after scroll, stopping early.")
			break
		}
	}

	return results, nil
}

func ScrapLinkedIn(db *gorm.DB, industryType, location string, numberOfRequest int) {
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

	leads, err := ScrapeLinkedInLeads(ctx, industryType, location, numberOfRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n🔗 FINAL LEADS (LinkedIn):\n")

	for i, l := range leads {
		fmt.Printf("%d.\nName: %s\nLinkedIn URL: %s\n\n",
			i+1,
			l.Name,
			l.URL,
		)
	}
}

func ScrapFacebook(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[Facebook Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
}

func ScrapInstagram(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[Instagram Scraper] Scraping leads for %s in %s...\n", industryType, location)
	// Placeholder
}
