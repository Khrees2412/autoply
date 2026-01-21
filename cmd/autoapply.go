package cmd

import (
	"fmt"

	"github.com/khrees2412/autoply/internal/applicator"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var autoApplyCmd = &cobra.Command{
	Use:   "auto-apply",
	Short: "Manage automatic job applications",
	Long:  "Apply to jobs automatically using browser automation. Requires resume and user profile setup.",
}

var autoApplyTestCmd = &cobra.Command{
	Use:   "test <job-id>",
	Short: "Test auto-apply on a single job",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply auto-apply test 5`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			return fmt.Errorf("invalid job ID: must be a number")
		}

		// Verify prerequisites
		job, err := database.GetJob(jobID)
		if err != nil {
			return fmt.Errorf("job not found")
		}

		user, err := database.GetUser()
		if err != nil || user == nil {
			return fmt.Errorf("user profile not configured. Run 'autoply profile setup'")
		}

		resume, err := database.GetDefaultResume()
		if err != nil || resume == nil {
			return fmt.Errorf("no default resume set. Run 'autoply resume list' and 'autoply resume set-default <id>'")
		}

		// Display what will be auto-applied
		fmt.Println("\n📋 Auto-Apply Test")
		fmt.Println("==================")
		fmt.Printf("Job: %s at %s\n", job.Title, job.Company)
		fmt.Printf("Source: %s\n", job.Source)
		fmt.Printf("URL: %s\n", job.URL)
		fmt.Printf("\nProfile: %s <%s>\n", user.Name, user.Email)
		fmt.Printf("Resume: %s\n", resume.Name)

		// Check if source is supported
		if !applicator.CanAutoApply(job) {
			fmt.Printf("\n❌ Auto-apply not supported for %s\n", job.Source)
			fmt.Printf("Supported sources: linkedin, greenhouse, lever\n")
			return nil
		}

		fmt.Println("\n✅ Auto-apply is supported for this job!")
		fmt.Println("\nPress Ctrl+C to cancel, or the application will proceed in 5 seconds...")

		// Get cover letter if available
		coverLetter, _ := database.GetCoverLetterByJobID(jobID)
		var clContent string
		if coverLetter != nil {
			clContent = coverLetter.Content
			fmt.Println("\n📝 Using generated cover letter")
		}

		// Test the application
		fmt.Println("\n⏳ Starting browser automation...")
		result := applicator.ApplyToJob(cmd.Context(), job, user, resume, clContent)

		if !result.Success {
			fmt.Printf("\n❌ Auto-apply failed: %s\n", result.Message)
			if result.Error != nil {
				fmt.Printf("Details: %v\n", result.Error)
			}
			return nil
		}

		fmt.Printf("\n✅ %s\n", result.Message)
		fmt.Println("\nTo create the application record, run:")
		fmt.Printf("  autoply apply %d --auto\n", jobID)

		return nil
	},
}

var autoApplySupportedCmd = &cobra.Command{
	Use:   "supported",
	Short: "Show supported job sources for auto-apply",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Supported job sources for auto-apply:")
		fmt.Println("  • linkedin    - LinkedIn Easy Apply")
		fmt.Println("  • greenhouse  - Greenhouse ATS")
		fmt.Println("  • lever       - Lever ATS")
		fmt.Println("\nOther sources require manual application.")
		return nil
	},
}

var autoApplyBulkCmd = &cobra.Command{
	Use:   "bulk <filter>",
	Short: "Auto-apply to multiple matching jobs",
	Long:  "Auto-apply to jobs matching criteria (all, supported, or match score)",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply auto-apply bulk all
   autoply auto-apply bulk supported
   autoply auto-apply bulk --score 0.7`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := args[0]
		scoreThreshold, _ := cmd.Flags().GetFloat64("score")

		// Get all jobs
		jobs, err := database.GetAllJobs()
		if err != nil {
			return fmt.Errorf("failed to fetch jobs: %w", err)
		}

		// Filter jobs
		var filteredJobs []*models.Job
		for _, job := range jobs {
			// Skip if already applied
			existing, _ := database.GetApplicationByJobID(job.ID)
			if existing != nil {
				continue
			}

			switch filter {
			case "all":
				filteredJobs = append(filteredJobs, job)
			case "supported":
				if applicator.CanAutoApply(job) {
					filteredJobs = append(filteredJobs, job)
				}
			default:
				if job.MatchScore >= scoreThreshold {
					filteredJobs = append(filteredJobs, job)
				}
			}
		}

		if len(filteredJobs) == 0 {
			fmt.Printf("No jobs found matching filter '%s'\n", filter)
			return nil
		}

		fmt.Printf("Found %d jobs matching filter '%s'\n", len(filteredJobs), filter)
		fmt.Println("Supported for auto-apply:")

		supportedCount := 0
		unsupportedCount := 0

		for _, job := range filteredJobs {
			if applicator.CanAutoApply(job) {
				fmt.Printf("  ✓ %s at %s (%s)\n", job.Title, job.Company, job.Source)
				supportedCount++
			} else {
				unsupportedCount++
			}
		}

		if unsupportedCount > 0 {
			fmt.Printf("\nUnsupported for auto-apply: %d jobs\n", unsupportedCount)
		}

		if supportedCount == 0 {
			fmt.Println("\nNo supported jobs to auto-apply to.")
			return nil
		}

		fmt.Printf("\nReady to auto-apply to %d jobs. Run:\n", supportedCount)
		fmt.Println("  autoply apply --batch jobs.txt --auto")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(autoApplyCmd)
	autoApplyCmd.AddCommand(autoApplyTestCmd)
	autoApplyCmd.AddCommand(autoApplySupportedCmd)
	autoApplyCmd.AddCommand(autoApplyBulkCmd)

	// Flags for bulk command
	autoApplyBulkCmd.Flags().Float64("score", 0.0, "Match score threshold (0.0-1.0)")
}
