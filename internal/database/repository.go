package database

import (
	"database/sql"
	"time"

	"github.com/khrees2412/autoply/pkg/models"
)

// User operations

func CreateUser(user *models.User) error {
	query := `INSERT INTO users (name, email, phone, location, linkedin_url, github_url, preferences) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, user.Name, user.Email, user.Phone, user.Location, 
		user.LinkedInURL, user.GitHubURL, user.Preferences)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

func GetUser() (*models.User, error) {
	query := `SELECT id, name, email, phone, location, linkedin_url, github_url, preferences, 
			  created_at, updated_at FROM users LIMIT 1`
	user := &models.User{}
	err := DB.QueryRow(query).Scan(&user.ID, &user.Name, &user.Email, &user.Phone, 
		&user.Location, &user.LinkedInURL, &user.GitHubURL, &user.Preferences, 
		&user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func UpdateUser(user *models.User) error {
	query := `UPDATE users SET name=?, email=?, phone=?, location=?, linkedin_url=?, 
			  github_url=?, preferences=?, updated_at=? WHERE id=?`
	_, err := DB.Exec(query, user.Name, user.Email, user.Phone, user.Location, 
		user.LinkedInURL, user.GitHubURL, user.Preferences, time.Now(), user.ID)
	return err
}

// Job operations

func CreateJob(job *models.Job) error {
	query := `INSERT INTO jobs (title, company, location, url, description, salary_range, 
			  source, posted_date, match_score) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, job.Title, job.Company, job.Location, job.URL, 
		job.Description, job.SalaryRange, job.Source, job.PostedDate, job.MatchScore)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	job.ID = int(id)
	return nil
}

func GetJob(id int) (*models.Job, error) {
	query := `SELECT id, title, company, location, url, description, salary_range, 
			  source, posted_date, scraped_at, match_score FROM jobs WHERE id=?`
	job := &models.Job{}
	err := DB.QueryRow(query, id).Scan(&job.ID, &job.Title, &job.Company, &job.Location, 
		&job.URL, &job.Description, &job.SalaryRange, &job.Source, &job.PostedDate, 
		&job.ScrapedAt, &job.MatchScore)
	return job, err
}

func GetAllJobs() ([]*models.Job, error) {
	query := `SELECT id, title, company, location, url, description, salary_range, 
			  source, posted_date, scraped_at, match_score FROM jobs ORDER BY scraped_at DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []*models.Job{}
	for rows.Next() {
		job := &models.Job{}
		err := rows.Scan(&job.ID, &job.Title, &job.Company, &job.Location, &job.URL, 
			&job.Description, &job.SalaryRange, &job.Source, &job.PostedDate, 
			&job.ScrapedAt, &job.MatchScore)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func DeleteJob(id int) error {
	query := `DELETE FROM jobs WHERE id=?`
	_, err := DB.Exec(query, id)
	return err
}

func GetJobByURL(url string) (*models.Job, error) {
	query := `SELECT id, title, company, location, url, description, salary_range, 
			  source, posted_date, scraped_at, match_score FROM jobs WHERE url=?`
	job := &models.Job{}
	err := DB.QueryRow(query, url).Scan(&job.ID, &job.Title, &job.Company, &job.Location, 
		&job.URL, &job.Description, &job.SalaryRange, &job.Source, &job.PostedDate, 
		&job.ScrapedAt, &job.MatchScore)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return job, err
}

// Resume operations

func CreateResume(resume *models.Resume) error {
	// If setting as default, unset all other defaults first
	if resume.IsDefault {
		_, _ = DB.Exec("UPDATE resumes SET is_default=0")
	}
	
	query := `INSERT INTO resumes (name, file_path, content_text, is_default) 
			  VALUES (?, ?, ?, ?)`
	result, err := DB.Exec(query, resume.Name, resume.FilePath, resume.ContentText, resume.IsDefault)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	resume.ID = int(id)
	return nil
}

func SetDefaultResume(resumeID int) error {
	// Unset all defaults first
	_, err := DB.Exec("UPDATE resumes SET is_default=0")
	if err != nil {
		return err
	}
	// Set the specified resume as default
	_, err = DB.Exec("UPDATE resumes SET is_default=1 WHERE id=?", resumeID)
	return err
}

func GetAllResumes() ([]*models.Resume, error) {
	query := `SELECT id, name, file_path, content_text, is_default, created_at 
			  FROM resumes ORDER BY created_at DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resumes := []*models.Resume{}
	for rows.Next() {
		resume := &models.Resume{}
		err := rows.Scan(&resume.ID, &resume.Name, &resume.FilePath, &resume.ContentText, 
			&resume.IsDefault, &resume.CreatedAt)
		if err != nil {
			return nil, err
		}
		resumes = append(resumes, resume)
	}
	return resumes, nil
}

