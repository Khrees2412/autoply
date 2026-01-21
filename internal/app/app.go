package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/khrees2412/autoply/internal/config"
	_ "github.com/mattn/go-sqlite3"
)

// App is the dependency container for the CLI application
type App struct {
	DB         *sql.DB
	Config     *config.Config
	HTTPClient *http.Client
}

// NewApp initializes and returns a new App instance
func NewApp(ctx context.Context) (*App, error) {
	// Initialize config
	if err := config.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Open database with proper pragmas
	db, err := initializeDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Verify database connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &App{
		DB:         db,
		Config:     config.AppConfig,
		HTTPClient: httpClient,
	}, nil
}

// Close closes all resources
func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}

// initializeDatabase creates and opens the SQLite database with proper settings
func initializeDatabase() (*sql.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .autoply directory if it doesn't exist
	autoplyDir := filepath.Join(homeDir, ".autoply")
	if err := os.MkdirAll(autoplyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create autoply directory: %w", err)
	}

	dbPath := filepath.Join(autoplyDir, "autoply.db")

	// Open with DSN options for SQLite pragmas
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations creates all necessary tables
func runMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT,
		phone TEXT,
		location TEXT,
		linkedin_url TEXT,
		github_url TEXT,
		preferences TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS resumes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		content_text TEXT,
		is_default BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS skills (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		skill_name TEXT NOT NULL,
		proficiency_level TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS experiences (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		company TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		start_date DATE NOT NULL,
		end_date DATE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS jobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		company TEXT NOT NULL,
		location TEXT,
		url TEXT UNIQUE,
		description TEXT,
		salary_range TEXT,
		source TEXT DEFAULT 'manual',
		posted_date DATE,
		scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		match_score REAL DEFAULT 0,
		CHECK(source IN ('manual', 'linkedin', 'indeed', 'url', 'greenhouse', 'lever'))
	);

	CREATE TABLE IF NOT EXISTS applications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		resume_id INTEGER,
		cover_letter TEXT,
		status TEXT DEFAULT 'pending',
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		notes TEXT,
		follow_up_date DATE,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
		FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE SET NULL,
		CHECK(status IN ('pending', 'applied', 'interview', 'rejected', 'offer', 'accepted'))
	);

	CREATE TABLE IF NOT EXISTS cover_letters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		generated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_sent BOOLEAN DEFAULT 0,
		FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_company ON jobs(company);
	CREATE INDEX IF NOT EXISTS idx_jobs_source ON jobs(source);
	CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
	CREATE INDEX IF NOT EXISTS idx_applications_job_id ON applications(job_id);
	`

	_, err := db.Exec(schema)
	return err
}
