package applicator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/khrees2412/autoply/pkg/models"
)

func TestCanAutoApply(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected bool
	}{
		{
			name:     "LinkedIn supported",
			source:   "linkedin",
			expected: true,
		},
		{
			name:     "Greenhouse supported",
			source:   "greenhouse",
			expected: true,
		},
		{
			name:     "Lever supported",
			source:   "lever",
			expected: true,
		},
		{
			name:     "LinkedIn uppercase",
			source:   "LINKEDIN",
			expected: true,
		},
		{
			name:     "Indeed unsupported",
			source:   "indeed",
			expected: false,
		},
		{
			name:     "Glassdoor unsupported",
			source:   "glassdoor",
			expected: false,
		},
		{
			name:     "Manual unsupported",
			source:   "manual",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &models.Job{Source: tt.source}
			result := CanAutoApply(job)
			if result != tt.expected {
				t.Errorf("CanAutoApply(%q) = %v, expected %v", tt.source, result, tt.expected)
			}
		})
	}
}

func TestGetApplicationFormFields(t *testing.T) {
	fields := GetApplicationFormFields()

	expectedFields := []string{
		"first_name", "last_name", "email", "phone", "location",
		"linkedin_url", "github_url", "portfolio_url", "cover_letter",
		"resume_file", "years_experience", "availability", "salary_expectation",
	}

	for _, field := range expectedFields {
		if _, ok := fields[field]; !ok {
			t.Errorf("Field %q not found in GetApplicationFormFields()", field)
		}
	}
}

func TestApplyToJobNoResume(t *testing.T) {
	ctx := context.Background()
	job := &models.Job{
		ID:     1,
		Title:  "Test Job",
		Source: "linkedin",
		URL:    "https://example.com/job",
	}
	user := &models.User{Name: "Test User", Email: "test@example.com"}

	// Test with nil resume
	result := ApplyToJob(ctx, job, user, nil, "")
	if result.Success {
		t.Error("Expected failure with nil resume")
	}
	if result.Message != "No resume available for auto-apply" {
		t.Errorf("Unexpected error message: %q", result.Message)
	}

	// Test with empty resume file path
	resume := &models.Resume{ID: 1, Name: "test.pdf", FilePath: ""}
	result = ApplyToJob(ctx, job, user, resume, "")
	if result.Success {
		t.Error("Expected failure with empty resume path")
	}
}

func TestApplyToJobMissingResume(t *testing.T) {
	ctx := context.Background()
	job := &models.Job{
		ID:     1,
		Title:  "Test Job",
		Source: "linkedin",
		URL:    "https://example.com/job",
	}
	user := &models.User{Name: "Test User", Email: "test@example.com"}
	resume := &models.Resume{ID: 1, Name: "test.pdf", FilePath: "/nonexistent/path/resume.pdf"}

	result := ApplyToJob(ctx, job, user, resume, "")
	if result.Success {
		t.Error("Expected failure with nonexistent resume file")
	}
	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestApplyToJobUnsupportedSource(t *testing.T) {
	ctx := context.Background()

	// Create temporary resume file
	tmpFile, err := os.CreateTemp("", "test-resume-*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	job := &models.Job{
		ID:     1,
		Title:  "Test Job",
		Source: "indeed",
		URL:    "https://indeed.com/job",
	}
	user := &models.User{Name: "Test User", Email: "test@example.com"}
	resume := &models.Resume{ID: 1, Name: "test.pdf", FilePath: tmpFile.Name()}

	result := ApplyToJob(ctx, job, user, resume, "")
	if result.Success {
		t.Error("Expected failure with unsupported source")
	}
	if result.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestUploadFileToForm(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test with valid file (without actually opening browser)
	absPath, _ := filepath.Abs(tmpFile.Name())
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("File validation failed: %v", err)
	}

	// The actual chromedp call would fail without a running context,
	// but the file validation should pass - this just verifies the path handling
}

func TestWaitForElement(t *testing.T) {
	// Note: This function requires an active chromedp context
	// which we don't have in unit tests. Full integration test
	// would use a real browser context. This is skipped in unit tests.
	t.Skip("Requires active chromedp browser context")
}

func TestApplicationResult(t *testing.T) {
	result := &ApplicationResult{
		Success:        true,
		Message:        "Test message",
		ScreenshotPath: "/path/to/screenshot.png",
		Error:          nil,
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Message != "Test message" {
		t.Errorf("Unexpected message: %q", result.Message)
	}
}
