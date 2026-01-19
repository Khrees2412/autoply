package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/khrees2412/autoply/internal/config"
	"github.com/khrees2412/autoply/pkg/models"
)

// SearchProgress provides feedback during job searches
type SearchProgress struct {
	mu           sync.Mutex
	currentBoard string
	status       string
	jobsFound    int
}

func (p *SearchProgress) SetBoard(board string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentBoard = board
	p.status = "searching"
	p.jobsFound = 0
	fmt.Printf("\r\033[K⏳ Searching %s...", board)
}

func (p *SearchProgress) SetStatus(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = status
	fmt.Printf("\r\033[K⏳ %s: %s...", p.currentBoard, status)
}

func (p *SearchProgress) Complete(jobsFound int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.jobsFound = jobsFound
	p.status = "complete"
	fmt.Printf("\r\033[K✓ %s: found %d jobs\n", p.currentBoard, jobsFound)
}

func (p *SearchProgress) Error(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = "error"
	fmt.Printf("\r\033[K✗ %s: %v\n", p.currentBoard, err)
}

const (
	pageLoadTimeout = 30 * time.Second
	rateLimitDelay  = 2 * time.Second
)

// createBrowserContext creates a new browser context with appropriate options
func createBrowserContext(parent context.Context) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("useAutomationExtension", false),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(parent, opts...)
	// Suppress noisy chromedp log messages - redirect to discard for unmarshal warnings
	ctx, cancel2 := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(format string, v ...interface{}) {
		// Filter out noisy unmarshal warnings
		msg := fmt.Sprintf(format, v...)
		if strings.Contains(msg, "could not unmarshal event") ||
			strings.Contains(msg, "unknown PrivateNetworkRequestPolicy") ||
			strings.Contains(msg, "unknown ClientNavigationReason") {
			// Suppress these warnings
			return
		}
		// Log actual errors
		log.Printf(format, v...)
	}))

	// Combine cancel functions
	return ctx, func() {
		cancel2()
		cancel()
	}
}

// SearchAllSources searches all available job boards with progress feedback
func SearchAllSources(query, location string) ([]*models.Job, error) {
	var allJobs []*models.Job
	progress := &SearchProgress{}

	fmt.Println("🔍 Starting job search across all boards...")
	fmt.Println()

	// Search LinkedIn
	progress.SetBoard("LinkedIn")
	jobs, err := SearchLinkedIn(query, location)
	if err != nil {
		progress.Error(err)
	} else {
		progress.Complete(len(jobs))
		allJobs = append(allJobs, jobs...)
	}

	fmt.Println()
	fmt.Printf("📋 Total jobs found: %d\n", len(allJobs))

	return allJobs, nil
}

// SearchJobs searches a specific job board
func SearchJobs(source, query, location string) ([]*models.Job, error) {
	progress := &SearchProgress{}
	progress.SetBoard(source)

	var jobs []*models.Job
	var err error

	switch strings.ToLower(source) {
	case "linkedin":
		jobs, err = SearchLinkedIn(query, location)
	case "greenhouse":
		jobs, err = SearchGreenhouse(query, location)
	case "lever":
		jobs, err = SearchLever(query, location)
	default:
		return nil, fmt.Errorf("unsupported source: %s. Available: linkedin, greenhouse, lever", source)
	}

	if err != nil {
		progress.Error(err)
		return nil, err
	}
	progress.Complete(len(jobs))
	return jobs, nil
}

