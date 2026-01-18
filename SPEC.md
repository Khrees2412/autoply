Problem Statement
Job seekers spend countless hours manually searching, filtering, and applying to jobs across multiple platforms. Existing solutions like Sorce.jobs (Tinder-style mobile app) and Simplify.jobs (browser extension) work well but lack a developer-friendly CLI/TUI interface for power users who prefer terminal-based workflows.
Goal
Build Autoply - a CLI/TUI application in Go that combines job aggregation, AI-powered matching, automated applications, and tracking in a fast, scriptable terminal interface.
Current State
Empty project directory at /Users/mac/Desktop/code/autoply. Starting from scratch.
Proposed Solution
Architecture Overview
Core Components:
CLI Layer - Command-line interface using cobra for commands
TUI Layer - Interactive terminal UI using bubbletea + lipgloss
Job Aggregator - Scrape/API integration with job boards
AI Engine - LLM integration for cover letters, resume tailoring, job matching
Storage - Local SQLite database for jobs, applications, user profile
Application Engine - Automated form filling and submission
Tech Stack
Language: Go 1.21+
CLI Framework: cobra + viper (config management)
TUI Framework: bubbletea + lipgloss + bubbles
Database: SQLite with modernc.org/sqlite or mattn/go-sqlite3
HTTP Client: Standard net/http with retry logic
AI Integration: OpenAI/Anthropic API clients
Job Scraping: chromedp for headless browser automation
Project Structure
autoply/
├── cmd/
│   ├── root.go           # Root command
│   ├── search.go         # Job search commands
│   ├── apply.go          # Application commands
│   ├── profile.go        # Profile management
│   ├── status.go         # Application tracking
│   └── tui.go           # Launch TUI interface
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # SQLite schemas and queries
│   ├── scraper/         # Job board scrapers
│   ├── ai/              # LLM integration
│   ├── matcher/         # Job matching algorithm
│   ├── applicator/      # Application automation
│   └── tui/             # Bubbletea UI components
├── pkg/
│   └── models/          # Shared data models
├── configs/
│   └── config.yaml      # Default configuration
├── go.mod
├── go.sum
├── main.go
└── README.md
Core Features
Phase 1: Foundation (MVP)
Project Setup
Initialize Go module
Set up cobra CLI structure
SQLite database schema (users, jobs, applications, resumes)
Configuration system (store API keys, preferences)
Profile Management
autoply init - Interactive setup wizard
autoply profile set --name "value" - Update profile fields
autoply resume add <file> - Upload/store resumes
Store: personal info, skills, experience, preferences
Job Search (Manual Entry)
autoply job add --url <url> - Manually add job postings
Parse job URL to extract company, title, description
Store in local database
autoply job list - View saved jobs
AI Cover Letter Generation
autoply generate cover-letter <job-id> - Generate tailored cover letter
Use OpenAI/Anthropic API with job description + user profile
Save generated content for review/editing
Application Tracking
autoply status - View all applications with status
autoply status update <job-id> --status <status> - Update manually
Track: applied date, status (applied/interview/rejected/offer), notes
Phase 2: Automation
Job Board Integration
Integrate with LinkedIn, Indeed, Greenhouse, Lever APIs/scraping
autoply search --query "software engineer" --location "remote"
Auto-fetch and store jobs matching criteria
Deduplication logic
Smart Filtering & Matching
AI-powered job scoring based on user profile
autoply search --auto-match - Show only high-match jobs
Filter by salary range, experience level, keywords
autoply recommend - Get AI job recommendations
Resume Tailoring
autoply resume tailor <job-id> - AI-optimize resume for specific job
Highlight relevant experience, add missing keywords
Generate multiple resume versions
Automated Application
autoply apply <job-id> - Auto-fill and submit applications
Use chromedp for browser automation
Support Easy Apply (LinkedIn), quick apply forms
Handle common application questions with stored answers
Phase 3: TUI & Advanced Features
Interactive TUI
autoply tui - Launch full TUI interface
Views: Job browser (swipe/browse), application tracker, analytics
Keyboard navigation (j/k scroll, enter select, y/n swipe)
Split panes: job list + details
Batch Operations
autoply apply --batch - Apply to multiple jobs
autoply search --save-query - Save search queries
Cron-friendly commands for daily job checks
Analytics & Insights
autoply stats - Response rate, time-to-interview, etc.
Track which resume versions perform best
Identify common rejection patterns
Notifications
Desktop notifications for new matching jobs
Email integration to track responses
Reminder for follow-ups
Database Schema
users (singleton config)
id, name, email, phone, location, linkedin_url, github_url
preferences (JSON: desired_roles, locations, salary_min, remote_only)
resumes
id, name, file_path, content_text, created_at, is_default
skills
id, user_id, skill_name, proficiency_level
experience
id, user_id, company, title, description, start_date, end_date
jobs
id, title, company, location, url, description, salary_range
source, posted_date, scraped_at, match_score
applications
id, job_id, resume_id, cover_letter, applied_at
status (pending/applied/interview/rejected/offer)
notes, follow_up_date
cover_letters
id, job_id, content, generated_at, is_sent
CLI Command Structure
autoply
├── init                          # Setup wizard
├── profile
│   ├── show                      # Display profile
│   ├── edit                      # Interactive edit
│   └── set --field value         # Update field
├── resume
│   ├── add <file>                # Upload resume
│   ├── list                      # Show all resumes
│   └── tailor <job-id>           # AI-optimize for job
├── search
│   ├── --query --location        # Search jobs
│   ├── --auto-match              # Only high-scoring jobs
│   └── --save-query <name>       # Save search
├── job
│   ├── add --url <url>           # Manual add
│   ├── list                      # View all jobs
│   ├── show <job-id>             # Job details
│   └── remove <job-id>           # Delete job
├── apply
│   ├── <job-id>                  # Apply to job
│   ├── --batch <file>            # Batch apply
│   └── --dry-run                 # Preview without applying
├── generate
│   └── cover-letter <job-id>     # Generate cover letter
├── status                        # View applications
│   ├── --pending                 # Filter by status
│   └── update <job-id>           # Update status
├── stats                         # Analytics
├── tui                           # Launch TUI
└── config
    ├── show                      # Display config
    └── set --key value           # Update config
