package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			MarginTop(1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage your profile",
	Long:  "Create and update your profile information used for job applications",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize your profile with an interactive wizard",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if user already exists
		user, err := database.GetUser()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking for existing profile: %v\n", err)
			os.Exit(1)
		}

		if user != nil {
			fmt.Println(titleStyle.Render("Profile Already Exists"))
			fmt.Println("Use 'autoply profile show' to view or 'autoply profile set' to update.")
			return
		}

		fmt.Println(titleStyle.Render("Welcome to Autoply! Let's set up your profile."))

		reader := bufio.NewReader(os.Stdin)

		// Collect user information
		fmt.Print(labelStyle.Render("Full Name: "))
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)

		fmt.Print(labelStyle.Render("Email: "))
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)

		fmt.Print(labelStyle.Render("Phone (optional): "))
		phone, _ := reader.ReadString('\n')
		phone = strings.TrimSpace(phone)

		fmt.Print(labelStyle.Render("Location: "))
		location, _ := reader.ReadString('\n')
		location = strings.TrimSpace(location)

		fmt.Print(labelStyle.Render("LinkedIn URL (optional): "))
		linkedin, _ := reader.ReadString('\n')
		linkedin = strings.TrimSpace(linkedin)

		fmt.Print(labelStyle.Render("GitHub URL (optional): "))
		github, _ := reader.ReadString('\n')
		github = strings.TrimSpace(github)

		// Create default preferences
		prefs := models.UserPreferences{
			DesiredRoles: []string{},
			Locations:    []string{},
			RemoteOnly:   false,
		}
		prefsJSON, _ := json.Marshal(prefs)

		// Create user
		user = &models.User{
			Name:        name,
			Email:       email,
			Phone:       phone,
			Location:    location,
			LinkedInURL: linkedin,
			GitHubURL:   github,
			Preferences: string(prefsJSON),
		}

		if err := database.CreateUser(user); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(titleStyle.Render("\n✓ Profile created successfully!"))
		fmt.Println("Next steps:")
		fmt.Println("  1. Configure your AI API key: autoply config set --key openai_key --value YOUR_KEY")
		fmt.Println("  2. Add your resume: autoply resume add /path/to/resume.pdf")
		fmt.Println("  3. Start adding jobs: autoply job add --url JOB_URL")
	},
}

var showProfileCmd = &cobra.Command{
	Use:   "show",
	Short: "Display your profile information",
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.GetUser()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching profile: %v\n", err)
			os.Exit(1)
		}

		if user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create one.")
			return
		}

		fmt.Println(titleStyle.Render("Your Profile"))
		fmt.Printf("%s %s\n", labelStyle.Render("Name:"), valueStyle.Render(user.Name))
		fmt.Printf("%s %s\n", labelStyle.Render("Email:"), valueStyle.Render(user.Email))
		if user.Phone != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("Phone:"), valueStyle.Render(user.Phone))
		}
		fmt.Printf("%s %s\n", labelStyle.Render("Location:"), valueStyle.Render(user.Location))
		if user.LinkedInURL != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("LinkedIn:"), valueStyle.Render(user.LinkedInURL))
		}
		if user.GitHubURL != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("GitHub:"), valueStyle.Render(user.GitHubURL))
		}

		// Get skills
		skills, err := database.GetUserSkills(user.ID)
		if err == nil && len(skills) > 0 {
			fmt.Println(labelStyle.Render("\nSkills:"))
			for _, skill := range skills {
				fmt.Printf("  • %s", skill.SkillName)
				if skill.ProficiencyLevel != "" {
					fmt.Printf(" (%s)", skill.ProficiencyLevel)
				}
				fmt.Println()
			}
		}

		// Get experiences
		experiences, err := database.GetUserExperiences(user.ID)
		if err == nil && len(experiences) > 0 {
			fmt.Println(labelStyle.Render("\nExperience:"))
			for _, exp := range experiences {
				fmt.Printf("  • %s at %s\n", exp.Title, exp.Company)
			}
		}
	},
}

