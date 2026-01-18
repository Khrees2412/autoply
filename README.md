# Autoply

AI-powered job application automation CLI/TUI built in Go. Autoply helps you streamline your job search by aggregating jobs, generating personalized cover letters, and tracking applications—all from your terminal.

## Features

- 📝 **Profile Management**: Store your information, skills, and experience
- 💼 **Job Tracking**: Add and manage job postings from any source
- 🤖 **AI Cover Letters**: Generate personalized cover letters using OpenAI or Anthropic
- 📊 **Application Tracking**: Track application status (pending, applied, interview, offer, rejected)
- 📄 **Resume Management**: Store and manage multiple resume versions
- 🎨 **Beautiful CLI**: Clean, colorful terminal interface with lipgloss styling

## Installation

### Prerequisites

- Go 1.21 or higher
- An OpenAI or Anthropic API key (for AI features)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/khrees2412/autoply.git
cd autoply

# Build the binary
go build -o autoply

# Optionally, install globally
go install
```

## Quick Start

### 1. Initialize Your Profile

```bash
autoply init
```

This will guide you through an interactive wizard to set up your profile with:
- Name, email, phone
- Location
- LinkedIn and GitHub URLs

### 2. Configure AI API Key

```bash
# For OpenAI
autoply config set --key openai_key --value sk-your-api-key

# For Anthropic
autoply config set --key ai_provider --value anthropic
autoply config set --key anthropic_key --value sk-ant-your-api-key
```

### 3. Add a Resume

```bash
autoply resume add ~/Documents/resume.pdf --default
```

### 4. Add Jobs

```bash
# From URL (attempts to auto-parse)
autoply job add --url https://jobs.lever.co/company/job-id

# Manual entry
autoply job add --title "Software Engineer" --company "Acme Inc" --location "Remote"
```

### 5. Generate Cover Letter

```bash
# View all jobs
autoply job list

# Generate cover letter for a specific job
autoply generate cover-letter 1 --save
```

### 6. Track Applications

```bash
# Mark as applied
autoply apply 1 --notes "Applied via company website"

# View all applications
autoply status

# Update status
autoply status update 1 --status interview --notes "Phone screen scheduled"
```

## Commands

### Profile Management

```bash
autoply init                               # Create your profile
autoply profile show                       # View your profile
autoply profile set --name "John Doe"      # Update profile fields
```

### Job Management

```bash
autoply job add --url <url>                # Add job from URL
autoply job add --title "..." --company    # Add job manually
autoply job list                           # List all jobs
autoply job show <id>                      # View job details
autoply job remove <id>                    # Remove a job
```

### Resume Management

```bash
autoply resume add <file>                  # Add a resume
autoply resume add <file> --default        # Set as default
autoply resume list                        # List all resumes
```

### AI Generation

```bash
autoply generate cover-letter <job-id>     # Generate cover letter
autoply generate cover-letter <job-id> --save  # Save to database
```

### Application Tracking

```bash
autoply apply <job-id>                     # Mark as applied
autoply status                             # View all applications
autoply status --filter applied            # Filter by status
autoply status update <job-id> --status interview  # Update status
```

### Configuration

```bash
autoply config show                        # View configuration
autoply config set --key <key> --value <v> # Update config
```

**Available config keys:**
- `openai_key`: Your OpenAI API key
- `anthropic_key`: Your Anthropic API key
- `ai_provider`: `openai` or `anthropic`
- `default_model`: Model to use (e.g., `gpt-4`, `gpt-4o`)

## Configuration

Configuration is stored at `~/.autoply/config.yaml`:

```yaml
ai_provider: openai
default_model: gpt-4
openai_key: "sk-..."
anthropic_key: ""
```

## Data Storage

All data is stored locally:
- **Database**: `~/.autoply/autoply.db` (SQLite)
- **Resumes**: `~/.autoply/resumes/`
- **Config**: `~/.autoply/config.yaml`

## Example Workflow

```bash
# 1. Set up profile
autoply init

# 2. Configure AI
autoply config set --key openai_key --value sk-...

# 3. Add resume
autoply resume add ~/resume.pdf --default

# 4. Add some jobs
autoply job add --url https://jobs.lever.co/company/backend-engineer
autoply job add --url https://boards.greenhouse.io/company/frontend-role

# 5. View jobs
autoply job list

# 6. Generate cover letter for job #1
autoply generate cover-letter 1 --save

# 7. Mark as applied
autoply apply 1 --notes "Applied with generated cover letter"

# 8. Check status
autoply status

# 9. Update when you hear back
autoply status update 1 --status interview --notes "Phone screen next Tuesday"
```

## Architecture

```
autoply/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command & initialization
│   ├── profile.go         # Profile management
│   ├── job.go             # Job management
│   ├── resume.go          # Resume management
│   ├── generate.go        # AI generation
│   ├── status.go          # Application tracking
│   └── config.go          # Configuration
├── internal/
│   ├── config/            # Config management
│   ├── database/          # SQLite operations
│   ├── ai/                # AI client (OpenAI/Anthropic)
│   ├── scraper/           # Job board scrapers (future)
│   ├── matcher/           # Job matching (future)
│   └── applicator/        # Auto-apply (future)
├── pkg/
│   └── models/            # Data models
└── main.go                # Entry point
```

## Roadmap

### Phase 1: MVP ✅
- [x] Profile management
- [x] Job tracking
- [x] AI cover letter generation
- [x] Application tracking
- [x] Resume management

### Phase 2: Automation (Coming Soon)
- [ ] Job board API integration (LinkedIn, Indeed, Greenhouse)
- [ ] Automated job search and matching
- [ ] Resume tailoring per job
- [ ] Browser automation for applications

### Phase 3: TUI & Advanced Features
- [ ] Interactive TUI with bubbletea
- [ ] Batch operations
- [ ] Analytics and insights
- [ ] Email integration
- [ ] Desktop notifications

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Privacy

All your data is stored locally on your machine. API keys are stored in `~/.autoply/config.yaml` with restricted permissions (0600). Never commit your config file to version control.

## Credits

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [SQLite](https://www.sqlite.org/) - Local database

## Support

For issues and questions, please open an issue on GitHub.

---

Made with ❤️ for job seekers everywhere
