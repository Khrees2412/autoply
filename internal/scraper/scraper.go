package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/khrees2412/autoply/internal/config"
	"github.com/khrees2412/autoply/pkg/models"
)

const (
	maxJobsPerSource = 50
	pageLoadTimeout  = 30 * time.Second
	elementWaitTime  = 5 * time.Second
	rateLimitDelay   = 2 * time.Second
)

// createBrowserContext creates a new browser context with appropriate options
func createBrowserContext(parent context.Context) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(parent, opts...)
	ctx, cancel2 := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	// Combine cancel functions
	return ctx, func() {
		cancel2()
		cancel()
	}
}

// SearchAllSources searches all available job boards
func SearchAllSources(query, location string) ([]*models.Job, error) {
	var allJobs []*models.Job

	// Search LinkedIn
	jobs, err := SearchLinkedIn(query, location)
	if err != nil {
		log.Printf("Error searching LinkedIn: %v", err)
	} else {
		allJobs = append(allJobs, jobs...)
	}

	// Search Indeed
	jobs, err = SearchIndeed(query, location)
	if err != nil {
		log.Printf("Error searching Indeed: %v", err)
	} else {
		allJobs = append(allJobs, jobs...)
	}

	// Search Greenhouse (company-specific, would need company list)
	// jobs, _ = SearchGreenhouse(query, location)
	// allJobs = append(allJobs, jobs...)

	// Search Lever (company-specific, would need company list)
	// jobs, _ = SearchLever(query, location)
	// allJobs = append(allJobs, jobs...)

	return allJobs, nil
}

// SearchJobs searches a specific job board
func SearchJobs(source, query, location string) ([]*models.Job, error) {
	switch strings.ToLower(source) {
	case "linkedin":
		return SearchLinkedIn(query, location)
	case "indeed":
		return SearchIndeed(query, location)
	case "greenhouse":
		return SearchGreenhouse(query, location)
	case "lever":
		return SearchLever(query, location)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}
}

