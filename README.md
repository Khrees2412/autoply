# Autoply

Automated job application CLI - Apply to jobs with AI-generated resumes and cover letters.

## Installation

### From Source

```bash
# Clone and install dependencies
git clone https://github.com/khrees2412/autoply.git
cd autoply
bun install

# Install Playwright browsers (required for job scraping)
bunx playwright install chromium

# Build the executable
bun run build

# Move to your PATH (optional)
mv dist/autoply /usr/local/bin/
```

### Build for Other Platforms

```bash
bun run build:mac         # macOS ARM (Apple Silicon)
bun run build:mac-intel   # macOS Intel
bun run build:linux       # Linux x64
bun run build:windows     # Windows x64
bun run build:all         # All platforms
```

## Quick Start

```bash
# 1. Initialize your profile
autoply init

# 2. Configure your AI provider
autoply config set ai.provider anthropic
autoply config set ai.apiKey sk-ant-...

# 3. Apply to a job
autoply apply https://linkedin.com/jobs/view/123456
```

## Commands

### `autoply init`

Set up your profile with work experience, skills, and education.

```bash
autoply init
```

### `autoply apply [urls...]`

Apply to one or more jobs with AI-generated documents.

```bash
# Single job
autoply apply https://linkedin.com/jobs/view/123456

# Multiple jobs
autoply apply https://job1.com https://job2.com https://job3.com

# From a file (one URL per line)
autoply apply -f jobs.txt

# Dry run - generate documents without submitting
autoply apply -d https://linkedin.com/jobs/view/123456
```

### `autoply generate`

Generate tailored documents without applying.

```bash
# Generate a resume
autoply generate resume https://linkedin.com/jobs/view/123456

# Generate a cover letter
autoply generate cover-letter https://linkedin.com/jobs/view/123456

# Generate both
autoply generate both https://linkedin.com/jobs/view/123456
```

### `autoply profile`

Manage your profile.

```bash
autoply profile show      # Display your profile
autoply profile edit      # Edit your profile
autoply profile delete    # Delete your profile
autoply profile import resume.pdf  # Import from existing resume
```

### `autoply config`

Configure AI providers and settings.

```bash
# Set AI provider (anthropic, openai, google)
autoply config set ai.provider anthropic
autoply config set ai.apiKey sk-ant-...

# View configuration
autoply config list
autoply config get ai.provider

# Test AI connection
autoply config test

# Reset to defaults
autoply config reset
```

### `autoply history`

View your application history.

```bash
autoply history                    # View recent applications
autoply history -s submitted       # Filter by status
autoply history -c "Acme Inc"      # Filter by company
autoply history -l 50              # Show more results
autoply history show <id>          # View application details
autoply history clear              # Clear history
```

## Supported Job Platforms

- LinkedIn
- Greenhouse
- Lever
- Workday
- Ashby
- Jobvite
- SmartRecruiters
- Pinpoint
- Teamtailor

## AI Providers

Configure your preferred AI provider:

| Provider | Config Value | API Key Format |
|----------|--------------|----------------|
| Anthropic | `anthropic` | `sk-ant-...` |
| OpenAI | `openai` | `sk-...` |
| Google | `google` | `AIza...` |

```bash
autoply config providers  # List all available providers
```

## License

MIT
