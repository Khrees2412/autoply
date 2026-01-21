package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/khrees2412/autoply/internal/app"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage job postings",
	Long:  "Add, list, view, and remove job postings",
}

var addJobCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a job posting",
	Example: `  autoply job add --url https://company.com/jobs/123
  autoply job add --title "Software Engineer" --company "Acme Inc" --location "Remote"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		application := app.GetAppFromContext(ctx)
		if application == nil {
			return fmt.Errorf("application not initialized")
		}

		url, _ := cmd.Flags().GetString("url")
		title, _ := cmd.Flags().GetString("title")
		company, _ := cmd.Flags().GetString("company")
		location, _ := cmd.Flags().GetString("location")
		description, _ := cmd.Flags().GetString("description")

		if url != "" {
			// Try to parse job from URL
			cmd.Printf("Fetching job details from %s...\n", url)
			jobData, err := parseJobFromURL(ctx, application.HTTPClient, url)
			if err != nil {
				cmd.Printf("Warning: could not parse job URL: %v\n", err)
				cmd.Println("You can manually provide job details using --title, --company, etc.")
				if title == "" || company == "" {
					return fmt.Errorf("job title and company required when URL parsing fails")
				}
			} else {
				jobData.URL = url
				if err := database.CreateJob(jobData); err != nil {
					if strings.Contains(err.Error(), "UNIQUE constraint failed") {
						cmd.Println("This job has already been added.")
						return nil
					}
					return fmt.Errorf("save job: %w", err)
				}
				cmd.Printf("✓ Job added: %s at %s (ID: %d)\n", jobData.Title, jobData.Company, jobData.ID)
				return nil
			}
		}

		// Manual entry
		if title == "" || company == "" {
			return fmt.Errorf("either --url or both --title and --company are required")
		}

		job := &models.Job{
			Title:       title,
			Company:     company,
			Location:    location,
			Description: description,
			Source:      "manual",
		}

		if err := database.CreateJob(job); err != nil {
			return fmt.Errorf("save job: %w", err)
		}

		cmd.Printf("✓ Job added: %s at %s (ID: %d)\n", job.Title, job.Company, job.ID)
		return nil
	},
}

var listJobsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		jobs, err := database.GetAllJobs()
		if err != nil {
			return fmt.Errorf("fetch jobs: %w", err)
		}

		if len(jobs) == 0 {
			cmd.Println("No jobs found. Add jobs with 'autoply job add --url URL'")
			return nil
		}

		cmd.Println(titleStyle.Render("Saved Jobs"))
		for i, job := range jobs {
			cmd.Printf("\n%s. %s\n", labelStyle.Render(fmt.Sprintf("%d", i+1)), job.Title)
			cmd.Printf("   %s %s\n", labelStyle.Render("Company:"), job.Company)
			if job.Location != "" {
				cmd.Printf("   %s %s\n", labelStyle.Render("Location:"), job.Location)
			}
			cmd.Printf("   %s %d\n", labelStyle.Render("ID:"), job.ID)
			if job.URL != "" {
				cmd.Printf("   %s %s\n", labelStyle.Render("URL:"), job.URL)
			}
			cmd.Printf("   %s %s\n", labelStyle.Render("Added:"), job.ScrapedAt.Format("Jan 2, 2006"))
		}
		return nil
	},
}

var showJobCmd = &cobra.Command{
	Use:   "show <job-id>",
	Short: "Show details of a specific job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			return fmt.Errorf("invalid job ID: must be a number")
		}

		job, err := database.GetJob(jobID)
		if err != nil {
			return fmt.Errorf("fetch job: %w", err)
		}

		cmd.Println(titleStyle.Render(job.Title))
		cmd.Printf("%s %s\n", labelStyle.Render("Company:"), job.Company)
		if job.Location != "" {
			cmd.Printf("%s %s\n", labelStyle.Render("Location:"), job.Location)
		}
		if job.SalaryRange != "" {
			cmd.Printf("%s %s\n", labelStyle.Render("Salary:"), job.SalaryRange)
		}
		if job.URL != "" {
			cmd.Printf("%s %s\n", labelStyle.Render("URL:"), job.URL)
		}
		cmd.Printf("%s %s\n", labelStyle.Render("Source:"), job.Source)
		cmd.Printf("%s %s\n", labelStyle.Render("Added:"), job.ScrapedAt.Format("Jan 2, 2006 15:04"))

		if job.Description != "" {
			cmd.Println(labelStyle.Render("\nDescription:"))
			cmd.Println(job.Description)
		}

		// Check if already applied
		application, _ := database.GetApplicationByJobID(jobID)
		if application != nil {
			cmd.Printf("\n%s %s\n", labelStyle.Render("Application Status:"), application.Status)
			cmd.Printf("%s %s\n", labelStyle.Render("Applied At:"), application.AppliedAt.Format("Jan 2, 2006"))
		}
		return nil
	},
}

var removeJobCmd = &cobra.Command{
	Use:   "remove <job-id>",
	Short: "Remove a job posting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			return fmt.Errorf("invalid job ID: must be a number")
		}

		// Check if job exists
		job, err := database.GetJob(jobID)
		if err != nil {
			return fmt.Errorf("job not found")
		}

		if err := database.DeleteJob(jobID); err != nil {
			return fmt.Errorf("remove job: %w", err)
		}

		cmd.Printf("✓ Removed job: %s at %s\n", job.Title, job.Company)
		return nil
	},
}

// parseJobFromURL attempts to extract job information from a URL
func parseJobFromURL(ctx context.Context, client *http.Client, url string) (*models.Job, error) {
	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set a proper user-agent (some sites block default Go UA)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Autoply/1.0)")

	// Fetch the page
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Limit body size to 2MB to avoid huge downloads
	limitedBody := io.LimitReader(resp.Body, 2<<20)
	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	html := string(body)

	// Basic parsing (this is simplified - real implementation would be more robust)
	job := &models.Job{
		URL:    url,
		Source: "url",
	}

	// Try to extract title
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	if match := titleRegex.FindStringSubmatch(html); len(match) > 1 {
		job.Title = strings.TrimSpace(match[1])
		// Clean up common title suffixes
		job.Title = strings.Split(job.Title, " - ")[0]
		job.Title = strings.Split(job.Title, " | ")[0]
	}

	// Try to extract company from URL or page
	if strings.Contains(url, "greenhouse.io") {
		parts := strings.Split(url, "/")
		for i, part := range parts {
			if part == "boards" && i+1 < len(parts) {
				job.Company = titleCase(parts[i+1])
				break
			}
		}
	} else if strings.Contains(url, "lever.co") {
		parts := strings.Split(url, "/")
		if len(parts) > 2 {
			company := strings.Split(parts[2], ".")[0]
			job.Company = titleCase(company)
		}
	}

	// If we couldn't extract company, use domain
	if job.Company == "" {
		domainRegex := regexp.MustCompile(`https?://([^/]+)`)
		if match := domainRegex.FindStringSubmatch(url); len(match) > 1 {
			domain := match[1]
			domain = strings.TrimPrefix(domain, "www.")
			parts := strings.Split(domain, ".")
			if len(parts) > 0 {
				job.Company = titleCase(parts[0])
			}
		}
	}

	// Extract meta description for job description
	descRegex := regexp.MustCompile(`<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	if match := descRegex.FindStringSubmatch(html); len(match) > 1 {
		job.Description = strings.TrimSpace(match[1])
	}

	if job.Title == "" {
		return nil, fmt.Errorf("could not extract job title from URL")
	}

	return job, nil
}

// titleCase converts a string to title case using proper locale-aware capitalization
func titleCase(s string) string {
	return cases.Title(language.English).String(s)
}

func init() {
	rootCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(addJobCmd)
	jobCmd.AddCommand(listJobsCmd)
	jobCmd.AddCommand(showJobCmd)
	jobCmd.AddCommand(removeJobCmd)

	// Flags for add command
	addJobCmd.Flags().String("url", "", "Job posting URL")
	addJobCmd.Flags().String("title", "", "Job title")
	addJobCmd.Flags().String("company", "", "Company name")
	addJobCmd.Flags().String("location", "", "Job location")
	addJobCmd.Flags().String("description", "", "Job description")
}
