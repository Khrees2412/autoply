# Autoply - Automated Job Application CLI

## Overview

Autoply is a CLI tool that automates job applications across major job platforms. It uses AI to generate tailored CVs and cover letters based on user profile data, then auto-fills and submits applications.

## Supported Job Platforms

- **Greenhouse** - `boards.greenhouse.io/*`
- **LinkedIn** - `linkedin.com/jobs/*`
- **Lever** - `jobs.lever.co/*`
- **Jobvite** - `jobs.jobvite.com/*`
- **SmartRecruiters** - `jobs.smartrecruiters.com/*`
- **Pinpoint** - `*.pinpointhq.com/*`
- **Teamtailor** - `*.teamtailor.com/*`

## Core Features

### 1. URL Processing
- Accept single or multiple job URLs via CLI arguments
- Validate URLs against supported platforms
- Queue multiple applications for batch processing
- Support reading URLs from a file (`--file urls.txt`)

### 2. Profile Management (Local DB)
Store user data in SQLite database (`~/.autoply/autoply.db`):

```
profiles
├── id (primary key)
├── name
├── email
├── phone
├── location
├── linkedin_url
├── github_url
├── portfolio_url
├── base_resume (text/markdown)
├── base_cover_letter (text/markdown)
├── preferences (JSON)
│   ├── remote_only: boolean
│   ├── min_salary: number
│   ├── preferred_locations: string[]
│   ├── excluded_companies: string[]
│   └── job_types: string[] (full-time, contract, etc.)
├── skills (JSON array)
├── experience (JSON array of work history)
├── education (JSON array)
└── created_at / updated_at
```

### 3. AI Service (Multi-Provider)
Abstract AI interface supporting multiple providers:

#### Supported Providers
| Provider | Type | Configuration |
|----------|------|---------------|
| OpenAI | Cloud | `OPENAI_API_KEY` |
| Anthropic | Cloud | `ANTHROPIC_API_KEY` |
| Ollama | Local | `OLLAMA_BASE_URL` (default: `http://localhost:11434`) |
| LMStudio | Local | `LMSTUDIO_BASE_URL` (default: `http://localhost:1234`) |

#### AI Capabilities
- **Resume Tailoring**: Rewrite resume to match job requirements
- **Cover Letter Generation**: Create personalized cover letters
- **Form Field Mapping**: Intelligently map profile data to form fields
- **Question Answering**: Generate responses to custom application questions

### 4. Job Scraping
For each platform, extract:
- Job title
- Company name
- Job description
- Required qualifications
- Application form fields
- Custom questions

### 5. Application Automation
- Parse application forms using headless browser (Playwright)
- Map user profile to form fields
- Upload generated resume (PDF export)
- Fill custom questions using AI
- Support dry-run mode (`--dry-run`)
- Save application history

## CLI Commands

```bash
# Initialize profile
autoply init

# Manage profile
autoply profile show
autoply profile edit
autoply profile import <resume.pdf>

# Configure AI provider
autoply config set ai.provider ollama
autoply config set ai.model llama3.2
autoply config set ai.baseUrl http://localhost:11434
autoply config list

# Apply to jobs
autoply apply <url>
autoply apply <url1> <url2> <url3>
autoply apply --file urls.txt
autoply apply <url> --dry-run

# View history
autoply history
autoply history --status pending
autoply history --company "Acme Corp"

# Generate documents (without applying)
autoply generate resume <url> --output resume.pdf
autoply generate cover-letter <url> --output cover.pdf
```

## Configuration

Config stored in `~/.autoply/config.json`:

```json
{
  "ai": {
    "provider": "ollama",
    "model": "llama3.2",
    "baseUrl": "http://localhost:11434",
    "temperature": 0.7
  },
  "browser": {
    "headless": true,
    "timeout": 30000
  },
  "application": {
    "autoSubmit": false,
    "saveScreenshots": true,
    "retryAttempts": 3
  }
}
```

