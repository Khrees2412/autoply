package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Initialize creates and opens the SQLite database
func Initialize() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .autoply directory if it doesn't exist
	autoplyDir := filepath.Join(homeDir, ".autoply")
	if err := os.MkdirAll(autoplyDir, 0755); err != nil {
		return fmt.Errorf("failed to create autoply directory: %w", err)
	}

	dbPath := filepath.Join(autoplyDir, "autoply.db")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	DB = db

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// runMigrations creates all necessary tables
func runMigrations() error {
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
		match_score REAL DEFAULT 0
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
		FOREIGN KEY (resume_id) REFERENCES resumes(id) ON DELETE SET NULL
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

	_, err := DB.Exec(schema)
	return err
}
