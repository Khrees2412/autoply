package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/khrees2412/autoply/internal/config"
	"github.com/khrees2412/autoply/pkg/models"
)

// GenerateCoverLetter generates a cover letter for a job using AI
func GenerateCoverLetter(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) (string, error) {
	provider := config.AppConfig.AIProvider

	switch provider {
	case "openai":
		return generateWithOpenAI(job, user, skills, experiences)
	case "anthropic":
		return generateWithAnthropic(job, user, skills, experiences)
	case "ollama":
		return generateWithOllama(job, user, skills, experiences)
	case "lmstudio":
		return generateWithLMStudio(job, user, skills, experiences)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// buildPrompt creates the prompt for cover letter generation
func buildPrompt(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) string {
	skillsList := []string{}
	for _, skill := range skills {
		skillsList = append(skillsList, skill.SkillName)
	}

	expList := []string{}
	for _, exp := range experiences {
		expList = append(expList, fmt.Sprintf("%s at %s", exp.Title, exp.Company))
	}

	prompt := fmt.Sprintf(`Generate a professional cover letter for the following job application.

Job Details:
- Title: %s
- Company: %s
- Location: %s
- Description: %s

Applicant Details:
- Name: %s
- Email: %s
- Location: %s
- Skills: %s
- Experience: %s

Write a compelling, personalized cover letter that:
1. Demonstrates enthusiasm for the role and company
2. Highlights relevant skills and experience from the applicant's background
3. Shows understanding of the job requirements
4. Is professional yet engaging
5. Is 3-4 paragraphs long
6. Does not include placeholders like [Your Name] or [Date]

Return only the cover letter text, no additional commentary.`,
		job.Title,
		job.Company,
		job.Location,
		job.Description,
		user.Name,
		user.Email,
		user.Location,
		strings.Join(skillsList, ", "),
		strings.Join(expList, "; "),
	)

	return prompt
}

// generateWithOpenAI generates a cover letter using OpenAI API
func generateWithOpenAI(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) (string, error) {
	apiKey := config.AppConfig.OpenAIKey
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured. Run: autoply config set --key openai_key --value YOUR_KEY")
	}

	prompt := buildPrompt(job, user, skills, experiences)
	model := config.AppConfig.DefaultModel
	if model == "" {
		model = "gpt-4"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  1000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format from OpenAI")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	return strings.TrimSpace(content), nil
}

// generateWithAnthropic generates a cover letter using Anthropic API
func generateWithAnthropic(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) (string, error) {
	apiKey := config.AppConfig.AnthropicKey
	if apiKey == "" {
		return "", fmt.Errorf("Anthropic API key not configured. Run: autoply config set --key anthropic_key --value YOUR_KEY")
	}

	prompt := buildPrompt(job, user, skills, experiences)

	reqBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 1024,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("unexpected response format from Anthropic")
	}

	contentBlock := content[0].(map[string]interface{})
	text := contentBlock["text"].(string)

	return strings.TrimSpace(text), nil
}

// TailorResume generates a tailored resume for a specific job
func TailorResume(resume *models.Resume, job *models.Job, user *models.User) (string, error) {
	provider := config.AppConfig.AIProvider

	switch provider {
	case "openai":
		return tailorWithOpenAI(resume, job, user)
	case "anthropic":
		return tailorWithAnthropic(resume, job, user)
	case "ollama":
		return tailorWithOllama(resume, job, user)
	case "lmstudio":
		return tailorWithLMStudio(resume, job, user)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// buildTailorPrompt creates the prompt for resume tailoring
func buildTailorPrompt(resume *models.Resume, job *models.Job, user *models.User) string {
	prompt := fmt.Sprintf(`Optimize the following resume for this specific job posting.

Job Details:
- Title: %s
- Company: %s
- Location: %s
- Description: %s

Current Resume:
%s

Applicant Information:
- Name: %s
- Email: %s
- Location: %s

Instructions:
1. Highlight relevant experience and skills that match the job requirements
2. Add missing keywords from the job description naturally
3. Reorder sections to emphasize most relevant qualifications
4. Keep the resume professional and truthful
5. Maintain the original structure but optimize content
6. Do not fabricate experience or skills

Return the optimized resume content.`,
		job.Title,
		job.Company,
		job.Location,
		job.Description,
		resume.ContentText,
		user.Name,
		user.Email,
		user.Location,
	)

	return prompt
}

// tailorWithOpenAI tailors a resume using OpenAI
func tailorWithOpenAI(resume *models.Resume, job *models.Job, user *models.User) (string, error) {
	apiKey := config.AppConfig.OpenAIKey
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	prompt := buildTailorPrompt(resume, job, user)
	model := config.AppConfig.DefaultModel
	if model == "" {
		model = "gpt-4"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format from OpenAI")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	return strings.TrimSpace(content), nil
}

// tailorWithAnthropic tailors a resume using Anthropic
func tailorWithAnthropic(resume *models.Resume, job *models.Job, user *models.User) (string, error) {
	apiKey := config.AppConfig.AnthropicKey
	if apiKey == "" {
		return "", fmt.Errorf("Anthropic API key not configured")
	}

	prompt := buildTailorPrompt(resume, job, user)

	reqBody := map[string]interface{}{
		"model":      "claude-3-5-sonnet-20241022",
		"max_tokens": 2048,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		return "", fmt.Errorf("unexpected response format from Anthropic")
	}

	contentBlock := content[0].(map[string]interface{})
	text := contentBlock["text"].(string)

	return strings.TrimSpace(text), nil
}
// generateWithOllama generates a cover letter using Ollama API
func generateWithOllama(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) (string, error) {
	url := config.AppConfig.OllamaURL
	if url == "" {
		url = "http://localhost:11434"
	}

	prompt := buildPrompt(job, user, skills, experiences)
	model := config.AppConfig.DefaultModel
	if model == "" {
		model = "llama3.2"
	}

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Ollama")
	}

	return strings.TrimSpace(response), nil
}

// generateWithLMStudio generates a cover letter using LMStudio API
func generateWithLMStudio(job *models.Job, user *models.User, skills []*models.Skill, experiences []*models.Experience) (string, error) {
	url := config.AppConfig.LMStudioURL
	if url == "" {
		url = "http://localhost:1234"
	}

	prompt := buildPrompt(job, user, skills, experiences)
	model := config.AppConfig.DefaultModel
	if model == "" {
		model = "local-model"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  1000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("LMStudio API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format from LMStudio")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	return strings.TrimSpace(content), nil
}

// tailorWithOllama tailors a resume using Ollama
func tailorWithOllama(resume *models.Resume, job *models.Job, user *models.User) (string, error) {
	url := config.AppConfig.OllamaURL
	if url == "" {
		url = "http://localhost:11434"
	}

	prompt := buildTailorPrompt(resume, job, user)
	model := config.AppConfig.DefaultModel
	if model == "" {
		model = "llama3.2"
	}

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Ollama API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Ollama")
	}

	return strings.TrimSpace(response), nil
}

// tailorWithLMStudio tailors a resume using LMStudio
func tailorWithLMStudio(resume *models.Resume, job *models.Job, user *models.User) (string, error) {
	url := config.AppConfig.LMStudioURL
	if url == "" {
		url = "http://localhost:1234"
	}

	prompt := buildTailorPrompt(resume, job, user)
	model := config.AppConfig.DefaultModel
	if model == "" {
		model = "local-model"
	}

	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("LMStudio API error: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("unexpected response format from LMStudio")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	return strings.TrimSpace(content), nil
}