## Architecture

```
src/
├── cli/
│   ├── index.ts          # CLI entry point (Commander.js)
│   ├── commands/
│   │   ├── init.ts
│   │   ├── profile.ts
│   │   ├── config.ts
│   │   ├── apply.ts
│   │   ├── generate.ts
│   │   └── history.ts
│   └── prompts/          # Interactive prompts (Inquirer)
├── core/
│   ├── application.ts    # Application orchestrator
│   ├── queue.ts          # Job queue for batch processing
│   └── document.ts       # PDF generation
├── ai/
│   ├── provider.ts       # AI provider interface
│   ├── providers/
│   │   ├── openai.ts
│   │   ├── anthropic.ts
│   │   ├── ollama.ts
│   │   └── lmstudio.ts
│   ├── resume.ts         # Resume tailoring logic
│   └── cover-letter.ts   # Cover letter generation
├── scrapers/
│   ├── base.ts           # Base scraper class
│   ├── greenhouse.ts
│   ├── linkedin.ts
│   ├── lever.ts
│   ├── jobvite.ts
│   ├── smartrecruiters.ts
│   ├── pinpoint.ts
│   └── teamtailor.ts
├── db/
│   ├── index.ts          # Database connection
│   ├── migrations/
│   └── repositories/
│       ├── profile.ts
│       ├── application.ts
│       └── config.ts
├── utils/
│   ├── url-parser.ts     # URL validation & platform detection
│   ├── pdf.ts            # PDF utilities
│   └── logger.ts
└── types/
    └── index.ts          # TypeScript interfaces
```

## Data Flow

```
1. User runs: autoply apply https://boards.greenhouse.io/company/jobs/123

2. URL Parser
   └── Validates URL
   └── Detects platform (Greenhouse)

3. Job Scraper (Platform-specific)
   └── Launches headless browser
   └── Extracts job details & form structure
   └── Returns structured job data

4. AI Service
   └── Receives: job data + user profile
   └── Generates: tailored resume + cover letter
   └── Returns: generated documents

5. Form Filler
   └── Maps profile fields to form inputs
   └── Uploads generated resume
   └── Fills custom questions via AI
   └── Takes screenshot (if enabled)

6. Submission (if autoSubmit enabled)
   └── Submits application
   └── Saves to history database

7. History
   └── Logs application with status
   └── Stores generated documents
```

## Database Schema

### profiles
| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key |
| name | TEXT | Full name |
| email | TEXT | Email address |
| phone | TEXT | Phone number |
| location | TEXT | City, Country |
| linkedin_url | TEXT | LinkedIn profile |
| github_url | TEXT | GitHub profile |
| portfolio_url | TEXT | Portfolio website |
| base_resume | TEXT | Base resume (markdown) |
| base_cover_letter | TEXT | Base cover letter template |
| preferences | JSON | Job preferences |
| skills | JSON | Array of skills |
| experience | JSON | Work history |
| education | JSON | Education history |
| created_at | DATETIME | Creation timestamp |
| updated_at | DATETIME | Last update |

### applications
| Column | Type | Description |
|--------|------|-------------|
| id | INTEGER | Primary key |
| profile_id | INTEGER | FK to profiles |
| url | TEXT | Job posting URL |
| platform | TEXT | Platform name |
| company | TEXT | Company name |
| job_title | TEXT | Position title |
| status | TEXT | pending/submitted/failed |
| generated_resume | TEXT | Tailored resume |
| generated_cover_letter | TEXT | Generated cover letter |
| form_data | JSON | Submitted form data |
| error_message | TEXT | Error if failed |
| applied_at | DATETIME | Submission time |
| created_at | DATETIME | Creation timestamp |

### config
| Column | Type | Description |
|--------|------|-------------|
| key | TEXT | Config key (primary) |
| value | TEXT | Config value (JSON) |

## Testing Strategy

### Unit Tests (Critical Services)

