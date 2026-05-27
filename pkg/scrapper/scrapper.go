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

// ─── Google Maps ──────────────────────────────────────────────────────────────

func ScrapeGoogleMapsLeads(ctx context.Context, industry, location string, target int) ([]Lead, error) {
	query := industry + " in " + location
	results := []Lead{}

	err := chromedp.Run(ctx,
		// Patch navigator before any page loads
		InjectStealthJS(),

		chromedp.Navigate("https://www.google.com/maps?hl=en"),

		// Human-like wait after page load: 5–9 s
		HumanSleep(5*time.Second, 9*time.Second),

		// Click the search box first (humans click before typing)
		chromedp.Click(`input[name="q"]`, chromedp.ByQuery),
		HumanSleep(300*time.Millisecond, 800*time.Millisecond),

		// Type character-by-character like a real user
		HumanType(`input[name="q"]`, query),
		HumanSleep(400*time.Millisecond, 900*time.Millisecond),

		// Press Enter
		chromedp.KeyEvent("\r"),

		// Wait for results to load: 8–13 s
		HumanSleep(8*time.Second, 13*time.Second),
	)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	stallCount := 0 // detect end-of-results / stall

	for len(results) < target {
		var batch []Lead

		err := chromedp.Run(ctx,
			// Wait for feed
			chromedp.WaitVisible(`div[role="feed"]`, chromedp.ByQuery),

			// Extract visible results
			chromedp.Evaluate(`
			(() => {
				const items = document.querySelectorAll('a.hfpxzc');
				return Array.from(items).map(el => ({
					Name: el.getAttribute('aria-label'),
					URL:  el.href
				}));
			})()
			`, &batch),

			// Randomised scroll
			HumanScrollFeed(),

			// Human-like inter-scroll pause: 1.5–4 s
			HumanSleep(1500*time.Millisecond, 4*time.Second),
		)
		if err != nil {
			return nil, err
		}

		prevLen := len(results)

		// Dedupe
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

		// Stall detection – stop if we haven't gathered new leads for 3 cycles
		if len(results) == prevLen {
			stallCount++
			if stallCount >= 3 {
				fmt.Println("[GoogleMaps] No new results after 3 scrolls, stopping early.")
				break
			}
			// Extra wait on stall before retrying
			chromedp.Sleep(2 * time.Second).Do(ctx) //nolint:errcheck
		} else {
			stallCount = 0
		}
	}

	return results, nil
}

func ScrapGoogleMaps(db *gorm.DB, industryType, location string, numberOfRequest int) {
	proxyURL := GlobalProxyPool.GetNext()
	opts := StealthAllocatorOptions(proxyURL)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Generous timeout (stealth mode is slower)
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	leads, err := ScrapeGoogleMapsLeads(ctx, industryType, location, numberOfRequest)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n📍 FINAL LEADS:\n")
	for i, l := range leads {
		fmt.Printf("%d.\nName: %s\nMaps URL: %s\n\n", i+1, l.Name, l.URL)
	}
}

// ─── LinkedIn ─────────────────────────────────────────────────────────────────

func ScrapLinkedIn(db *gorm.DB, industry, location string, target int) ([]Lead, error) {
	fmt.Printf("[LinkedIn Scraper] Scraping leads for %s in %s...\n", industry, location)
	return nil, nil
}

// ─── Stubs ────────────────────────────────────────────────────────────────────

func ScrapFacebook(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[Facebook Scraper] Scraping leads for %s in %s...\n", industryType, location)
}

func ScrapInstagram(db *gorm.DB, industryType, location string, numberOfRequest int) {
	fmt.Printf("[Instagram Scraper] Scraping leads for %s in %s...\n", industryType, location)
}
