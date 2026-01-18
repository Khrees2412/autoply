package models

import "time"

// User represents the user's profile information
type User struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Location    string    `json:"location"`
	LinkedInURL string    `json:"linkedin_url"`
	GitHubURL   string    `json:"github_url"`
	Preferences string    `json:"preferences"` // JSON string
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Resume represents a user's resume
type Resume struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	FilePath    string    `json:"file_path"`
	ContentText string    `json:"content_text"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

// Skill represents a user skill
type Skill struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id"`
	SkillName        string `json:"skill_name"`
	ProficiencyLevel string `json:"proficiency_level"` // beginner, intermediate, advanced, expert
}

// Experience represents work experience
type Experience struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Company     string    `json:"company"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"` // nullable for current positions
}

// Job represents a job posting
type Job struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Location    string    `json:"location"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
	SalaryRange string    `json:"salary_range"`
	Source      string    `json:"source"` // linkedin, indeed, manual, etc.
	PostedDate  *time.Time `json:"posted_date"`
	ScrapedAt   time.Time `json:"scraped_at"`
	MatchScore  float64   `json:"match_score"`
}

// Application represents a job application
type Application struct {
	ID           int       `json:"id"`
	JobID        int       `json:"job_id"`
	ResumeID     int       `json:"resume_id"`
	CoverLetter  string    `json:"cover_letter"`
	Status       string    `json:"status"` // pending, applied, interview, rejected, offer
	AppliedAt    time.Time `json:"applied_at"`
	Notes        string    `json:"notes"`
	FollowUpDate *time.Time `json:"follow_up_date"`
}

// CoverLetter represents a generated cover letter
type CoverLetter struct {
	ID          int       `json:"id"`
	JobID       int       `json:"job_id"`
	Content     string    `json:"content"`
	GeneratedAt time.Time `json:"generated_at"`
	IsSent      bool      `json:"is_sent"`
}

// UserPreferences represents user job search preferences
type UserPreferences struct {
	DesiredRoles  []string `json:"desired_roles"`
	Locations     []string `json:"locations"`
	SalaryMin     int      `json:"salary_min"`
	RemoteOnly    bool     `json:"remote_only"`
	ExperienceLevel string `json:"experience_level"` // entry, mid, senior, lead, principal
}
