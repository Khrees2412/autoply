package cmd

import (
	"fmt"
	"os"

	"github.com/khrees2412/autoply/internal/ai"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate AI content",
	Long:  "Generate cover letters and other AI-powered content",
}

var generateCoverLetterCmd = &cobra.Command{
	Use:   "cover-letter <job-id>",
	Short: "Generate a cover letter for a job",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply generate cover-letter 1
  autoply generate cover-letter 5 --save`,
	Run: func(cmd *cobra.Command, args []string) {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			fmt.Println("Invalid job ID. Must be a number.")
			return
		}

		save, _ := cmd.Flags().GetBool("save")

		// Get job details
		job, err := database.GetJob(jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching job: %v\n", err)
			os.Exit(1)
		}

		// Get user profile
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create your profile first.")
			return
		}

		// Get user skills and experience
		skills, err := database.GetUserSkills(user.ID)
		if err != nil {
			skills = []*models.Skill{}
		}

		experiences, err := database.GetUserExperiences(user.ID)
		if err != nil {
			experiences = []*models.Experience{}
		}

		fmt.Println("Generating cover letter with AI...")
		fmt.Printf("Job: %s at %s\n\n", job.Title, job.Company)

		// Generate cover letter
		coverLetter, err := ai.GenerateCoverLetter(job, user, skills, experiences)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating cover letter: %v\n", err)
			os.Exit(1)
		}

		// Display the cover letter
		fmt.Println(titleStyle.Render("Generated Cover Letter"))
		fmt.Println(coverLetter)

		// Save to database if requested
		if save {
			cl := &models.CoverLetter{
				JobID:   jobID,
				Content: coverLetter,
				IsSent:  false,
			}
			if err := database.CreateCoverLetter(cl); err != nil {
				fmt.Fprintf(os.Stderr, "\nError saving cover letter: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\n✓ Cover letter saved to database")
		} else {
			fmt.Println("\nTo save this cover letter, run with --save flag")
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(generateCoverLetterCmd)

	// Flags for cover-letter command
	generateCoverLetterCmd.Flags().Bool("save", false, "Save the generated cover letter to database")
}