// SearchLinkedIn searches LinkedIn jobs using browser automation
func SearchLinkedIn(query, location string) ([]*models.Job, error) {
	email := config.Get("linkedin_email")
	password := config.Get("linkedin_password")

	if email == "" || password == "" {
		return nil, fmt.Errorf("LinkedIn credentials not configured. Please set linkedin_email and linkedin_password in config")
	}

	ctx, cancel := createBrowserContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 2*pageLoadTimeout)
	defer cancel()

	var jobs []*models.Job
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.linkedin.com/login"),
		chromedp.WaitVisible(`input[name="session_key"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="session_key"]`, email, chromedp.ByQuery),
		chromedp.SendKeys(`input[name="session_password"]`, password, chromedp.ByQuery),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for login
		chromedp.Navigate(buildLinkedInSearchURL(query, location)),
		chromedp.Sleep(3*time.Second), // Wait for page load
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Scroll to load more jobs
			for i := 0; i < 3; i++ {
				chromedp.Evaluate("window.scrollTo(0, document.body.scrollHeight)", nil).Do(ctx)
				chromedp.Sleep(2 * time.Second).Do(ctx)
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Extract job listings
			var jobElements []map[string]string
			err := chromedp.Evaluate(`
				(() => {
					const jobs = [];
					const jobCards = document.querySelectorAll('.jobs-search-results__list-item, .job-card-container, [data-job-id]');
					
					jobCards.forEach((card, index) => {
						if (index >= 50) return; // Limit to 50 jobs
						
						const titleEl = card.querySelector('.job-card-list__title, .job-card-container__link, a[data-control-name="job_card_title"]');
						const companyEl = card.querySelector('.job-card-container__company-name, .job-card-container__primary-description, [data-control-name="job_card_company_link"]');
						const locationEl = card.querySelector('.job-card-container__metadata-item, .job-card-list__metadata-item');
						const linkEl = card.querySelector('a[href*="/jobs/view/"], a[href*="/jobs/collections/"]');
						
						if (titleEl) {
							const job = {
								title: titleEl.textContent.trim(),
								company: companyEl ? companyEl.textContent.trim() : '',
								location: locationEl ? locationEl.textContent.trim() : '',
								url: linkEl ? (linkEl.href.startsWith('http') ? linkEl.href : 'https://www.linkedin.com' + linkEl.getAttribute('href')) : ''
							};
							if (job.title) {
								jobs.push(job);
							}
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
				if job.Title != "" && job.Company != "" {
					jobs = append(jobs, job)
				}
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("LinkedIn scraping error: %w", err)
	}

	// Fetch detailed descriptions for each job
	for i, job := range jobs {
		if i >= maxJobsPerSource {
			break
		}
		if job.URL != "" {
			desc, err := fetchLinkedInJobDescription(ctx, job.URL)
			if err == nil {
				jobs[i].Description = desc
			}
			time.Sleep(rateLimitDelay) // Rate limiting
		}
	}

	return jobs, nil
}

// fetchLinkedInJobDescription fetches the full job description from a LinkedIn job URL
func fetchLinkedInJobDescription(ctx context.Context, url string) (string, error) {
	var description string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`.jobs-description-content__text, .show-more-less-html__markup`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to click "Show more" button if it exists
			chromedp.Click(`button[aria-label*="Show more"], button[aria-label*="see more"]`, chromedp.ByQuery).Do(ctx)
			chromedp.Sleep(1 * time.Second).Do(ctx)
			return nil
		}),
		chromedp.Text(`.jobs-description-content__text, .show-more-less-html__markup`, &description, chromedp.ByQuery),
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
	params = append(params, "f_E=2")       // Full-time (can be customized)
	
	if len(params) > 0 {
		return baseURL + "?" + strings.Join(params, "&")
	}
	return baseURL
}

// SearchIndeed searches Indeed jobs using browser automation
func SearchIndeed(query, location string) ([]*models.Job, error) {
	ctx, cancel := createBrowserContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, pageLoadTimeout)
	defer cancel()

	var jobs []*models.Job
	url := buildIndeedSearchURL(query, location)

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second), // Wait for page load
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Scroll to load more jobs
			for i := 0; i < 2; i++ {
				chromedp.Evaluate("window.scrollTo(0, document.body.scrollHeight)", nil).Do(ctx)
				chromedp.Sleep(2 * time.Second).Do(ctx)
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Extract job listings
			var jobElements []map[string]string
			err := chromedp.Evaluate(`
				(() => {
					const jobs = [];
					const jobCards = document.querySelectorAll('div[data-jk], [data-testid="job-card"], .job_seen_beacon');
					
					jobCards.forEach((card, index) => {
						if (index >= 50) return; // Limit to 50 jobs
						
						const titleEl = card.querySelector('h2.jobTitle a, a[data-testid="job-title"], .jobTitle a');
						const companyEl = card.querySelector('[data-testid="company-name"], .companyName, .company');
						const locationEl = card.querySelector('[data-testid="job-location"], .companyLocation, .location');
						const salaryEl = card.querySelector('[data-testid="attribute_snippet_testid"], .salary-snippet-container, .salaryText');
						const linkEl = card.querySelector('a[href*="/viewjob"], h2.jobTitle a, a[data-testid="job-title"]');
						
						if (titleEl) {
							const job = {
								title: titleEl.textContent.trim(),
								company: companyEl ? companyEl.textContent.trim() : '',
								location: locationEl ? locationEl.textContent.trim() : '',
								salary: salaryEl ? salaryEl.textContent.trim() : '',
								url: linkEl ? (linkEl.href.startsWith('http') ? linkEl.href : 'https://www.indeed.com' + linkEl.getAttribute('href')) : ''
							};
							if (job.title) {
								jobs.push(job);
							}
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
					Title:       jobData["title"],
					Company:     jobData["company"],
					Location:    jobData["location"],
					SalaryRange: jobData["salary"],
					URL:         jobData["url"],
					Source:      "indeed",
					ScrapedAt:   time.Now(),
					MatchScore:  0,
				}
				if job.Title != "" && job.Company != "" {
					jobs = append(jobs, job)
				}
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("Indeed scraping error: %w", err)
	}

	// Fetch detailed descriptions for a subset of jobs
	for i, job := range jobs {
		if i >= 20 { // Limit detailed fetches to avoid rate limiting
			break
		}
		if job.URL != "" {
			desc, err := fetchIndeedJobDescription(ctx, job.URL)
			if err == nil {
				jobs[i].Description = desc
			}
			time.Sleep(rateLimitDelay) // Rate limiting
		}
	}

	return jobs, nil
}

// fetchIndeedJobDescription fetches the full job description from an Indeed job URL
func fetchIndeedJobDescription(ctx context.Context, url string) (string, error) {
	var description string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.WaitVisible(`#jobDescriptionText, [data-testid="job-description"]`, chromedp.ByQuery),
		chromedp.Text(`#jobDescriptionText, [data-testid="job-description"]`, &description, chromedp.ByQuery),
	)
	return description, err
}

// buildIndeedSearchURL constructs Indeed job search URL
func buildIndeedSearchURL(query, location string) string {
	baseURL := "https://www.indeed.com/jobs"
	params := []string{}
	
	if query != "" {
		params = append(params, "q="+strings.ReplaceAll(query, " ", "+"))
	}
	if location != "" {
		params = append(params, "l="+strings.ReplaceAll(location, " ", "+"))
	}
	params = append(params, "fromage=1") // Last 24 hours
	
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