// SearchLinkedIn searches LinkedIn jobs using browser automation
func SearchLinkedIn(query, location string) ([]*models.Job, error) {
	email := config.Get("linkedin_email")
	password := config.Get("linkedin_password")

	if email == "" || password == "" {
		return nil, fmt.Errorf("LinkedIn credentials not configured. Set them with:\n  autoply config set linkedin_email your@email.com\n  autoply config set linkedin_password yourpassword")
	}

	ctx, cancel := createBrowserContext(context.Background())
	defer cancel()

	// Longer timeout for LinkedIn (login + search + scrolling)
	ctx, cancel = context.WithTimeout(ctx, 3*pageLoadTimeout)
	defer cancel()

	var jobs []*models.Job
	var loginErr error

	// Step 1: Login to LinkedIn
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.linkedin.com/login"),
		chromedp.WaitVisible(`input[name="session_key"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`input[name="session_key"]`, email, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.SendKeys(`input[name="session_password"]`, password, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // Wait for login to complete
		// Verify login was successful by checking for common post-login elements
		chromedp.ActionFunc(func(ctx context.Context) error {
			var currentURL string
			if err := chromedp.Location(&currentURL).Do(ctx); err != nil {
				return err
			}
			// Check if we're still on login page or hit a checkpoint
			if strings.Contains(currentURL, "/login") || strings.Contains(currentURL, "/checkpoint") {
				loginErr = fmt.Errorf("login failed - check credentials or verify account at linkedin.com")
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("LinkedIn login error: %w", err)
	}
	if loginErr != nil {
		return nil, loginErr
	}

	// Step 2: Navigate to job search
	searchURL := buildLinkedInSearchURL(query, location)
	err = chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		chromedp.Sleep(4*time.Second), // Wait for page load
		// Wait for job results container to appear
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to wait for job listings, but don't fail if not found immediately
			chromedp.WaitVisible(`.jobs-search-results-list, .scaffold-layout__list, [data-job-id]`, chromedp.ByQuery).Do(ctx)
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("LinkedIn navigation error: %w", err)
	}

	// Step 3: Scroll to load more jobs
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Find the scrollable job list container and scroll within it
			for i := 0; i < 5; i++ {
				// Try scrolling the job list container specifically
				chromedp.Evaluate(`
					(() => {
						const container = document.querySelector('.jobs-search-results-list, .scaffold-layout__list-container');
						if (container) {
							container.scrollTop = container.scrollHeight;
						} else {
							window.scrollTo(0, document.body.scrollHeight);
						}
					})()
				`, nil).Do(ctx)
				chromedp.Sleep(1500 * time.Millisecond).Do(ctx)
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("LinkedIn scroll error: %w", err)
	}

	// Step 4: Extract job listings
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var jobElements []map[string]string
			err := chromedp.Evaluate(`
				(() => {
					const jobs = [];
					// Multiple selector strategies for different LinkedIn layouts
					const selectors = [
						'.job-card-container',
						'.jobs-search-results__list-item',
						'[data-job-id]',
						'.scaffold-layout__list-item',
						'li.jobs-search-results__list-item',
						'.job-card-list__entity-lockup'
					];

					let jobCards = [];
					for (const sel of selectors) {
						const cards = document.querySelectorAll(sel);
						if (cards.length > 0) {
							jobCards = cards;
							break;
						}
					}

					jobCards.forEach((card, index) => {
						if (index >= 50) return;

						// Multiple selector strategies for job title
						const titleSelectors = [
							'.job-card-list__title',
							'.job-card-container__link',
							'a.job-card-list__title',
							'.artdeco-entity-lockup__title',
							'[class*="job-card"] a[href*="/jobs/view"]',
							'a[data-control-name="job_card_title"]',
							'.base-search-card__title'
						];

						// Multiple selector strategies for company
						const companySelectors = [
							'.job-card-container__primary-description',
							'.job-card-container__company-name',
							'.artdeco-entity-lockup__subtitle',
							'[class*="company"]',
							'.base-search-card__subtitle'
						];

						// Multiple selector strategies for location
						const locationSelectors = [
							'.job-card-container__metadata-item',
							'.job-card-container__metadata-wrapper li',
							'.artdeco-entity-lockup__caption',
							'[class*="location"]',
							'.base-search-card__metadata'
						];

						let title = '', company = '', location = '', url = '';

						// Find title
						for (const sel of titleSelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								title = el.textContent.trim();
								if (el.href) url = el.href;
								break;
							}
						}

						// Find company
						for (const sel of companySelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								company = el.textContent.trim();
								break;
							}
						}

						// Find location
						for (const sel of locationSelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								location = el.textContent.trim();
								break;
							}
						}

						// Find URL if not already found
						if (!url) {
							const linkEl = card.querySelector('a[href*="/jobs/view/"]') || card.querySelector('a[href*="/jobs/"]');
							if (linkEl) {
								url = linkEl.href.startsWith('http') ? linkEl.href : 'https://www.linkedin.com' + linkEl.getAttribute('href');
							}
						}

						if (title && title.length > 2) {
							jobs.push({ title, company, location, url });
						}
					});

					return jobs;
				})()
			`, &jobElements).Do(ctx)
			if err != nil {
				return err
			}

			for _, jobData := range jobElements {
				job := &models.Job{
					Title:      jobData["title"],
					Company:    jobData["company"],
					Location:   jobData["location"],
					URL:        jobData["url"],
					Source:     "linkedin",
					ScrapedAt:  time.Now(),
					MatchScore: 0,
				}
				if job.Title != "" {
					jobs = append(jobs, job)
				}
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("LinkedIn extraction error: %w", err)
	}

	if len(jobs) == 0 {
		return jobs, nil // Return empty list without error
	}

	// Step 5: Fetch detailed descriptions for top jobs
	for i, job := range jobs {
		if i >= 10 { // Limit to first 10 to avoid rate limiting
			break
		}
		if job.URL != "" {
			desc, err := fetchLinkedInJobDescription(ctx, job.URL)
			if err == nil && desc != "" {
				jobs[i].Description = desc
			}
			time.Sleep(rateLimitDelay)
		}
	}

	return jobs, nil
}

// fetchLinkedInJobDescription fetches the full job description from a LinkedIn job URL
func fetchLinkedInJobDescription(ctx context.Context, url string) (string, error) {
	var description string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to click "Show more" button if it exists
			showMoreSelectors := []string{
				`button[aria-label*="Show more"]`,
				`button[aria-label*="see more"]`,
				`.show-more-less-html__button`,
				`button.jobs-description__footer-button`,
			}
			for _, sel := range showMoreSelectors {
				chromedp.Click(sel, chromedp.ByQuery).Do(ctx)
			}
			chromedp.Sleep(1 * time.Second).Do(ctx)
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try multiple selectors for job description
			descSelectors := []string{
				`.jobs-description-content__text`,
				`.show-more-less-html__markup`,
				`.jobs-box__html-content`,
				`#job-details`,
				`.description__text`,
			}
			for _, sel := range descSelectors {
				var text string
				if err := chromedp.Text(sel, &text, chromedp.ByQuery).Do(ctx); err == nil && text != "" {
					description = strings.TrimSpace(text)
					return nil
				}
			}
			return nil
		}),
	)
	return description, err
}

