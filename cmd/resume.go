package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/khrees2412/autoply/internal/ai"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Manage resumes",
	Long:  "Add, list, and manage your resumes",
}

var addResumeCmd = &cobra.Command{
	Use:   "add <file-path>",
	Short: "Add a resume",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply resume add ~/Documents/resume.pdf
  autoply resume add ./my-resume.pdf --name "Software Engineer Resume"`,
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		name, _ := cmd.Flags().GetString("name")
		setDefault, _ := cmd.Flags().GetBool("default")

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "File not found: %s\n", filePath)
			os.Exit(1)
		}

		// Get home directory for storage
		homeDir, _ := os.UserHomeDir()
		resumeDir := filepath.Join(homeDir, ".autoply", "resumes")
		if err := os.MkdirAll(resumeDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating resume directory: %v\n", err)
			os.Exit(1)
		}

		// Copy file to autoply directory
		fileName := filepath.Base(filePath)
		destPath := filepath.Join(resumeDir, fileName)

		src, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		defer src.Close()

		dst, err := os.Create(destPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination file: %v\n", err)
			os.Exit(1)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			fmt.Fprintf(os.Stderr, "Error copying file: %v\n", err)
			os.Exit(1)
		}

		// Use filename as name if not provided
		if name == "" {
			name = fileName
		}

		// Create resume record
		resume := &models.Resume{
			Name:        name,
			FilePath:    destPath,
			ContentText: "", // TODO: Extract text from PDF
			IsDefault:   setDefault,
		}

		if err := database.CreateResume(resume); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving resume: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Resume added: %s (ID: %d)\n", resume.Name, resume.ID)
		if setDefault {
			fmt.Println("  Set as default resume")
		}
	},
}

var listResumesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resumes",
	Run: func(cmd *cobra.Command, args []string) {
		resumes, err := database.GetAllResumes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching resumes: %v\n", err)
			os.Exit(1)
		}

		if len(resumes) == 0 {
			fmt.Println("No resumes found. Add a resume with 'autoply resume add <file>'")
			return
		}

		fmt.Println(titleStyle.Render("Your Resumes"))
		for i, resume := range resumes {
			defaultMarker := ""
			if resume.IsDefault {
				defaultMarker = " [DEFAULT]"
			}
			fmt.Printf("\n%d. %s%s\n", i+1, resume.Name, defaultMarker)
			fmt.Printf("   %s %s\n", labelStyle.Render("ID:"), fmt.Sprintf("%d", resume.ID))
			fmt.Printf("   %s %s\n", labelStyle.Render("File:"), resume.FilePath)
			fmt.Printf("   %s %s\n", labelStyle.Render("Added:"), resume.CreatedAt.Format("Jan 2, 2006"))
		}
	},
}

var tailorResumeCmd = &cobra.Command{
	Use:   "tailor <job-id>",
	Short: "AI-optimize resume for a specific job",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply resume tailor 1
  autoply resume tailor 5 --resume-id 2`,
	Run: func(cmd *cobra.Command, args []string) {
		var jobID int
		if _, err := fmt.Sscanf(args[0], "%d", &jobID); err != nil {
			fmt.Println("Invalid job ID. Must be a number.")
			return
		}

		resumeID, _ := cmd.Flags().GetInt("resume-id")

		// Get job
		job, err := database.GetJob(jobID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching job: %v\n", err)
			os.Exit(1)
		}

		// Get resume
		var resume *models.Resume
		if resumeID > 0 {
			// Get specific resume (would need GetResume function)
			resumes, _ := database.GetAllResumes()
			for _, r := range resumes {
				if r.ID == resumeID {
					resume = r
					break
				}
			}
		} else {
			resume, _ = database.GetDefaultResume()
		}

		if resume == nil {
			fmt.Println("No resume found. Add a resume first with 'autoply resume add <file>'")
			return
		}

		// Get user profile
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' first.")
			return
		}

		fmt.Println("Tailoring resume with AI...")
		fmt.Printf("Job: %s at %s\n", job.Title, job.Company)
		fmt.Printf("Resume: %s\n\n", resume.Name)

		// Use AI to tailor resume
		tailoredContent, err := ai.TailorResume(resume, job, user)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error tailoring resume: %v\n", err)
			os.Exit(1)
		}

		// Display tailored resume
		fmt.Println(titleStyle.Render("Tailored Resume Content"))
		fmt.Println(tailoredContent)

		// Option to save as new resume
		save, _ := cmd.Flags().GetBool("save")
		if save {
			newResume := &models.Resume{
				Name:        fmt.Sprintf("%s - %s", resume.Name, job.Company),
				FilePath:    resume.FilePath, // Same file path
				ContentText: tailoredContent,
				IsDefault:   false,
			}
			if err := database.CreateResume(newResume); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving tailored resume: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n✓ Saved tailored resume (ID: %d)\n", newResume.ID)
		}
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
	resumeCmd.AddCommand(addResumeCmd)
	resumeCmd.AddCommand(listResumesCmd)
	resumeCmd.AddCommand(tailorResumeCmd)

	// Flags for add command
	addResumeCmd.Flags().String("name", "", "Name for the resume")
	addResumeCmd.Flags().Bool("default", false, "Set as default resume")

	// Flags for tailor command
	tailorResumeCmd.Flags().Int("resume-id", 0, "Resume ID to tailor (default: uses default resume)")
	tailorResumeCmd.Flags().Bool("save", false, "Save tailored resume as new version")
}
