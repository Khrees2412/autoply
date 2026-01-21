package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/khrees2412/autoply/internal/app"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "autoply",
	Short: "AI-powered job application automation CLI",
	Long: `Autoply is a CLI/TUI application that helps you automate your job search process.
It aggregates jobs, generates AI-powered cover letters, manages applications, and more.`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize app with all dependencies
		application, err := app.NewApp(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}

		// Store app in command context
		cmd.SetContext(app.SetAppInContext(cmd.Context(), application))

		// Register cleanup on exit
		return nil
	},
}

// Execute runs the root command
func Execute() {
	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootCmd.SetContext(ctx)

	// Register signal handlers for cleanup
	go func() {
		// In a real app, would use signal.Notify here
		// For now, just let deferred cleanup happen
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Cleanup: close app resources
	if appInstance := app.GetAppFromContext(ctx); appInstance != nil {
		appInstance.Close()
	}
}
