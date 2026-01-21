package cmd

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			return fmt.Errorf("invalid job ID: must be a number")
		}

		save, _ := cmd.Flags().GetBool("save")

		// Get job details
		job, err := database.GetJob(jobID)
		if err != nil {
			return fmt.Errorf("fetch job: %w", err)
		}

		// Get user profile
		user, err := database.GetUser()
		if err != nil {
			return fmt.Errorf("fetch user profile: %w", err)
		}
		if user == nil {
			cmd.Println("No profile found. Run 'autoply init' to create your profile first.")
			return nil
		}

		// Get user skills and experience (warn on error, continue with empty)
		skills, err := database.GetUserSkills(user.ID)
		if err != nil {
			cmd.Printf("Warning: could not fetch skills: %v\n", err)
			skills = []*models.Skill{}
		}

		experiences, err := database.GetUserExperiences(user.ID)
		if err != nil {
			cmd.Printf("Warning: could not fetch experiences: %v\n", err)
			experiences = []*models.Experience{}
		}

		cmd.Println("Generating cover letter with AI...")
		cmd.Printf("Job: %s at %s\n\n", job.Title, job.Company)

		// Generate cover letter (pass context for cancellation support)
		coverLetter, err := ai.GenerateCoverLetter(cmd.Context(), job, user, skills, experiences)
		if err != nil {
			return fmt.Errorf("generate cover letter: %w", err)
		}

		// Display the cover letter
		cmd.Println(titleStyle.Render("Generated Cover Letter"))
		cmd.Println(coverLetter)

		// Save to database if requested
		if save {
			cl := &models.CoverLetter{
				JobID:   jobID,
				Content: coverLetter,
				IsSent:  false,
			}
			if err := database.CreateCoverLetter(cl); err != nil {
				return fmt.Errorf("save cover letter: %w", err)
			}
			cmd.Println("\n✓ Cover letter saved to database")
		} else {
			cmd.Println("\nTo save this cover letter, run with --save flag")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(generateCoverLetterCmd)

	// Flags for cover-letter command
	generateCoverLetterCmd.Flags().Bool("save", false, "Save the generated cover letter to database")
}
