package cmd

import (
	"fmt"
	"os"

	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/internal/matcher"
	"github.com/khrees2412/autoply/internal/scraper"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for jobs",
	Long:  "Search for jobs from various job boards and filter by criteria",
	Example: `  autoply search --query "software engineer" --location "remote"
  autoply search --query "backend" --auto-match
  autoply search --query "frontend" --save-query "frontend-jobs"`,
	Run: func(cmd *cobra.Command, args []string) {
		query, _ := cmd.Flags().GetString("query")
		location, _ := cmd.Flags().GetString("location")
		autoMatch, _ := cmd.Flags().GetBool("auto-match")
		saveQuery, _ := cmd.Flags().GetString("save-query")
		source, _ := cmd.Flags().GetString("source")

		if query == "" {
			fmt.Println("Query is required. Use --query flag")
			return
		}

		fmt.Printf("Searching for jobs: '%s'", query)
		if location != "" {
			fmt.Printf(" in %s", location)
		}
		fmt.Println()

		// Search jobs using scraper
		var jobs []*models.Job
		var err error

		if source != "" {
			// Search specific source
			jobs, err = scraper.SearchJobs(source, query, location)
		} else {
			// Search all sources
			jobs, err = scraper.SearchAllSources(query, location)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching jobs: %v\n", err)
			os.Exit(1)
		}

		if len(jobs) == 0 {
			fmt.Println("No jobs found matching your criteria.")
			return
		}

		// Calculate match scores if auto-match is enabled
		if autoMatch {
			user, err := database.GetUser()
			if err == nil && user != nil {
				skills, _ := database.GetUserSkills(user.ID)
				experiences, _ := database.GetUserExperiences(user.ID)
				for _, job := range jobs {
					score := matcher.CalculateMatchScore(job, user, skills, experiences)
					job.MatchScore = score
				}
			}
			// Filter to only high-scoring jobs (>= 0.7)
			filtered := []*models.Job{}
			for _, job := range jobs {
				if job.MatchScore >= 0.7 {
					filtered = append(filtered, job)
				}
			}
			jobs = filtered
		}

		// Save jobs to database
		savedCount := 0
		for _, job := range jobs {
			// Check if job already exists
			existing, _ := database.GetJobByURL(job.URL)
			if existing != nil {
				continue
			}
			if err := database.CreateJob(job); err != nil {
				continue // Skip duplicates
			}
			savedCount++
		}

		// Display results
		fmt.Println(titleStyle.Render(fmt.Sprintf("Found %d jobs", len(jobs))))
		for i, job := range jobs {
			fmt.Printf("\n%d. %s\n", i+1, job.Title)
			fmt.Printf("   %s %s\n", labelStyle.Render("Company:"), job.Company)
			if job.Location != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("Location:"), job.Location)
			}
			if job.SalaryRange != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("Salary:"), job.SalaryRange)
			}
			if job.URL != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("URL:"), job.URL)
			}
			if job.MatchScore > 0 {
				fmt.Printf("   %s %.1f%%\n", labelStyle.Render("Match:"), job.MatchScore*100)
			}
		}

		if savedCount > 0 {
			fmt.Printf("\n✓ Saved %d new jobs to database\n", savedCount)
		}

		// Save search query if requested
		if saveQuery != "" {
			if err := database.SaveSearchQuery(saveQuery, query, location, source); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not save search query: %v\n", err)
			} else {
				fmt.Printf("✓ Saved search query as: %s\n", saveQuery)
			}
		}
	},
}

var recommendCmd = &cobra.Command{
	Use:   "recommend",
	Short: "Get AI job recommendations",
	Long:  "Get personalized job recommendations based on your profile",
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create your profile first.")
			return
		}

		skills, _ := database.GetUserSkills(user.ID)
		experiences, _ := database.GetUserExperiences(user.ID)

		// Get all jobs and score them
		jobs, err := database.GetAllJobs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching jobs: %v\n", err)
			os.Exit(1)
		}

		if len(jobs) == 0 {
			fmt.Println("No jobs found. Search for jobs first with 'autoply search'")
			return
		}

		// Calculate match scores
		for _, job := range jobs {
			score := matcher.CalculateMatchScore(job, user, skills, experiences)
			job.MatchScore = score
		}

		// Sort by match score (would need to implement sorting)
		// For now, just filter high-scoring jobs
		recommended := []*models.Job{}
		for _, job := range jobs {
			if job.MatchScore >= 0.7 {
				recommended = append(recommended, job)
			}
		}

		if len(recommended) == 0 {
			fmt.Println("No highly matching jobs found. Try searching for more jobs.")
			return
		}

		fmt.Println(titleStyle.Render("Recommended Jobs"))
		for i, job := range recommended {
			if i >= 10 { // Limit to top 10
				break
			}
			fmt.Printf("\n%d. %s at %s\n", i+1, job.Title, job.Company)
			fmt.Printf("   %s %.1f%%\n", labelStyle.Render("Match Score:"), job.MatchScore*100)
			if job.Location != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("Location:"), job.Location)
			}
			if job.URL != "" {
				fmt.Printf("   %s %s\n", labelStyle.Render("URL:"), job.URL)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(recommendCmd)

	searchCmd.Flags().String("query", "", "Search query (required)")
	searchCmd.Flags().String("location", "", "Job location")
	searchCmd.Flags().Bool("auto-match", false, "Only show high-matching jobs")
	searchCmd.Flags().String("save-query", "", "Save this search query with a name")
	searchCmd.Flags().String("source", "", "Job board source (linkedin, startup.jobs, greenhouse, lever)")
}
