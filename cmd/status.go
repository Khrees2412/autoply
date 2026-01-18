package cmd

import (
	"fmt"
	"os"

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
	Short: "Mark a job as applied",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply apply 1
  autoply apply 5 --notes "Applied via LinkedIn"`,
	Run: func(cmd *cobra.Command, args []string) {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			fmt.Println("Invalid job ID. Must be a number.")
			return
		}

		notes, _ := cmd.Flags().GetString("notes")

		// Check if job exists
		job, err := database.GetJob(jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Job not found\n")
			return
		}

		// Check if already applied
		existing, _ := database.GetApplicationByJobID(jobID)
		if existing != nil {
			fmt.Printf("Already applied to this job (Status: %s)\n", existing.Status)
			fmt.Println("Use 'autoply status update <job-id>' to change status")
			return
		}

		// Get default resume
		resume, _ := database.GetDefaultResume()
		var resumeID int
		if resume != nil {
			resumeID = resume.ID
		}

		// Check for generated cover letter
		coverLetter, _ := database.GetCoverLetterByJobID(jobID)
		var clContent string
		if coverLetter != nil {
			clContent = coverLetter.Content
		}

		// Create application
		app := &models.Application{
			JobID:       jobID,
			ResumeID:    resumeID,
			CoverLetter: clContent,
			Status:      "applied",
			Notes:       notes,
		}

		if err := database.CreateApplication(app); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Marked as applied: %s at %s\n", job.Title, job.Company)
		if resume != nil {
			fmt.Printf("  Using resume: %s\n", resume.Name)
		}
		if coverLetter != nil {
			fmt.Println("  Using generated cover letter")
		}
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
}
