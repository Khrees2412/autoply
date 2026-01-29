# Autoply

AI-powered CLI that automates job applications. It scrapes job postings, generates tailored resumes and cover letters, fills out forms, and submits applications — all from your terminal.

## Installation

### Quick Install

```bash
curl -fsSL https://autoply.khrees.com/install | bash
```

### From Source

```bash
git clone https://github.com/khrees2412/autoply.git
cd autoply
bun install
bunx playwright install chromium
bun run build
mv dist/autoply /usr/local/bin/
```

> **Requires [Bun](https://bun.sh)** — install with `curl -fsSL https://bun.sh/install | bash`

## Setup

### 1. Create your profile

```bash
autoply init
```

This walks you through entering your name, contact info, skills, work experience, education, and job preferences. Everything is stored locally in `~/.autoply/autoply.db`.

### 2. Configure an AI provider

Autoply needs an AI provider to generate documents. Choose one:

**Cloud providers** (require API keys):

```bash
# Anthropic
autoply config set ai.provider anthropic
autoply config set ai.model claude-sonnet-4-5-20250929

# OpenAI
autoply config set ai.provider openai
autoply config set ai.model gpt-5.2

# Google
autoply config set ai.provider google
autoply config set ai.model gemini-pro-3
```

Set your API key as an environment variable. Add one of these to your `.env` file in the project root, or export in your shell profile:

```bash
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GOOGLE_API_KEY=AIza...
```

**Local providers** (no API key needed):

```bash
# Ollama (default)
autoply config set ai.provider ollama
autoply config set ai.model llama3.2

# LM Studio
autoply config set ai.provider lmstudio
```

Make sure the local server is running before using Autoply.

### 3. Verify it works

```bash
autoply config test
```

### 4. (Optional) Save a browser session

For platforms that require login (e.g. LinkedIn):

```bash
autoply login linkedin
```

This opens a browser — log in manually, and the session is saved for future use.

## Usage

### Apply to a job

```bash
autoply apply https://boards.greenhouse.io/company/jobs/123456
```

### Apply to multiple jobs

```bash
autoply apply https://job1.com https://job2.com https://job3.com

# Or from a file (one URL per line)
autoply apply -f jobs.txt
```

### Dry run (generate documents without submitting)

```bash
autoply apply -d https://boards.greenhouse.io/company/jobs/123456
```

### Generate documents only

```bash
autoply generate resume https://boards.greenhouse.io/company/jobs/123456
autoply generate cover-letter https://boards.greenhouse.io/company/jobs/123456
autoply generate both https://boards.greenhouse.io/company/jobs/123456 -d ./output
```

### View application history

```bash
autoply history
autoply history -s submitted
autoply history -c "Anthropic"
```

### Manage your profile

```bash
autoply profile show
autoply profile edit
autoply profile delete
```

### Configuration

```bash
autoply config list              # Show all settings
autoply config set <key> <value> # Set a value
autoply config get <key>         # Get a value
autoply config reset             # Reset to defaults
autoply config providers         # List AI providers
autoply config test              # Test AI connection
```

**All config keys:**

| Key | Default | Description |
|-----|---------|-------------|
| `ai.provider` | `ollama` | AI provider (`openai`, `anthropic`, `google`, `ollama`, `lmstudio`) |
| `ai.model` | varies | Model name |
| `ai.baseUrl` | varies | API base URL (local providers) |
| `ai.temperature` | `0.7` | Generation temperature |
| `browser.headless` | `false` | Run browser without UI |
| `browser.timeout` | `30000` | Browser timeout (ms) |
| `application.autoSubmit` | `false` | Auto-submit after generating docs |
| `application.saveScreenshots` | `true` | Save screenshots on submission |
| `application.retryAttempts` | `3` | Retry count for failed operations |

## Supported Platforms

| Platform | URL Pattern |
|----------|-------------|
| Greenhouse | `boards.greenhouse.io/*` |
| LinkedIn | `linkedin.com/jobs/*` |
| Lever | `jobs.lever.co/*` |
| Workday | `*.myworkdayjobs.com/*` |
| Ashby | `jobs.ashbyhq.com/*` |
| Jobvite | `jobs.jobvite.com/*` |
| SmartRecruiters | `jobs.smartrecruiters.com/*` |
| Pinpoint | `*.pinpointhq.com/*` |
| Teamtailor | `*.teamtailor.com/*` |

## Data Storage

All data is stored locally in `~/.autoply/`:

```
~/.autoply/
├── autoply.db           # SQLite database (profiles, applications, config)
├── config.json          # App configuration
├── browser-state.json   # Saved browser session (after login)
├── documents/           # Generated resumes and cover letters
└── screenshots/         # Submission screenshots
```

## Development

```bash
bun install
bun run dev              # Run CLI in dev mode
bun test                 # Run tests
bun run build            # Build for current platform
bun run build:all        # Build for all platforms
```

### Build targets

```bash
bun run build:mac        # macOS ARM (Apple Silicon)
bun run build:mac-intel  # macOS Intel
bun run build:linux      # Linux x64
bun run build:windows    # Windows x64
```

## License

MIT
