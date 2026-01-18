package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
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
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		title, _ := cmd.Flags().GetString("title")
		company, _ := cmd.Flags().GetString("company")
		location, _ := cmd.Flags().GetString("location")
		description, _ := cmd.Flags().GetString("description")

		if url != "" {
			// Try to parse job from URL
			fmt.Printf("Fetching job details from %s...\n", url)
			jobData, err := parseJobFromURL(url)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing job URL: %v\n", err)
				fmt.Println("You can manually provide job details using --title, --company, etc.")
				return
			}
			
			jobData.URL = url
			if err := database.CreateJob(jobData); err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint failed") {
					fmt.Println("This job has already been added.")
					return
				}
				fmt.Fprintf(os.Stderr, "Error saving job: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Job added: %s at %s (ID: %d)\n", jobData.Title, jobData.Company, jobData.ID)
			return
		}

		// Manual entry
		if title == "" || company == "" {
			fmt.Println("Either --url or both --title and --company are required")
			return
		}

		job := &models.Job{
			Title:       title,
			Company:     company,
			Location:    location,
			Description: description,
			Source:      "manual",
		}

		if err := database.CreateJob(job); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving job: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Job added: %s at %s (ID: %d)\n", job.Title, job.Company, job.ID)
	},
}

var listJobsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved jobs",
	Run: func(cmd *cobra.Command, args []string) {
		jobs, err := database.GetAllJobs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching jobs: %v\n", err)
			os.Exit(1)
		}

		if len(jobs) == 0 {
			fmt.Println("No jobs found. Add jobs with 'autoply job add --url URL'")
			return
		}

		fmt.Println(titleStyle.Render("Saved Jobs"))
		for i, job := range jobs {
			fmt.Printf("\n%s. %s\n", labelStyle.Render(fmt.Sprintf("%d", i+1)), job.Title)
			fmt.Printf("   %s %s\n", labelStyle.Render("Company:"), job.Company)
			if job.Location != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("Location:"), job.Location)
			}
			fmt.Printf("   %s %d\n", labelStyle.Render("ID:"), job.ID)
			if job.URL != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("URL:"), job.URL)
			}
			fmt.Printf("   %s %s\n", labelStyle.Render("Added:"), job.ScrapedAt.Format("Jan 2, 2006"))
		}
	},
}

var showJobCmd = &cobra.Command{
	Use:   "show <job-id>",
	Short: "Show details of a specific job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			fmt.Println("Invalid job ID. Must be a number.")
			return
		}

		job, err := database.GetJob(jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching job: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(titleStyle.Render(job.Title))
		fmt.Printf("%s %s\n", labelStyle.Render("Company:"), job.Company)
		if job.Location != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("Location:"), job.Location)
		}
		if job.SalaryRange != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("Salary:"), job.SalaryRange)
		}
		if job.URL != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("URL:"), job.URL)
		}
		fmt.Printf("%s %s\n", labelStyle.Render("Source:"), job.Source)
		fmt.Printf("%s %s\n", labelStyle.Render("Added:"), job.ScrapedAt.Format("Jan 2, 2006 15:04"))

		if job.Description != "" {
			fmt.Println(labelStyle.Render("\nDescription:"))
			fmt.Println(job.Description)
		}

		// Check if already applied
		app, _ := database.GetApplicationByJobID(jobID)
		if app != nil {
			fmt.Printf("\n%s %s\n", labelStyle.Render("Application Status:"), app.Status)
			fmt.Printf("%s %s\n", labelStyle.Render("Applied At:"), app.AppliedAt.Format("Jan 2, 2006"))
		}
	},
}

var removeJobCmd = &cobra.Command{
	Use:   "remove <job-id>",
	Short: "Remove a job posting",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			fmt.Println("Invalid job ID. Must be a number.")
			return
		}

		// Check if job exists
		job, err := database.GetJob(jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Job not found\n")
			return
		}

		if err := database.DeleteJob(jobID); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing job: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Removed job: %s at %s\n", job.Title, job.Company)
	},
}

// parseJobFromURL attempts to extract job information from a URL
func parseJobFromURL(url string) (*models.Job, error) {
	// Fetch the page
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	// Basic parsing (this is simplified - real implementation would be more robust)
	job := &models.Job{
		URL:    url,
		Source: "manual",
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
				job.Company = strings.Title(parts[i+1])
				break
			}
		}
	} else if strings.Contains(url, "lever.co") {
		parts := strings.Split(url, "/")
		if len(parts) > 2 {
			company := strings.Split(parts[2], ".")[0]
			job.Company = strings.Title(company)
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
				job.Company = strings.Title(parts[0])
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
