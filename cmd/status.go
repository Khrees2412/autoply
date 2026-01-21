package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/khrees2412/autoply/internal/applicator"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "View application status",
	Long:  "View and manage your job application statuses",
	Run: func(cmd *cobra.Command, args []string) {
		filterStatus, _ := cmd.Flags().GetString("filter")

		apps, err := database.GetApplicationsWithJobs()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching applications: %v\n", err)
			os.Exit(1)
		}

		if len(apps) == 0 {
			fmt.Println("No applications yet. Apply to jobs with 'autoply apply <job-id>'")
			return
		}

		// Filter if requested
		filtered := []map[string]interface{}{}
		for _, app := range apps {
			status := app["status"].(string)
			if filterStatus == "" || status == filterStatus {
				filtered = append(filtered, app)
			}
		}

		if len(filtered) == 0 {
			fmt.Printf("No applications with status '%s'\n", filterStatus)
			return
		}

		fmt.Println(titleStyle.Render("Your Applications"))
		
		// Group by status
		statusGroups := map[string][]map[string]interface{}{
			"pending":   {},
			"applied":   {},
			"interview": {},
			"offer":     {},
			"rejected":  {},
		}

		for _, app := range filtered {
			status := app["status"].(string)
			statusGroups[status] = append(statusGroups[status], app)
		}

		// Display each group
		for _, status := range []string{"pending", "applied", "interview", "offer", "rejected"} {
			apps := statusGroups[status]
			if len(apps) == 0 {
				continue
			}

			fmt.Printf("\n%s (%d)\n", labelStyle.Render(getStatusLabel(status)), len(apps))
			for _, app := range apps {
				fmt.Printf("  • %s at %s\n", app["title"], app["company"])
				fmt.Printf("    %s %d | Applied: %s\n", 
					labelStyle.Render("ID:"), 
					app["job_id"], 
					app["applied_at"])
				if notes, ok := app["notes"].(string); ok && notes != "" {
					fmt.Printf("    %s %s\n", labelStyle.Render("Notes:"), notes)
				}
			}
		}

		fmt.Printf("\n%s %d\n", labelStyle.Render("Total Applications:"), len(filtered))
	},
}

var updateStatusCmd = &cobra.Command{
	Use:   "update <job-id>",
	Short: "Update application status",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply status update 1 --status interview
  autoply status update 5 --status rejected --notes "Not a good fit"`,
	Run: func(cmd *cobra.Command, args []string) {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			fmt.Println("Invalid job ID. Must be a number.")
			return
		}

		newStatus, _ := cmd.Flags().GetString("status")
		notes, _ := cmd.Flags().GetString("notes")

		if newStatus == "" {
			fmt.Println("Status is required. Use --status flag")
			return
		}

		// Validate status
		validStatuses := []string{"pending", "applied", "interview", "offer", "rejected"}
		valid := false
		for _, s := range validStatuses {
			if s == newStatus {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Invalid status. Must be one of: %v\n", validStatuses)
			return
		}

		// Check if application exists
		app, err := database.GetApplicationByJobID(jobID)
		if err != nil || app == nil {
			fmt.Println("No application found for this job. Create one with 'autoply apply <job-id>'")
			return
		}

		// Update status
		if err := database.UpdateApplicationStatus(app.ID, newStatus, notes); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating status: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Application status updated to: %s\n", newStatus)
		if notes != "" {
			fmt.Printf("  Notes: %s\n", notes)
		}
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply <job-id>",
	Short: "Apply to a job (manually or automatically)",
	Args:  cobra.MinimumNArgs(1),
	Example: `  autoply apply 1
   autoply apply 5 --notes "Applied via LinkedIn"
   autoply apply 5 --auto
   autoply apply --batch job-ids.txt --auto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		batchFile, _ := cmd.Flags().GetString("batch")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		autoApply, _ := cmd.Flags().GetBool("auto")

		// Handle batch operations
		if batchFile != "" {
			return handleBatchApply(cmd.Context(), batchFile, dryRun, autoApply)
		}

		// Single job application
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			return fmt.Errorf("invalid job ID: must be a number")
		}

		notes, _ := cmd.Flags().GetString("notes")

		// Check if job exists
		job, err := database.GetJob(jobID)
		if err != nil {
			return fmt.Errorf("job not found")
		}

		// Check if already applied
		existing, _ := database.GetApplicationByJobID(jobID)
		if existing != nil {
			fmt.Printf("Already applied to this job (Status: %s)\n", existing.Status)
			fmt.Println("Use 'autoply status update <job-id>' to change status")
			return nil
		}

		// Get user profile
		user, err := database.GetUser()
		if err != nil || user == nil {
			return fmt.Errorf("user profile not configured. Run 'autoply profile setup'")
		}

		// Get default resume
		resume, _ := database.GetDefaultResume()
		if resume == nil {
			return fmt.Errorf("no default resume set. Run 'autoply resume list' and 'autoply resume set-default <id>'")
		}

		// Check for generated cover letter
		coverLetter, _ := database.GetCoverLetterByJobID(jobID)
		var clContent string
		if coverLetter != nil {
			clContent = coverLetter.Content
		}

		// Attempt auto-apply if requested
		if autoApply {
			return applyJobAuto(cmd.Context(), job, user, resume, clContent, jobID, notes)
		}

		// Manual apply - just mark as applied
		return applyJobManual(jobID, resume, clContent, notes)
	},
}