// buildLinkedInSearchURL constructs LinkedIn job search URL
func buildLinkedInSearchURL(query, location string) string {
	baseURL := "https://www.linkedin.com/jobs/search"
	params := []string{}

	if query != "" {
		params = append(params, "keywords="+strings.ReplaceAll(query, " ", "%20"))
	}
	if location != "" {
		params = append(params, "location="+strings.ReplaceAll(location, " ", "%20"))
	}
	params = append(params, "f_TPR=r86400") // Last 24 hours
	params = append(params, "f_E=2")        // Full-time (can be customized)

	if len(params) > 0 {
		return baseURL + "?" + strings.Join(params, "&")
	}
	return baseURL
}

// SearchGreenhouse searches Greenhouse jobs (company-specific)
func SearchGreenhouse(query, location string) ([]*models.Job, error) {
	// Greenhouse requires company-specific URLs
	// Example: https://boards.greenhouse.io/companyname
	// This would need a list of companies to search
	return []*models.Job{}, fmt.Errorf("Greenhouse search requires company-specific URLs. Use manual job entry instead")
}

// SearchLever searches Lever jobs (company-specific)
func SearchLever(query, location string) ([]*models.Job, error) {
	// Lever requires company-specific URLs
	// Example: https://jobs.lever.co/companyname
	// This would need a list of companies to search
	return []*models.Job{}, fmt.Errorf("Lever search requires company-specific URLs. Use manual job entry instead")
}

