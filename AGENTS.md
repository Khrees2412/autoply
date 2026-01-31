# Autoply Development Guide

## Commands

```bash
# Development
bun install              # Install dependencies
bun run dev              # Run CLI in development mode
bun run start            # Alias for dev

# Testing
bun test                 # Run all tests
bun test --watch         # Run tests in watch mode
bun test <file>          # Run specific test file

# Type checking
bun run typecheck        # Run TypeScript type checker

# Building
bun run build            # Build for current platform
bun run build:all        # Build for all platforms
bun run build:mac        # Build for macOS ARM
bun run build:mac-intel  # Build for macOS Intel
bun run build:linux      # Build for Linux x64
bun run build:windows    # Build for Windows x64

# Linting
bun run lint             # Run ESLint
bun run lint:fix         # Run ESLint with auto-fix
bun run format           # Run Prettier
bun run format:check     # Check Prettier formatting
```

## Project Structure

```
src/
├── ai/                  # AI provider integrations (OpenAI, Anthropic, etc.)
├── cli/                 # CLI commands and prompts
│   ├── commands/        # Individual command implementations
│   └── prompts/         # Interactive prompts
├── core/                # Core business logic
│   ├── application.ts   # Application orchestrator
│   ├── document.ts      # PDF generation
│   ├── form-filler/     # Form filling logic (split into modules)
│   └── queue.ts         # Job queue management
├── db/                  # Database layer
│   └── repositories/    # Data access objects
├── scrapers/            # Platform-specific scrapers
├── types/               # TypeScript type definitions
└── utils/               # Utility functions
```

## Code Style

- Use TypeScript strict mode
- Prefer `async/await` over raw promises
- Use Zod for runtime validation
- Keep functions small and focused
- Extract magic strings to constants
- Add JSDoc comments to public APIs

## Testing

- Test files should be colocated with source files as `*.test.ts`
- Use `describe` blocks to group related tests
- Mock external dependencies (AI providers, browser, etc.)
- Integration tests go in `*.integration.test.ts` files

## Environment Variables

Required for cloud AI providers:
- `ANTHROPIC_API_KEY` - For Anthropic/Claude
- `OPENAI_API_KEY` - For OpenAI/GPT
- `GOOGLE_API_KEY` - For Google/Gemini