Key Implementation Details
AI Integration
Use structured prompts with job description + user profile
Implement retry logic and error handling
Cache LLM responses to reduce API costs
Allow custom prompts via config
Support multiple providers (OpenAI, Anthropic, local LLMs)
Job Scraping
Respectful scraping with rate limiting
User-agent rotation
Handle pagination and dynamic content
Fallback to manual entry if scraping fails
Store raw HTML for re-parsing
Browser Automation
Use chromedp for headless Chrome control
Handle different application form types
Support file uploads (resume, cover letter)
Screenshot on errors for debugging
Cookie/session management for authenticated sites
Configuration
Store at ~/.autoply/config.yaml
API keys for OpenAI/Anthropic
Job board credentials (LinkedIn, etc.)
User preferences and defaults
Privacy: never commit config to git
Data Privacy
All data stored locally (SQLite at ~/.autoply/autoply.db)
Encrypted storage for sensitive data (API keys, passwords)
Optional cloud sync (future feature)
Easy data export
Development Phases
Week 1-2: Foundation
Project structure, cobra CLI setup
SQLite database and models
Config management
Basic profile commands
Week 3-4: Core Features
Job manual entry and storage
AI cover letter generation
Application tracking
Basic search and filtering
Week 5-6: Automation
Job board scraping/API integration
Browser automation for applications
Resume tailoring
Matching algorithm
Week 7-8: TUI & Polish
Bubbletea TUI implementation
Analytics and stats
Testing and bug fixes
Documentation
Success Metrics
Time saved per application (target: 80% reduction)
Response rate improvement (target: 25% better than manual)
User satisfaction with AI-generated content
Number of applications submitted per day
