package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/khrees2412/autoply/internal/database"
	"github.com/khrees2412/autoply/pkg/models"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View application statistics and insights",
	Long:  "Display analytics about your job applications, response rates, and trends",
	Run: func(cmd *cobra.Command, args []string) {
		apps, err := database.GetAllApplications()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching applications: %v\n", err)
			os.Exit(1)
		}

		if len(apps) == 0 {
			fmt.Println("No applications yet. Apply to jobs with 'autoply apply <job-id>'")
			return
		}

		// Calculate statistics
		stats := calculateStats(apps)

		fmt.Println(titleStyle.Render("Application Statistics"))

		// Overall stats
		fmt.Printf("\n%s\n", labelStyle.Render("Overview"))
		fmt.Printf("  Total Applications: %d\n", stats.Total)
		fmt.Printf("  Applied: %d\n", stats.Applied)
		fmt.Printf("  Interviews: %d\n", stats.Interviews)
		fmt.Printf("  Offers: %d\n", stats.Offers)
		fmt.Printf("  Rejected: %d\n", stats.Rejected)
		fmt.Printf("  Pending: %d\n", stats.Pending)

		// Response rates
		if stats.Applied > 0 {
			responseRate := float64(stats.Interviews+stats.Offers+stats.Rejected) / float64(stats.Applied) * 100
			fmt.Printf("\n%s\n", labelStyle.Render("Response Rate"))
			fmt.Printf("  Response Rate: %.1f%%\n", responseRate)
			if stats.Interviews > 0 {
				interviewRate := float64(stats.Interviews) / float64(stats.Applied) * 100
				fmt.Printf("  Interview Rate: %.1f%%\n", interviewRate)
			}
			if stats.Offers > 0 {
				offerRate := float64(stats.Offers) / float64(stats.Applied) * 100
				fmt.Printf("  Offer Rate: %.1f%%\n", offerRate)
			}
		}

		// Time to response
		if stats.AvgTimeToResponse > 0 {
			fmt.Printf("\n%s\n", labelStyle.Render("Response Time"))
			fmt.Printf("  Average Time to Response: %.1f days\n", stats.AvgTimeToResponse)
		}

		// Status breakdown
		fmt.Printf("\n%s\n", labelStyle.Render("Status Breakdown"))
		for status, count := range stats.StatusBreakdown {
			percentage := float64(count) / float64(stats.Total) * 100
			fmt.Printf("  %s: %d (%.1f%%)\n", status, count, percentage)
		}

		// Recent activity
		if len(stats.RecentActivity) > 0 {
			fmt.Printf("\n%s\n", labelStyle.Render("Recent Activity"))
			for _, activity := range stats.RecentActivity {
				fmt.Printf("  %s: %s\n", activity.Date.Format("Jan 2"), activity.Description)
			}
		}
	},
}

type Stats struct {
	Total              int
	Applied            int
	Interviews         int
	Offers             int
	Rejected           int
	Pending            int
	AvgTimeToResponse  float64
	StatusBreakdown    map[string]int
	RecentActivity     []Activity
}

type Activity struct {
	Date        time.Time
	Description string
}

func calculateStats(apps []*models.Application) Stats {
	stats := Stats{
		StatusBreakdown: make(map[string]int),
		RecentActivity:   []Activity{},
	}

	stats.Total = len(apps)
	var responseTimes []float64

	for _, app := range apps {
		stats.StatusBreakdown[app.Status]++

		switch app.Status {
		case "applied":
			stats.Applied++
		case "interview":
			stats.Interviews++
		case "offer":
			stats.Offers++
		case "rejected":
			stats.Rejected++
			// Calculate time to rejection
			if !app.AppliedAt.IsZero() {
				days := time.Since(app.AppliedAt).Hours() / 24
				responseTimes = append(responseTimes, days)
			}
		case "pending":
			stats.Pending++
		}

		// Recent activity
		if time.Since(app.AppliedAt) < 30*24*time.Hour {
			stats.RecentActivity = append(stats.RecentActivity, Activity{
				Date:        app.AppliedAt,
				Description: fmt.Sprintf("Applied to job #%d (%s)", app.JobID, app.Status),
			})
		}
	}

	// Calculate average response time
	if len(responseTimes) > 0 {
		sum := 0.0
		for _, t := range responseTimes {
			sum += t
		}
		stats.AvgTimeToResponse = sum / float64(len(responseTimes))
	}

	return stats
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