func GetDefaultResume() (*models.Resume, error) {
	query := `SELECT id, name, file_path, content_text, is_default, created_at 
			  FROM resumes WHERE is_default=1 LIMIT 1`
	resume := &models.Resume{}
	err := DB.QueryRow(query).Scan(&resume.ID, &resume.Name, &resume.FilePath, 
		&resume.ContentText, &resume.IsDefault, &resume.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return resume, err
}

// Application operations

func CreateApplication(app *models.Application) error {
	query := `INSERT INTO applications (job_id, resume_id, cover_letter, status, notes) 
			  VALUES (?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, app.JobID, app.ResumeID, app.CoverLetter, app.Status, app.Notes)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	app.ID = int(id)
	return nil
}

func GetAllApplications() ([]*models.Application, error) {
	query := `SELECT id, job_id, resume_id, cover_letter, status, applied_at, notes, follow_up_date 
			  FROM applications ORDER BY applied_at DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []*models.Application{}
	for rows.Next() {
		app := &models.Application{}
		var resumeID sql.NullInt64
		err := rows.Scan(&app.ID, &app.JobID, &resumeID, &app.CoverLetter, &app.Status, 
			&app.AppliedAt, &app.Notes, &app.FollowUpDate)
		if err != nil {
			return nil, err
		}
		if resumeID.Valid {
			app.ResumeID = int(resumeID.Int64)
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func UpdateApplicationStatus(id int, status string, notes string) error {
	query := `UPDATE applications SET status=?, notes=? WHERE id=?`
	_, err := DB.Exec(query, status, notes, id)
	return err
}

func GetApplicationByJobID(jobID int) (*models.Application, error) {
	query := `SELECT id, job_id, resume_id, cover_letter, status, applied_at, notes, follow_up_date 
			  FROM applications WHERE job_id=? LIMIT 1`
	app := &models.Application{}
	var resumeID sql.NullInt64
	err := DB.QueryRow(query, jobID).Scan(&app.ID, &app.JobID, &resumeID, &app.CoverLetter, 
		&app.Status, &app.AppliedAt, &app.Notes, &app.FollowUpDate)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if resumeID.Valid {
		app.ResumeID = int(resumeID.Int64)
	}
	return app, nil
}

// Cover Letter operations

func CreateCoverLetter(cl *models.CoverLetter) error {
	query := `INSERT INTO cover_letters (job_id, content, is_sent) VALUES (?, ?, ?)`
	result, err := DB.Exec(query, cl.JobID, cl.Content, cl.IsSent)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	cl.ID = int(id)
	return nil
}

func GetCoverLetterByJobID(jobID int) (*models.CoverLetter, error) {
	query := `SELECT id, job_id, content, generated_at, is_sent 
			  FROM cover_letters WHERE job_id=? ORDER BY generated_at DESC LIMIT 1`
	cl := &models.CoverLetter{}
	err := DB.QueryRow(query, jobID).Scan(&cl.ID, &cl.JobID, &cl.Content, &cl.GeneratedAt, &cl.IsSent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return cl, err
}

// Skill operations

func CreateSkill(skill *models.Skill) error {
	query := `INSERT INTO skills (user_id, skill_name, proficiency_level) VALUES (?, ?, ?)`
	result, err := DB.Exec(query, skill.UserID, skill.SkillName, skill.ProficiencyLevel)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	skill.ID = int(id)
	return nil
}

func GetUserSkills(userID int) ([]*models.Skill, error) {
	query := `SELECT id, user_id, skill_name, proficiency_level FROM skills WHERE user_id=?`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []*models.Skill{}
	for rows.Next() {
		skill := &models.Skill{}
		err := rows.Scan(&skill.ID, &skill.UserID, &skill.SkillName, &skill.ProficiencyLevel)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

func DeleteSkill(id int) error {
	query := `DELETE FROM skills WHERE id=?`
	_, err := DB.Exec(query, id)
	return err
}

// Experience operations

func CreateExperience(exp *models.Experience) error {
	query := `INSERT INTO experiences (user_id, company, title, description, start_date, end_date) 
			  VALUES (?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(query, exp.UserID, exp.Company, exp.Title, exp.Description, 
		exp.StartDate, exp.EndDate)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	exp.ID = int(id)
	return nil
}

func GetUserExperiences(userID int) ([]*models.Experience, error) {
	query := `SELECT id, user_id, company, title, description, start_date, end_date 
			  FROM experiences WHERE user_id=? ORDER BY start_date DESC`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	experiences := []*models.Experience{}
	for rows.Next() {
		exp := &models.Experience{}
		err := rows.Scan(&exp.ID, &exp.UserID, &exp.Company, &exp.Title, &exp.Description, 
			&exp.StartDate, &exp.EndDate)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, exp)
	}
	return experiences, nil
}

func DeleteExperience(id int) error {
	query := `DELETE FROM experiences WHERE id=?`
	_, err := DB.Exec(query, id)
	return err
}

// Helper function to format application data with job details
func GetApplicationsWithJobs() ([]map[string]interface{}, error) {
	query := `
		SELECT a.id, a.status, a.applied_at, a.notes, 
			   j.id, j.title, j.company, j.location, j.url
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		ORDER BY a.applied_at DESC
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		var appID, jobID int
		var status, notes, title, company, location, url string
		var appliedAt time.Time
		
		err := rows.Scan(&appID, &status, &appliedAt, &notes, &jobID, &title, &company, &location, &url)
		if err != nil {
			return nil, err
		}
		
		results = append(results, map[string]interface{}{
			"app_id":     appID,
			"status":     status,
			"applied_at": appliedAt,
			"notes":      notes,
			"job_id":     jobID,
			"title":      title,
			"company":    company,
			"location":   location,
			"url":        url,
		})
	}
	return results, nil
}

// SaveSearchQuery saves a search query for later use
func SaveSearchQuery(name, query, location, source string) error {
	// Create a simple table for saved queries if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS saved_queries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		query TEXT NOT NULL,
		location TEXT,
		source TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}

	// Insert or replace
	insertQuery := `INSERT OR REPLACE INTO saved_queries (name, query, location, source) VALUES (?, ?, ?, ?)`
	_, err = DB.Exec(insertQuery, name, query, location, source)
	return err
}

// GetSavedQueries retrieves all saved search queries
func GetSavedQueries() ([]map[string]interface{}, error) {
	query := `SELECT name, query, location, source, created_at FROM saved_queries ORDER BY created_at DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []map[string]interface{}{}
	for rows.Next() {
		var name, query, location, source string
		var createdAt time.Time
		err := rows.Scan(&name, &query, &location, &source, &createdAt)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"name":       name,
			"query":      query,
			"location":   location,
			"source":     source,
			"created_at": createdAt,
		})
	}
	return results, nil
}