var editProfileCmd = &cobra.Command{
	Use:   "edit",
	Short: "Interactively edit your profile",
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create one.")
			return
		}

		fmt.Println(titleStyle.Render("Edit Profile"))
		fmt.Println("Press Enter to keep current value, or type a new value")

		reader := bufio.NewReader(os.Stdin)

		// Name
		fmt.Printf("%s [%s]: ", labelStyle.Render("Full Name"), user.Name)
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name != "" {
			user.Name = name
		}

		// Email
		fmt.Printf("%s [%s]: ", labelStyle.Render("Email"), user.Email)
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if email != "" {
			user.Email = email
		}

		// Phone
		fmt.Printf("%s [%s]: ", labelStyle.Render("Phone"), user.Phone)
		phone, _ := reader.ReadString('\n')
		phone = strings.TrimSpace(phone)
		if phone != "" {
			user.Phone = phone
		}

		// Location
		fmt.Printf("%s [%s]: ", labelStyle.Render("Location"), user.Location)
		location, _ := reader.ReadString('\n')
		location = strings.TrimSpace(location)
		if location != "" {
			user.Location = location
		}

		// LinkedIn
		fmt.Printf("%s [%s]: ", labelStyle.Render("LinkedIn URL"), user.LinkedInURL)
		linkedin, _ := reader.ReadString('\n')
		linkedin = strings.TrimSpace(linkedin)
		if linkedin != "" {
			user.LinkedInURL = linkedin
		}

		// GitHub
		fmt.Printf("%s [%s]: ", labelStyle.Render("GitHub URL"), user.GitHubURL)
		github, _ := reader.ReadString('\n')
		github = strings.TrimSpace(github)
		if github != "" {
			user.GitHubURL = github
		}

		if err := database.UpdateUser(user); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n✓ Profile updated successfully!")
	},
}

var setProfileCmd = &cobra.Command{
	Use:   "set",
	Short: "Update a profile field",
	Example: `  autoply profile set --name "John Doe"
  autoply profile set --email "john@example.com"
  autoply profile set --location "San Francisco, CA"`,
	Run: func(cmd *cobra.Command, args []string) {
		user, err := database.GetUser()
		if err != nil || user == nil {
			fmt.Println("No profile found. Run 'autoply init' to create one.")
			return
		}

		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		phone, _ := cmd.Flags().GetString("phone")
		location, _ := cmd.Flags().GetString("location")
		linkedin, _ := cmd.Flags().GetString("linkedin")
		github, _ := cmd.Flags().GetString("github")

		updated := false

		if name != "" {
			user.Name = name
			updated = true
		}
		if email != "" {
			user.Email = email
			updated = true
		}
		if phone != "" {
			user.Phone = phone
			updated = true
		}
		if location != "" {
			user.Location = location
			updated = true
		}
		if linkedin != "" {
			user.LinkedInURL = linkedin
			updated = true
		}
		if github != "" {
			user.GitHubURL = github
			updated = true
		}

		if !updated {
			fmt.Println("No fields to update. Use flags like --name, --email, etc.")
			return
		}

		if err := database.UpdateUser(user); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating profile: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Profile updated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(initCmd)
	profileCmd.AddCommand(showProfileCmd)
	profileCmd.AddCommand(editProfileCmd)
	profileCmd.AddCommand(setProfileCmd)

	// Flags for set command
	setProfileCmd.Flags().String("name", "", "Update name")
	setProfileCmd.Flags().String("email", "", "Update email")
	setProfileCmd.Flags().String("phone", "", "Update phone")
	setProfileCmd.Flags().String("location", "", "Update location")
	setProfileCmd.Flags().String("linkedin", "", "Update LinkedIn URL")
	setProfileCmd.Flags().String("github", "", "Update GitHub URL")
}
