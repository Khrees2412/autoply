package cmd

import (
	"fmt"
	"os"

	"github.com/khrees2412/autoply/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "View and update configuration settings",
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(titleStyle.Render("Configuration"))
		fmt.Printf("%s %s\n", labelStyle.Render("Config File:"), config.GetConfigPath())
		fmt.Printf("%s %s\n", labelStyle.Render("AI Provider:"), config.AppConfig.AIProvider)
		fmt.Printf("%s %s\n", labelStyle.Render("Default Model:"), config.AppConfig.DefaultModel)
		
		// Show if API keys are configured (but don't show the actual keys)
		if config.AppConfig.OpenAIKey != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("OpenAI Key:"), "✓ Configured")
		} else {
			fmt.Printf("%s %s\n", labelStyle.Render("OpenAI Key:"), "✗ Not configured")
		}
		
		if config.AppConfig.AnthropicKey != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("Anthropic Key:"), "✓ Configured")
		} else {
			fmt.Printf("%s %s\n", labelStyle.Render("Anthropic Key:"), "✗ Not configured")
		}
		
		// Show LinkedIn credentials status
		if config.AppConfig.LinkedInEmail != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("LinkedIn Email:"), "✓ Configured")
		} else {
			fmt.Printf("%s %s\n", labelStyle.Render("LinkedIn Email:"), "✗ Not configured")
		}
		
		if config.AppConfig.LinkedInPassword != "" {
			fmt.Printf("%s %s\n", labelStyle.Render("LinkedIn Password:"), "✓ Configured")
		} else {
			fmt.Printf("%s %s\n", labelStyle.Render("LinkedIn Password:"), "✗ Not configured")
		}
	},
}

var setConfigCmd = &cobra.Command{
	Use:   "set",
	Short: "Update a configuration value",
	Example: `  autoply config set --key openai_key --value sk-...
  autoply config set --key ai_provider --value anthropic
  autoply config set --key default_model --value gpt-4o
  autoply config set --key linkedin_email --value your-email@example.com
  autoply config set --key linkedin_password --value your-password`,
	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		if key == "" || value == "" {
			fmt.Println("Both --key and --value are required")
			return
		}

		// Validate key
		validKeys := []string{"openai_key", "anthropic_key", "ai_provider", "default_model", "linkedin_email", "linkedin_password"}
		valid := false
		for _, k := range validKeys {
			if k == key {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Invalid key. Must be one of: %v\n", validKeys)
			return
		}

		if err := config.Set(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Configuration updated: %s\n", key)
		
		// Reload config
		if err := config.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not reload config: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(showConfigCmd)
	configCmd.AddCommand(setConfigCmd)

	// Flags for set command
	setConfigCmd.Flags().String("key", "", "Configuration key")
	setConfigCmd.Flags().String("value", "", "Configuration value")
}