// SearchStartupJobs searches startup.jobs for job listings
func SearchStartupJobs(query, location string) ([]*models.Job, error) {
	ctx, cancel := createBrowserContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, pageLoadTimeout)
	defer cancel()

	var jobs []*models.Job
	url := buildStartupJobsSearchURL(query, location)

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Remove automation indicators
			chromedp.Evaluate(`
				Object.defineProperty(navigator, 'webdriver', {
					get: () => undefined,
				});
			`, nil).Do(ctx)
			return nil
		}),
		chromedp.Navigate(url),
		chromedp.Sleep(4*time.Second), // Longer wait for Cloudflare
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Check if we got blocked or if page loaded correctly
			var pageTitle string
			var pageText string
			chromedp.Title(&pageTitle).Do(ctx)
			chromedp.Text("body", &pageText).Do(ctx)

			// Check for common block indicators
			if strings.Contains(strings.ToLower(pageTitle), "access denied") ||
				strings.Contains(strings.ToLower(pageTitle), "blocked") ||
				strings.Contains(strings.ToLower(pageText), "access denied") ||
				strings.Contains(strings.ToLower(pageText), "cloudflare") ||
				strings.Contains(strings.ToLower(pageText), "robot check") ||
				strings.Contains(strings.ToLower(pageText), "checking your browser") ||
				strings.Contains(strings.ToLower(pageText), "enable javascript") {
				return fmt.Errorf("startup.jobs is blocking automated access - try manual browsing")
			}

			// Try to wait for job listings to appear
			chromedp.WaitVisible(`.job, .position, [data-job], article`, chromedp.ByQuery).Do(ctx)

			// Scroll to load more jobs
			for i := 0; i < 3; i++ {
				chromedp.Evaluate("window.scrollTo(0, document.body.scrollHeight)", nil).Do(ctx)
				chromedp.Sleep(2 * time.Second).Do(ctx)
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var jobElements []map[string]string
			err := chromedp.Evaluate(`
				(() => {
					const jobs = [];
					// Try multiple selector strategies for startup.jobs
					const selectors = [
						'[data-job]',
						'.job-card',
						'.job-listing',
						'article',
						'.job',
						'a[href*="/job/"]',
						'.position',
						'.listing',
						'.opportunity',
						'.role',
						'.job-item',
						'.vacancy'
					];

					let jobCards = [];
					for (const sel of selectors) {
						const cards = document.querySelectorAll(sel);
						if (cards.length > 0) {
							jobCards = cards;
							break;
						}
					}

					jobCards.forEach((card, index) => {
						if (index >= 50) return;

						// Multiple selector strategies for title
						const titleSelectors = [
							'h2', 'h3', 'h4',
							'.job-title',
							'[class*="title"]',
							'.position-title',
							'.role-title',
							'a',
							'.name'
						];

						// Multiple selector strategies for company
						const companySelectors = [
							'.company',
							'.company-name',
							'[class*="company"]',
							'.employer',
							'.organization',
							'.startup-name'
						];

						// Multiple selector strategies for location
						const locationSelectors = [
							'.location',
							'[class*="location"]',
							'.city',
							'.place',
							'.office',
							'.remote'
						];

						let title = '', company = '', location = '', url = '';

						// Find title
						for (const sel of titleSelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								title = el.textContent.trim();
								if (el.href) url = el.href;
								break;
							}
						}

						// Find company
						for (const sel of companySelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								company = el.textContent.trim();
								break;
							}
						}

						// Find location
						for (const sel of locationSelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								location = el.textContent.trim();
								break;
							}
						}

						// Find URL if not already found
						if (!url) {
							const linkEl = card.querySelector('a[href*="/job/"]') || 
											card.querySelector('a[href*="startup.jobs"]') ||
											card.querySelector('a');
							if (linkEl && linkEl.href) {
								url = linkEl.href.startsWith('http') ? linkEl.href : 'https://startup.jobs' + linkEl.getAttribute('href');
							}
						}

						if (title && title.length > 2) {
							jobs.push({ title, company, location, url });
						}
					});
					return jobs;
				})()
			`, &jobElements).Do(ctx)
			if err != nil {
				return err
			}

			for _, jobData := range jobElements {
				job := &models.Job{
					Title:      jobData["title"],
					Company:    jobData["company"],
					Location:   jobData["location"],
					URL:        jobData["url"],
					Source:     "startup.jobs",
					ScrapedAt:  time.Now(),
					MatchScore: 0,
				}
				if job.Title != "" {
					jobs = append(jobs, job)
				}
			}
			return nil
		}),
	)

	if err != nil {
		if strings.Contains(err.Error(), "blocking automated access") {
			return []*models.Job{}, fmt.Errorf("startup.jobs is blocking automated access - this site may require manual browsing")
		}
		return nil, fmt.Errorf("Startup.jobs scraping error: %w", err)
	}

	// If no jobs found, try a different approach or return empty
	if len(jobs) == 0 {
		return []*models.Job{}, nil
	}

	// Fetch detailed descriptions for a subset of jobs
	for i, job := range jobs {
		if i >= 20 {
			break
		}
		if job.URL != "" {
			desc, err := fetchStartupJobsDescription(ctx, job.URL)
			if err == nil {
				jobs[i].Description = desc
			}
			time.Sleep(rateLimitDelay)
		}
	}

	return jobs, nil
}

// fetchStartupJobsDescription fetches the full job description from a startup.jobs URL
func fetchStartupJobsDescription(ctx context.Context, url string) (string, error) {
	var description string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try multiple selectors for job description
			selectors := []string{
				".job-description",
				"[class*='description']",
				"article",
				".content",
				"main",
			}
			for _, sel := range selectors {
				var text string
				if err := chromedp.Text(sel, &text, chromedp.ByQuery).Do(ctx); err == nil && text != "" {
					description = text
					return nil
				}
			}
			return nil
		}),
	)
	return description, err
}

