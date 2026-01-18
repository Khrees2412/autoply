package cmd

import (
	"fmt"
	"os"

	"github.com/khrees2412/autoply/internal/config"
	"github.com/khrees2412/autoply/internal/database"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "autoply",
	Short: "AI-powered job application automation CLI",
	Long: `Autoply is a CLI/TUI application that helps you automate your job search process.
It aggregates jobs, generates AI-powered cover letters, manages applications, and more.`,
	Version: "0.1.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig initializes configuration and database
func initConfig() {
	// Initialize config
	if err := config.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Initialize database
	if err := database.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
		os.Exit(1)
	}
}
