package matcher

import (
	"strings"

	"github.com/khrees2412/autoply/pkg/models"
)

// CalculateMatchScore calculates how well a job matches a user's profile
// Returns a score between 0.0 and 1.0
func CalculateMatchScore(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) float64 {
	score := 0.0
	factors := 0

	// Factor 1: Skills match (40% weight)
	if len(skills) > 0 {
		skillScore := matchSkills(job, skills)
		score += skillScore * 0.4
		factors++
	}

	// Factor 2: Experience match (30% weight)
	if len(experiences) > 0 {
		expScore := matchExperience(job, experiences)
		score += expScore * 0.3
		factors++
	}

	// Factor 3: Location preference (15% weight)
	locationScore := matchLocation(job, user)
	score += locationScore * 0.15
	factors++

	// Factor 4: Job title keywords (15% weight)
	titleScore := matchTitle(job, experiences)
	score += titleScore * 0.15
	factors++

	// Normalize if we didn't have all factors
	if factors < 4 {
		score = score / (float64(factors) / 4.0)
	}

	return score
}

// matchSkills checks how many user skills match the job description
func matchSkills(job *models.Job, skills []*models.Skill) float64 {
	if job.Description == "" {
		return 0.5 // Neutral if no description
	}

	descLower := strings.ToLower(job.Description)
	matched := 0
	total := len(skills)

	if total == 0 {
		return 0.5
	}

	for _, skill := range skills {
		skillLower := strings.ToLower(skill.SkillName)
		if strings.Contains(descLower, skillLower) {
			matched++
		}
	}

	return float64(matched) / float64(total)
}

// matchExperience checks if user's experience matches job requirements
func matchExperience(job *models.Job, experiences []*models.Experience) float64 {
	if job.Description == "" {
		return 0.5
	}

	descLower := strings.ToLower(job.Description)
	matched := 0
	total := len(experiences)

	if total == 0 {
		return 0.5
	}

	for _, exp := range experiences {
		// Check if job title or company appears in description
		titleLower := strings.ToLower(exp.Title)
		companyLower := strings.ToLower(exp.Company)

		if strings.Contains(descLower, titleLower) || strings.Contains(descLower, companyLower) {
			matched++
		}
	}

	return float64(matched) / float64(total)
}

// matchLocation checks if job location matches user preferences
func matchLocation(job *models.Job, user *models.User) float64 {
	if job.Location == "" {
		return 0.5 // Neutral if no location specified
	}

	jobLocLower := strings.ToLower(job.Location)
	userLocLower := strings.ToLower(user.Location)

	// Check for exact match
	if strings.Contains(jobLocLower, userLocLower) || strings.Contains(userLocLower, jobLocLower) {
		return 1.0
	}

	// Check for remote
	if strings.Contains(jobLocLower, "remote") {
		return 0.8 // Slight preference for remote
	}

	// Partial match (same city/state)
	jobParts := strings.Fields(jobLocLower)
	userParts := strings.Fields(userLocLower)

	for _, jobPart := range jobParts {
		for _, userPart := range userParts {
			if len(jobPart) > 3 && len(userPart) > 3 && jobPart == userPart {
				return 0.6
			}
		}
	}

	return 0.3 // Low match if no overlap
}

// matchTitle checks if job title matches user's experience titles
func matchTitle(job *models.Job, experiences []*models.Experience) float64 {
	if job.Title == "" {
		return 0.5
	}

	jobTitleLower := strings.ToLower(job.Title)

	// Extract keywords from job title
	jobKeywords := extractKeywords(jobTitleLower)

	if len(jobKeywords) == 0 {
		return 0.5
	}

	matched := 0
	for _, exp := range experiences {
		expTitleLower := strings.ToLower(exp.Title)
		for _, keyword := range jobKeywords {
			if strings.Contains(expTitleLower, keyword) {
				matched++
				break
			}
		}
	}

	if len(experiences) == 0 {
		return 0.5
	}

	return float64(matched) / float64(len(experiences))
}

// extractKeywords extracts meaningful keywords from a job title
func extractKeywords(title string) []string {
	// Common stop words to ignore
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true,
	}

	words := strings.Fields(title)
	keywords := []string{}

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

