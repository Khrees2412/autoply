package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func init() {
	rootCmd.AddCommand(resumeCmd)
	resumeCmd.AddCommand(addResumeCmd)
	resumeCmd.AddCommand(listResumesCmd)

	// Flags for add command
	addResumeCmd.Flags().String("name", "", "Name for the resume")
	addResumeCmd.Flags().Bool("default", false, "Set as default resume")
}
