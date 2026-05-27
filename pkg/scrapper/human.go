package scrapper

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/chromedp/chromedp"
)

// ─── User-Agent Pool ──────────────────────────────────────────────────────────

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:123.0) Gecko/20100101 Firefox/123.0",
}

// RandomUserAgent picks a random User-Agent from the pool.
func RandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// ─── Stealth Allocator Options ────────────────────────────────────────────────

// StealthAllocatorOptions returns chromedp exec-allocator options that make
// the browser harder to fingerprint as headless / automated.
func StealthAllocatorOptions(proxyURL string) []chromedp.ExecAllocatorOption {
	ua := RandomUserAgent()
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],

		// ── Appearance ──────────────────────────────────────────────────────
		chromedp.Flag("headless", false), // run headed (most reliable stealth)
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("window-size", "1366,768"),
		chromedp.UserAgent(ua),

		// ── Fingerprint mitigations ─────────────────────────────────────────
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-extensions", false), // leave at default – bots usually set true
		chromedp.Flag("lang", "en-US,en;q=0.9"),

		// ── Network / TLS ────────────────────────────────────────────────────
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("allow-running-insecure-content", true),
	)

	// ── Proxy ────────────────────────────────────────────────────────────────
	if proxyURL != "" {
		opts = append(opts, chromedp.ProxyServer(proxyURL))
	}

	return opts
}

// ─── Timing Helpers ───────────────────────────────────────────────────────────

// randBetween returns a random duration in [min, max).
func randBetween(min, max time.Duration) time.Duration {
	diff := max - min
	if diff <= 0 {
		return min
	}
	return min + time.Duration(rand.Int63n(int64(diff)))
}

// HumanSleep sleeps for a random duration in [min, max) and is a chromedp
// Action so it can be composed inline.
func HumanSleep(min, max time.Duration) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		d := randBetween(min, max)
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// ─── Typing Helpers ───────────────────────────────────────────────────────────

// HumanType sends keys one character at a time with a small random pause
// between each keystroke (20–90 ms), mimicking real typing speed.
func HumanType(sel, text string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		for _, ch := range text {
			if err := chromedp.SendKeys(sel, string(ch), chromedp.ByQuery).Do(ctx); err != nil {
				return err
			}
			time.Sleep(randBetween(20*time.Millisecond, 90*time.Millisecond))
		}
		return nil
	}
}

// ─── Scroll Helpers ───────────────────────────────────────────────────────────

// randomScrollAmount returns a pixel value close to base with ±20% jitter.
func randomScrollAmount(base int) int {
	jitter := base / 5 // 20%
	return base - jitter + rand.Intn(2*jitter+1)
}

// HumanScrollFeed scrolls the Google Maps feed by a randomised amount.
func HumanScrollFeed() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		px := randomScrollAmount(1200)
		js := fmt.Sprintf(`
			(() => {
				const feed = document.querySelector('div[role="feed"]');
				if (feed) feed.scrollBy(0, %d);
			})()
		`, px)
		return chromedp.Evaluate(js, nil).Do(ctx)
	}
}

// HumanScrollWindow scrolls the window (e.g. for LinkedIn) by a random amount.
func HumanScrollWindow() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		px := randomScrollAmount(1200)
		js := fmt.Sprintf(`window.scrollBy(0, %d)`, px)
		return chromedp.Evaluate(js, nil).Do(ctx)
	}
}

// ─── Stealth JS Injection ─────────────────────────────────────────────────────

// InjectStealthJS patches common JavaScript properties that betrays headless /
// automated Chrome (navigator.webdriver, plugins length, etc.).
// Must be called after the browser context is created but before navigation.
func InjectStealthJS() chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		script := `
			// Hide webdriver flag
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });

			// Fake plugins (real browsers always have some)
			Object.defineProperty(navigator, 'plugins', {
				get: () => [1, 2, 3, 4, 5],
			});

			// Fake languages
			Object.defineProperty(navigator, 'languages', {
				get: () => ['en-US', 'en'],
			});

			// Chrome runtime stub to look like a real Chrome install
			window.chrome = { runtime: {} };

			// Permissions stub – bots tend to have a broken Notification API
			const origQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) =>
				parameters.name === 'notifications'
					? Promise.resolve({ state: Notification.permission })
					: origQuery(parameters);
		`
		return chromedp.Evaluate(script, nil).Do(ctx)
	})
}
