package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage your skills",
	Long:  "Add, list, and remove skills from your profile",
}

var addSkillCmd = &cobra.Command{
	Use:   "add <skill-name>",
	Short: "Add a skill",
	Args:  cobra.ExactArgs(1),
	Example: `  autoply skill add "Go"
  autoply skill add "Python" --level advanced`,
	Run: func(cmd *cobra.Command, args []string) {
		skillName := args[0]
		level, _ := cmd.Flags().GetString("level")

		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create your profile first.")
			return
		}

		if level == "" {
			level = "intermediate"
		}

		// Validate level
		validLevels := []string{"beginner", "intermediate", "advanced", "expert"}
		valid := false
		for _, l := range validLevels {
			if l == level {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Invalid level. Must be one of: %v\n", validLevels)
			return
		}

		skill := &models.Skill{
			UserID:           user.ID,
			SkillName:        skillName,
			ProficiencyLevel: level,
		}

		if err := database.CreateSkill(skill); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding skill: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Added skill: %s (%s)\n", skillName, level)
	},
}

var listSkillsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all skills",
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create your profile first.")
			return
		}

		skills, err := database.GetUserSkills(user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching skills: %v\n", err)
			os.Exit(1)
		}

		if len(skills) == 0 {
			fmt.Println("No skills found. Add skills with 'autoply skill add <skill-name>'")
			return
		}

		fmt.Println(titleStyle.Render("Your Skills"))
		for i, skill := range skills {
			fmt.Printf("%d. %s", i+1, skill.SkillName)
			if skill.ProficiencyLevel != "" {
				fmt.Printf(" (%s)", skill.ProficiencyLevel)
			}
			fmt.Println()
		}
	},
}

var removeSkillCmd = &cobra.Command{
	Use:   "remove <skill-id>",
	Short: "Remove a skill",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var skillID int
		if _, err := fmt.Sscanf(args[0], "%d", &skillID); err != nil {
			fmt.Println("Invalid skill ID. Must be a number.")
			return
		}

		if err := database.DeleteSkill(skillID); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing skill: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Removed skill (ID: %d)\n", skillID)
	},
}

var experienceCmd = &cobra.Command{
	Use:   "experience",
	Short: "Manage your work experience",
	Long:  "Add, list, and remove work experience entries",
}

var addExperienceCmd = &cobra.Command{
	Use:   "add",
	Short: "Add work experience",
	Example: `  autoply experience add --company "Acme Inc" --title "Software Engineer" --start "2020-01-01"`,
	Run: func(cmd *cobra.Command, args []string) {
		company, _ := cmd.Flags().GetString("company")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		startDateStr, _ := cmd.Flags().GetString("start")
		endDateStr, _ := cmd.Flags().GetString("end")

		if company == "" || title == "" || startDateStr == "" {
			fmt.Println("Company, title, and start date are required")
			return
		}

		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create your profile first.")
			return
		}

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			fmt.Println("Invalid start date format. Use YYYY-MM-DD")
			return
		}

		var endDate *time.Time
		if endDateStr != "" {
			ed, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				fmt.Println("Invalid end date format. Use YYYY-MM-DD")
				return
			}
			endDate = &ed
		}

		exp := &models.Experience{
			UserID:      user.ID,
			Company:     company,
			Title:       title,
			Description: description,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		if err := database.CreateExperience(exp); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding experience: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Added experience: %s at %s\n", title, company)
	},
}

var listExperiencesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all work experience",
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create your profile first.")
			return
		}

		experiences, err := database.GetUserExperiences(user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching experiences: %v\n", err)
			os.Exit(1)
		}

		if len(experiences) == 0 {
			fmt.Println("No experience found. Add experience with 'autoply experience add'")
			return
		}

		fmt.Println(titleStyle.Render("Your Experience"))
		for i, exp := range experiences {
			fmt.Printf("\n%d. %s at %s\n", i+1, exp.Title, exp.Company)
			if exp.EndDate != nil {
				fmt.Printf("   %s - %s\n", exp.StartDate.Format("Jan 2006"), exp.EndDate.Format("Jan 2006"))
			} else {
				fmt.Printf("   %s - Present\n", exp.StartDate.Format("Jan 2006"))
			}
			if exp.Description != "" {
				fmt.Printf("   %s\n", exp.Description)
			}
		}
	},
}

var removeExperienceCmd = &cobra.Command{
	Use:   "remove <experience-id>",
	Short: "Remove work experience",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var expID int
		if _, err := fmt.Sscanf(args[0], "%d", &expID); err != nil {
			fmt.Println("Invalid experience ID. Must be a number.")
			return
		}

		if err := database.DeleteExperience(expID); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing experience: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Removed experience (ID: %d)\n", expID)
	},
}

func init() {
	rootCmd.AddCommand(skillCmd)
	rootCmd.AddCommand(experienceCmd)

	skillCmd.AddCommand(addSkillCmd)
	skillCmd.AddCommand(listSkillsCmd)
	skillCmd.AddCommand(removeSkillCmd)

	experienceCmd.AddCommand(addExperienceCmd)
	experienceCmd.AddCommand(listExperiencesCmd)
	experienceCmd.AddCommand(removeExperienceCmd)

	// Flags for add skill
	addSkillCmd.Flags().String("level", "intermediate", "Proficiency level (beginner, intermediate, advanced, expert)")

	// Flags for add experience
	addExperienceCmd.Flags().String("company", "", "Company name (required)")
	addExperienceCmd.Flags().String("title", "", "Job title (required)")
	addExperienceCmd.Flags().String("description", "", "Job description")
	addExperienceCmd.Flags().String("start", "", "Start date (YYYY-MM-DD, required)")
	addExperienceCmd.Flags().String("end", "", "End date (YYYY-MM-DD, leave empty for current)")
}