func getStatusLabel(status string) string {
	labels := map[string]string{
		"pending":   "📝 Pending",
		"applied":   "✅ Applied",
		"interview": "💼 Interview",
		"offer":     "🎉 Offer",
		"rejected":  "❌ Rejected",
	}
	if label, ok := labels[status]; ok {
		return label
	}
	return status
}

func init() {
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(applyCmd)
	statusCmd.AddCommand(updateStatusCmd)

	// Flags for status command
	statusCmd.Flags().String("filter", "", "Filter by status (pending, applied, interview, offer, rejected)")

	// Flags for update command
	updateStatusCmd.Flags().String("status", "", "New status (pending, applied, interview, offer, rejected)")
	updateStatusCmd.Flags().String("notes", "", "Add notes to the application")

	// Flags for apply command
	applyCmd.Flags().String("notes", "", "Add notes to the application")
	applyCmd.Flags().String("batch", "", "Apply to multiple jobs from a file (one job ID per line)")
	applyCmd.Flags().Bool("dry-run", false, "Preview without actually applying")
	applyCmd.Flags().Bool("auto", false, "Automatically apply using browser automation (requires LinkedIn/Greenhouse/Lever)")
}

// handleBatchApply processes batch job applications
func handleBatchApply(ctx context.Context, batchFile string, dryRun bool, autoApply bool) error {
	// Read job IDs from file
	data, err := os.ReadFile(batchFile)
	if err != nil {
		return fmt.Errorf("error reading batch file: %w", err)
	}

	// Parse job IDs (one per line)
	lines := strings.Split(string(data), "\n")
	jobIDs := []int{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		var jobID int
		if _, err := fmt.Sscanf(line, "%d", &jobID); err == nil {
			jobIDs = append(jobIDs, jobID)
		}
	}

	if len(jobIDs) == 0 {
		fmt.Println("No valid job IDs found in batch file")
		return nil
	}

	fmt.Printf("Found %d jobs to apply to\n", len(jobIDs))
	if dryRun {
		fmt.Println("DRY RUN MODE - No applications will be created")
	}
	if autoApply {
		fmt.Println("AUTO-APPLY MODE - Using browser automation")
	}

	successCount := 0
	failCount := 0

	// Get user and resume once
	var user *models.User
	var resume *models.Resume
	if autoApply {
		var err error
		user, err = database.GetUser()
		if err != nil || user == nil {
			return fmt.Errorf("user profile not configured")
		}
		resume, _ = database.GetDefaultResume()
		if resume == nil {
			return fmt.Errorf("no default resume set")
		}
	}

	for _, jobID := range jobIDs {
		if dryRun {
			job, err := database.GetJob(jobID)
			if err != nil {
				fmt.Printf("  [DRY RUN] Job %d: Not found\n", jobID)
				failCount++
				continue
			}
			fmt.Printf("  [DRY RUN] Would apply to: %s at %s\n", job.Title, job.Company)
			successCount++
			continue
		}

		// Check if job exists
		job, err := database.GetJob(jobID)
		if err != nil {
			fmt.Printf("  ✗ Job %d: Not found\n", jobID)
			failCount++
			continue
		}

		// Check if already applied
		existing, _ := database.GetApplicationByJobID(jobID)
		if existing != nil {
			fmt.Printf("  ⊘ Job %d: Already applied (Status: %s)\n", jobID, existing.Status)
			continue
		}

		if autoApply {
			// Get cover letter if available
			coverLetter, _ := database.GetCoverLetterByJobID(jobID)
			var clContent string
			if coverLetter != nil {
				clContent = coverLetter.Content
			}

			// Attempt auto-apply
			if err := applyJobAuto(ctx, job, user, resume, clContent, jobID, ""); err != nil {
				fmt.Printf("  ✗ Job %d: Auto-apply failed - %v\n", jobID, err)
				failCount++
				continue
			}
			fmt.Printf("  ✓ Auto-applied to: %s at %s\n", job.Title, job.Company)
			successCount++
		} else {
			// Manual apply
			if err := applyJobManual(jobID, resume, "", ""); err != nil {
				fmt.Printf("  ✗ Job %d: Error - %v\n", jobID, err)
				failCount++
				continue
			}
			fmt.Printf("  ✓ Marked as applied: %s at %s\n", job.Title, job.Company)
			successCount++
		}
	}

	fmt.Printf("\n✓ Successfully applied to %d jobs\n", successCount)
	if failCount > 0 {
		fmt.Printf("✗ Failed to apply to %d jobs\n", failCount)
	}
	return nil
}