```
tests/
├── ai/
│   ├── provider.test.ts      # AI provider interface
│   ├── ollama.test.ts        # Ollama integration
│   ├── lmstudio.test.ts      # LMStudio integration
│   ├── resume.test.ts        # Resume generation
│   └── cover-letter.test.ts  # Cover letter generation
├── scrapers/
│   ├── url-parser.test.ts    # URL validation
│   ├── greenhouse.test.ts    # Greenhouse scraper
│   ├── linkedin.test.ts      # LinkedIn scraper
│   ├── lever.test.ts         # Lever scraper
│   └── ...
├── db/
│   ├── profile.test.ts       # Profile CRUD
│   └── application.test.ts   # Application history
├── core/
│   ├── application.test.ts   # Application flow
│   └── queue.test.ts         # Batch processing
└── utils/
    └── pdf.test.ts           # PDF generation
```

### Test Coverage Requirements
- AI Provider Interface: 90%+
- URL Parser: 100%
- Database Repositories: 90%+
- Core Application Logic: 85%+

### Integration Tests
- End-to-end application flow (dry-run mode)
- AI provider connectivity
- Database migrations

## Dependencies

```json
{
  "dependencies": {
    "commander": "^12.0.0",
    "inquirer": "^9.0.0",
    "playwright": "^1.40.0",
    "better-sqlite3": "^9.0.0",
    "pdf-lib": "^1.17.0",
    "marked": "^11.0.0",
    "puppeteer-html-pdf": "^4.0.0",
    "zod": "^3.22.0",
    "chalk": "^5.3.0",
    "ora": "^8.0.0",
    "openai": "^4.0.0",
    "@anthropic-ai/sdk": "^0.10.0"
  },
  "devDependencies": {
    "bun-types": "latest",
    "typescript": "^5.0.0"
  }
}
```

## Security Considerations

1. **Credential Storage**: Never store passwords; use session tokens where possible
2. **API Keys**: Store in environment variables, not config files
3. **Data Privacy**: All data stored locally; no external telemetry
4. **Browser Sessions**: Clear cookies/storage after each application
5. **Rate Limiting**: Respect platform rate limits to avoid bans

## Error Handling

| Error Type | Handling |
|------------|----------|
| Invalid URL | Show error, suggest correct format |
| Unsupported Platform | List supported platforms |
| Scraping Failed | Retry with exponential backoff |
| AI Generation Failed | Fall back to base resume/cover letter |
| Form Submission Failed | Save progress, allow retry |
| Network Error | Queue for later retry |

## Future Enhancements

1. **Browser Extension**: Quick-apply from job listing pages
2. **Application Tracking**: Kanban-style job application tracker
3. **Resume Versioning**: Track multiple resume versions
4. **Analytics Dashboard**: Application success rates
5. **Email Integration**: Track responses from companies
6. **More Platforms**: Indeed, Workday, iCIMS, etc.

## Development Milestones

### Phase 1: Foundation
- [ ] Project setup & CLI scaffolding
- [ ] Database schema & migrations
- [ ] Profile management commands
- [ ] Configuration system

### Phase 2: AI Integration
- [ ] AI provider interface
- [ ] Ollama provider implementation
- [ ] LMStudio provider implementation
- [ ] OpenAI/Anthropic providers
- [ ] Resume tailoring
- [ ] Cover letter generation

### Phase 3: Job Scrapers
- [ ] Base scraper class
- [ ] Greenhouse scraper
- [ ] Lever scraper
- [ ] LinkedIn scraper
- [ ] Additional platform scrapers

### Phase 4: Application Engine
- [ ] Form detection & mapping
- [ ] Document upload handling
- [ ] Custom question answering
- [ ] Submission logic
- [ ] History tracking

### Phase 5: Polish
- [ ] Comprehensive error handling
- [ ] Progress indicators
- [ ] Dry-run mode
- [ ] Batch processing
- [ ] Documentation
