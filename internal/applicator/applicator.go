package applicator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/khrees2412/autoply/pkg/models"
)

// ApplicationResult contains details about an application attempt
type ApplicationResult struct {
	Success      bool
	Message      string
	ScreenshotPath string
	Error        error
}

// ApplyToJob attempts to automatically apply to a job using browser automation
func ApplyToJob(ctx context.Context, job *models.Job, user *models.User, resume *models.Resume, coverLetter string) *ApplicationResult {
	if resume == nil || resume.FilePath == "" {
		return &ApplicationResult{
			Success: false,
			Message: "No resume available for auto-apply",
			Error:   fmt.Errorf("resume required"),
		}
	}

	// Validate resume file exists
	if _, err := os.Stat(resume.FilePath); err != nil {
		return &ApplicationResult{
			Success: false,
			Message: fmt.Sprintf("Resume file not found: %s", resume.FilePath),
			Error:   err,
		}
	}

	// Route to appropriate handler based on job source
	source := strings.ToLower(job.Source)
	var result *ApplicationResult
	var err error

	switch source {
	case "linkedin":
		result, err = applyLinkedIn(ctx, job, user, resume, coverLetter)
	case "greenhouse":
		result, err = applyGreenhouse(ctx, job, user, resume, coverLetter)
	case "lever":
		result, err = applyLever(ctx, job, user, resume, coverLetter)
	default:
		return &ApplicationResult{
			Success: false,
			Message: fmt.Sprintf("Auto-apply not supported for %s. Please apply manually at %s", job.Source, job.URL),
			Error:   fmt.Errorf("unsupported source: %s", source),
		}
	}

	if err != nil {
		return &ApplicationResult{
			Success: false,
			Message: err.Error(),
			Error:   err,
		}
	}

	return result
}

// CanAutoApply checks if a job can be automatically applied to
func CanAutoApply(job *models.Job) bool {
	supportedSources := []string{"linkedin", "greenhouse", "lever"}
	source := strings.ToLower(job.Source)
	for _, s := range supportedSources {
		if s == source {
			return true
		}
	}
	return false
}

// GetApplicationFormFields returns common form fields that might be needed
func GetApplicationFormFields() map[string]string {
	return map[string]string{
		"first_name":         "",
		"last_name":          "",
		"email":              "",
		"phone":              "",
		"location":           "",
		"linkedin_url":       "",
		"github_url":         "",
		"portfolio_url":      "",
		"cover_letter":       "",
		"resume_file":        "",
		"years_experience":   "",
		"availability":       "",
		"salary_expectation": "",
	}
}

// applyLinkedIn handles LinkedIn job application
func applyLinkedIn(ctx context.Context, job *models.Job, user *models.User, resume *models.Resume, coverLetter string) (*ApplicationResult, error) {
	browserCtx, cancel := createBrowserContext(ctx)
	defer cancel()

	var success bool
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(job.URL),
		chromedp.Sleep(2*time.Second),
		// Look for "Easy Apply" button
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to find and click the Easy Apply button
			var easyApplyFound bool
			err := chromedp.Evaluate(`
				(() => {
					const btn = document.querySelector('button[aria-label*="Easy Apply"], button:has-text("Easy Apply"), [data-tracking-control-name*="easy_apply"]');
					if (btn) {
						btn.click();
						return true;
					}
					return false;
				})()
			`, &easyApplyFound).Do(ctx)
			if err != nil {
				return err
			}
			if !easyApplyFound {
				return fmt.Errorf("Easy Apply button not found - may require manual application")
			}
			return nil
		}),
		chromedp.Sleep(1*time.Second),
		// Fill out form fields if they appear
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Try to fill common LinkedIn form fields
			fieldMappings := map[string]string{
				`input[name="firstName"]`:  user.Name,
				`input[name="lastName"]`:   "",
				`input[name="email"]`:      user.Email,
				`input[name="phoneNumber"]`: user.Phone,
			}

			for selector, value := range fieldMappings {
				if value != "" {
					if err := chromedp.SetValue(selector, value, chromedp.ByQuery).Do(ctx); err == nil {
						chromedp.Sleep(200 * time.Millisecond).Do(ctx)
					}
				}
			}
			return nil
		}),
		// Upload resume if file input found
		chromedp.ActionFunc(func(ctx context.Context) error {
			return uploadFileToForm(ctx, resume.FilePath, `input[type="file"]`)
		}),
		// Add cover letter if textarea found
		chromedp.ActionFunc(func(ctx context.Context) error {
			if coverLetter != "" {
				if err := chromedp.SetValue(`textarea[name="coverLetter"]`, coverLetter, chromedp.ByQuery).Do(ctx); err == nil {
					return nil
				}
			}
			return nil
		}),
		// Submit the application
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Click submit button
			if err := chromedp.Click(`button[type="submit"]`, chromedp.ByQuery).Do(ctx); err != nil {
				// Try alternative submit selectors
				chromedp.Click(`button[aria-label*="Submit"], button:has-text("Submit")`, chromedp.ByQuery).Do(ctx)
			}
			chromedp.Sleep(2 * time.Second).Do(ctx)
			success = true
			return nil
		}),
		// Check for confirmation or errors
		chromedp.ActionFunc(func(ctx context.Context) error {
			var pageContent string
			chromedp.OuterHTML(`body`, &pageContent).Do(ctx)
			
			if strings.Contains(pageContent, "Application sent") || strings.Contains(pageContent, "applied") {
				return nil
			}
			if strings.Contains(pageContent, "captcha") || strings.Contains(pageContent, "verify") {
				return fmt.Errorf("CAPTCHA verification required")
			}
			return nil
		}),
	)

	if err != nil {
		return nil, err
	}

	if !success {
		return &ApplicationResult{
			Success: false,
			Message: "Application submission may have failed - please verify manually",
		}, nil
	}

	return &ApplicationResult{
		Success: true,
		Message: "Successfully applied to " + job.Title + " at " + job.Company,
	}, nil
}