// buildStartupJobsSearchURL constructs startup.jobs search URL
func buildStartupJobsSearchURL(query, location string) string {
	baseURL := "https://startup.jobs"
	params := []string{}

	if query != "" {
		params = append(params, "q="+strings.ReplaceAll(query, " ", "+"))
	}
	if location != "" {
		params = append(params, "location="+strings.ReplaceAll(location, " ", "+"))
	}

	if len(params) > 0 {
		return baseURL + "?" + strings.Join(params, "&")
	}
	return baseURL
}

// SearchGlassdoor searches Glassdoor jobs
func SearchGlassdoor(query, location string) ([]*models.Job, error) {
	ctx, cancel := createBrowserContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, pageLoadTimeout)
	defer cancel()

	var jobs []*models.Job
	url := buildGlassdoorSearchURL(query, location)

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Remove automation indicators
			chromedp.Evaluate(`
				Object.defineProperty(navigator, 'webdriver', {
					get: () => undefined,
				});
			`, nil).Do(ctx)
			return nil
		}),
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Scroll to load more jobs
			for i := 0; i < 3; i++ {
				chromedp.Evaluate("window.scrollTo(0, document.body.scrollHeight)", nil).Do(ctx)
				chromedp.Sleep(1500 * time.Millisecond).Do(ctx)
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var jobElements []map[string]string
			err := chromedp.Evaluate(`
				(() => {
					const jobs = [];
					// Glassdoor job card selectors
					const selectors = [
						'.jobCard',
						'[data-test="job-card"]',
						'.job-listing',
						'.jl',
						'.job-search-result',
						'.jobResult',
						'.JobCard'
					];

					let jobCards = [];
					for (const sel of selectors) {
						const cards = document.querySelectorAll(sel);
						if (cards.length > 0) {
							jobCards = cards;
							break;
						}
					}

					jobCards.forEach((card, index) => {
						if (index >= 50) return;

						// Title selectors
						const titleSelectors = [
							'a[data-test="job-title"]',
							'.job-title',
							'.jobTitle',
							'h3',
							'.title',
							'a'
						];

						// Company selectors
						const companySelectors = [
							'[data-test="employer-name"]',
							'.employer-name',
							'.company',
							'.companyName',
							'[class*="company"]'
						];

						// Location selectors
						const locationSelectors = [
							'[data-test="employer-location"]',
							'.location',
							'.job-location',
							'[class*="location"]'
						];

						let title = '', company = '', location = '', url = '';

						// Find title
						for (const sel of titleSelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								title = el.textContent.trim();
								if (el.href) url = el.href;
								break;
							}
						}

						// Find company
						for (const sel of companySelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								company = el.textContent.trim();
								break;
							}
						}

						// Find location
						for (const sel of locationSelectors) {
							const el = card.querySelector(sel);
							if (el && el.textContent.trim()) {
								location = el.textContent.trim();
								break;
							}
						}

						// Find URL if not already found
						if (!url) {
							const linkEl = card.querySelector('a[href*="/partner/"]') || card.querySelector('a');
							if (linkEl && linkEl.href) {
								url = linkEl.href.startsWith('http') ? linkEl.href : 'https://www.glassdoor.com' + linkEl.getAttribute('href');
							}
						}

						if (title && title.length > 2) {
							jobs.push({ title, company, location, url });
						}
					});
					return jobs;
				})()
			`, &jobElements).Do(ctx)
			if err != nil {
				return err
			}

			for _, jobData := range jobElements {
				job := &models.Job{
					Title:      jobData["title"],
					Company:    jobData["company"],
					Location:   jobData["location"],
					URL:        jobData["url"],
					Source:     "glassdoor",
					ScrapedAt:  time.Now(),
					MatchScore: 0,
				}
				if job.Title != "" {
					jobs = append(jobs, job)
				}
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("Glassdoor scraping error: %w", err)
	}

	return jobs, nil
}

// buildGlassdoorSearchURL constructs Glassdoor job search URL
func buildGlassdoorSearchURL(query, location string) string {
	baseURL := "https://www.glassdoor.com/Job/jobs.htm"
	params := []string{}

	if query != "" {
		params = append(params, "keyword="+strings.ReplaceAll(query, " ", "+"))
	}
	if location != "" {
		params = append(params, "location="+strings.ReplaceAll(location, " ", "+"))
	}

	if len(params) > 0 {
		return baseURL + "?" + strings.Join(params, "&")
	}
	return baseURL
}