// applyJobAuto attempts automatic application using browser automation
func applyJobAuto(ctx context.Context, job *models.Job, user *models.User, resume *models.Resume, coverLetter string, jobID int, notes string) error {
	fmt.Printf("⏳ Auto-applying to %s at %s...\n", job.Title, job.Company)

	// Check if job board is supported
	if !applicator.CanAutoApply(job) {
		return fmt.Errorf("auto-apply not supported for %s", job.Source)
	}

	// Perform auto-apply
	result := applicator.ApplyToJob(ctx, job, user, resume, coverLetter)

	if !result.Success {
		return fmt.Errorf(result.Message)
	}

	// Create application record on success
	resumeID := resume.ID
	app := &models.Application{
		JobID:       jobID,
		ResumeID:    &resumeID,
		CoverLetter: coverLetter,
		Status:      "applied",
		Notes:       notes,
	}

	if err := database.CreateApplication(app); err != nil {
		return fmt.Errorf("failed to create application record: %w", err)
	}

	fmt.Printf("✓ %s\n", result.Message)
	return nil
}

// applyJobManual creates an application record for manual application
func applyJobManual(jobID int, resume *models.Resume, coverLetter string, notes string) error {
	resumeID := resume.ID
	app := &models.Application{
		JobID:       jobID,
		ResumeID:    &resumeID,
		CoverLetter: coverLetter,
		Status:      "applied",
		Notes:       notes,
	}

	if err := database.CreateApplication(app); err != nil {
		return fmt.Errorf("error creating application: %w", err)
	}

	job, _ := database.GetJob(jobID)
	if job != nil {
		fmt.Printf("✓ Marked as applied: %s at %s\n", job.Title, job.Company)
		if resume != nil {
			fmt.Printf("  Using resume: %s\n", resume.Name)
		}
	}
	return nil
}