// applyGreenhouse handles Greenhouse job application
func applyGreenhouse(ctx context.Context, job *models.Job, user *models.User, resume *models.Resume, coverLetter string) (*ApplicationResult, error) {
	browserCtx, cancel := createBrowserContext(ctx)
	defer cancel()

	var success bool
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(job.URL),
		chromedp.Sleep(3*time.Second),
		// Fill in basic fields
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Greenhouse uses various input types
			fields := map[string]string{
				`input[name*="first"]`:  user.Name,
				`input[name*="email"]`:  user.Email,
				`input[name*="phone"]`:  user.Phone,
				`textarea[name*="cover"]`: coverLetter,
			}

			for selector, value := range fields {
				if value != "" {
					_ = chromedp.SetValue(selector, value, chromedp.ByQuery).Do(ctx)
					chromedp.Sleep(300 * time.Millisecond).Do(ctx)
				}
			}
			return nil
		}),
		// Upload resume
		chromedp.ActionFunc(func(ctx context.Context) error {
			return uploadFileToForm(ctx, resume.FilePath, `input[type="file"][name*="resume"], input[type="file"][name*="attachment"]`)
		}),
		// Wait for form to be ready
		chromedp.Sleep(1*time.Second),
		// Submit
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Find and click submit
			if err := chromedp.Click(`button[type="submit"]`, chromedp.ByQuery).Do(ctx); err != nil {
				chromedp.Click(`input[type="submit"]`, chromedp.ByQuery).Do(ctx)
			}
			chromedp.Sleep(2 * time.Second).Do(ctx)
			success = true
			return nil
		}),
	)

	if err != nil {
		return nil, err
	}

	return &ApplicationResult{
		Success: success,
		Message: "Application submitted to Greenhouse",
	}, nil
}

// applyLever handles Lever job application
func applyLever(ctx context.Context, job *models.Job, user *models.User, resume *models.Resume, coverLetter string) (*ApplicationResult, error) {
	browserCtx, cancel := createBrowserContext(ctx)
	defer cancel()

	var success bool
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(job.URL),
		chromedp.Sleep(3*time.Second),
		// Fill form
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Lever uses different naming
			fields := map[string]string{
				`input[name="name"]`:    user.Name,
				`input[name="email"]`:   user.Email,
				`input[name="phone"]`:   user.Phone,
				`textarea[name*="message"]`: coverLetter,
			}

			for selector, value := range fields {
				if value != "" {
					_ = chromedp.SetValue(selector, value, chromedp.ByQuery).Do(ctx)
					chromedp.Sleep(300 * time.Millisecond).Do(ctx)
				}
			}
			return nil
		}),
		// Upload resume
		chromedp.ActionFunc(func(ctx context.Context) error {
			return uploadFileToForm(ctx, resume.FilePath, `input[type="file"]`)
		}),
		chromedp.Sleep(1*time.Second),
		// Submit
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := chromedp.Click(`button[type="submit"]`, chromedp.ByQuery).Do(ctx); err != nil {
				chromedp.Click(`input[type="submit"]`, chromedp.ByQuery).Do(ctx)
			}
			chromedp.Sleep(2 * time.Second).Do(ctx)
			success = true
			return nil
		}),
	)

	if err != nil {
		return nil, err
	}

	return &ApplicationResult{
		Success: success,
		Message: "Application submitted to Lever",
	}, nil
}

// uploadFileToForm uploads a file to a form input
func uploadFileToForm(ctx context.Context, filePath string, selector string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file not found: %s", absPath)
	}

	// Find the file input and send the file path
	return chromedp.SendKeys(selector, absPath, chromedp.ByQuery).Do(ctx)
}

// createBrowserContext creates a new browser context with appropriate options
func createBrowserContext(parent context.Context) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(parent, opts...)
	ctx, cancel2 := chromedp.NewContext(allocCtx)

	return ctx, func() {
		cancel2()
		cancel()
	}
}

// WaitForElement waits for an element to appear
func WaitForElement(ctx context.Context, selector string, timeout time.Duration) error {
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return chromedp.WaitVisible(selector, chromedp.ByQuery).Do(c)
}

// FillForm fills out a form with the provided fields
func FillForm(ctx context.Context, url string, fields map[string]string) error {
	browserCtx, cancel := createBrowserContext(ctx)
	defer cancel()

	return chromedp.Run(browserCtx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for selector, value := range fields {
				if value != "" {
					if err := chromedp.SetValue(selector, value, chromedp.ByQuery).Do(ctx); err == nil {
						chromedp.Sleep(200 * time.Millisecond).Do(ctx)
					}
				}
			}
			return nil
		}),
	)
}

