package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/khrees2412/autoply/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// createTestDB creates a temporary test database
func createTestDB(t *testing.T) *sql.DB {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open with pragmas
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Run migrations
	if err := RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// setupTest sets up a test database and returns a cleanup function
func setupTest(t *testing.T) (oldDB *sql.DB, cleanup func()) {
	db := createTestDB(t)
	oldDB = DB
	DB = db

	return oldDB, func() {
		DB = oldDB
		db.Close()
	}
}

// TestCreateJob tests job creation with unique constraint
func TestCreateJob(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	job := &models.Job{
		Title:   "Software Engineer",
		Company: "Acme Inc",
		Location: "Remote",
		URL:     "https://example.com/job/123",
		Source:  "manual",
	}

	// Create job
	if err := CreateJob(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	if job.ID == 0 {
		t.Error("job ID not set after creation")
	}

	// Try to create duplicate (should fail)
	job2 := &models.Job{
		Title:   "Different Title",
		Company: "Different Company",
		URL:     "https://example.com/job/123",
		Source:  "manual",
	}

	if err := CreateJob(job2); err == nil {
		t.Error("should have failed to create duplicate job with same URL")
	}
}

// TestGetJob tests job retrieval
func TestGetJob(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	job := &models.Job{
		Title:   "QA Engineer",
		Company: "Beta Corp",
		Location: "NYC",
		Source:  "manual",
	}

	CreateJob(job)

	retrieved, err := GetJob(job.ID)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}

	if retrieved.Title != job.Title || retrieved.Company != job.Company {
		t.Error("retrieved job data doesn't match")
	}
}

// TestDeleteJobCascade tests that applications are deleted when job is deleted
func TestDeleteJobCascade(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create a user
	user := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
	}
	CreateUser(user)

	// Create a job
	job := &models.Job{
		Title:   "Engineer",
		Company: "Test Company",
		Source:  "manual",
	}
	CreateJob(job)

	// Create an application
	app := &models.Application{
		JobID:  job.ID,
		Status: "pending",
	}
	CreateApplication(app)

	// Verify application exists
	retrievedApp, err := GetApplicationByJobID(job.ID)
	if err != nil {
		t.Fatalf("failed to get application: %v", err)
	}
	if retrievedApp == nil {
		t.Error("application should exist")
	}

	// Delete job (should cascade delete application)
	if err := DeleteJob(job.ID); err != nil {
		t.Fatalf("failed to delete job: %v", err)
	}

	// Verify application is deleted
	deletedApp, err := GetApplicationByJobID(job.ID)
	if err != nil && err != sql.ErrNoRows {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedApp != nil {
		t.Error("application should be deleted when job is deleted")
	}
}

// TestGetAllJobs tests listing jobs
func TestGetAllJobs(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create multiple jobs
	for i := 1; i <= 3; i++ {
		job := &models.Job{
			Title:   fmt.Sprintf("Job %d", i),
			Company: fmt.Sprintf("Company %d", i),
			URL:     fmt.Sprintf("https://example.com/job/%d", i),
			Source:  "manual",
		}
		if err := CreateJob(job); err != nil {
			t.Fatalf("failed to create job %d: %v", i, err)
		}
		if job.ID == 0 {
			t.Errorf("job %d ID not set", i)
		}
	}

	jobs, err := GetAllJobs()
	if err != nil {
		t.Fatalf("failed to get all jobs: %v", err)
	}

	if len(jobs) != 3 {
		t.Errorf("expected 3 jobs, got %d: %+v", len(jobs), jobs)
	}
}

// TestForeignKeyConstraint verifies foreign keys are enabled
func TestForeignKeyConstraint(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Try to create an application with non-existent job ID
	_, err := DB.Exec(`
		INSERT INTO applications (job_id, status) VALUES (99999, 'pending')
	`)

	if err == nil {
		t.Error("should have failed due to foreign key constraint")
	}
}

// BenchmarkCreateJob benchmarks job creation
func BenchmarkCreateJob(b *testing.B) {
	db := createTestDB(&testing.T{})
	oldDB := DB
	DB = db
	defer func() {
		DB = oldDB
		db.Close()
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &models.Job{
			Title:   fmt.Sprintf("Job %d", i),
			Company: fmt.Sprintf("Company %d", i),
			Source:  "manual",
		}
		CreateJob(job)
	}
}
