package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	Long:  "Launch the interactive terminal user interface for browsing jobs and managing applications",
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
}

func runTUI() {
	jobs, err := database.GetAllJobs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching jobs: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Println("No jobs found. Add jobs with 'autoply job add' or 'autoply search'")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Display job list
		fmt.Println(titleStyle.Render("Job Browser"))
		fmt.Println("Press 'q' to quit, or enter a job number to view details")
		fmt.Println()

		for i, job := range jobs {
			fmt.Printf("%d. %s at %s\n", i+1, job.Title, job.Company)
		}

		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "q" || input == "Q" {
			break
		}

		jobNum, err := strconv.Atoi(input)
		if err != nil || jobNum < 1 || jobNum > len(jobs) {
			fmt.Println("Invalid selection")
			continue
		}

		job := jobs[jobNum-1]
		displayJobDetails(job, reader)
	}
}

func displayJobDetails(job *models.Job, reader *bufio.Reader) {
	for {
		fmt.Println("\n" + strings.Repeat("=", 60))
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
		if job.MatchScore > 0 {
			fmt.Printf("%s %.1f%%\n", labelStyle.Render("Match Score:"), job.MatchScore*100)
		}

		if job.Description != "" {
			fmt.Println(labelStyle.Render("\nDescription:"))
			fmt.Println(job.Description)
		}

		// Check application status
		app, _ := database.GetApplicationByJobID(job.ID)
		if app != nil {
			fmt.Printf("\n%s %s\n", labelStyle.Render("Application Status:"), app.Status)
		}

		fmt.Println("\nOptions:")
		fmt.Println("  [a] Apply to this job")
		fmt.Println("  [g] Generate cover letter")
		fmt.Println("  [b] Back to list")
		fmt.Print("\n> ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "a":
			// Apply to job
			app := &models.Application{
				JobID:  job.ID,
				Status: "applied",
			}
			resume, _ := database.GetDefaultResume()
			if resume != nil {
				id := resume.ID
				app.ResumeID = &id
			}
			if err := database.CreateApplication(app); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("✓ Application created!")
			}
			return
		case "g":
			fmt.Println("Generating cover letter...")
			// This would call the generate command logic
			return
		case "b":
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

