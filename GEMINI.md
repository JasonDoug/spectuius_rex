# Vibe Scaffold - AI Technical Spec Generator

Vibe Scaffold is a technical specification generator that uses AI-powered chat to turn ideas into professional documentation (One Pager, Dev Spec, Prompt Plan, and AGENTS.md).

## Core Architecture

- **Wizard Engine (`lib/`)**: Decoupled business logic and AI orchestration.
  - `lib/ai-service.ts`: Core service for chat and document generation.
  - `lib/steps/`: Step configurations and AI prompts.
  - `lib/utils/`: Shared logic (ai provider, step access, samples).
  - `lib/types.ts`: Shared TypeScript interfaces.
- **Web Frontend (`app/`)**: Next.js 15 application.
  - `app/api/`: Thin transport layer routes that call `lib/ai-service`.
  - `app/wizard/`: React components for the wizard UI.
  - `app/store.ts`: Zustand state management for the web interface.
- **TUI Frontend (`tui/`)**: Go-based Terminal User Interface using `bubbletea` and `charmbracelet`. Connects to the local Next.js API.
- **CLI Demo (`lib/cli.ts`)**: Minimal TypeScript CLI for testing the decoupled `AIService`.

## Tech Stack

- **Framework**: Next.js 15.1.0 (App Router), React 19, TypeScript 5.
- **TUI**: Go 1.26+, Bubbletea, Lipgloss, Glamour.
- **AI Integration**: Vercel AI SDK (`ai` package) with OpenAI.
- **State Management**: Zustand with `localStorage` persistence (Web).
- **Runtime**: Edge Runtime for API routes.
- **Testing**: Vitest 4.0.10 with `@testing-library/react`.

## Key Commands

- `npm run dev`: Start the development server at [http://localhost:3000](http://localhost:3000).
- `go run tui/main.go`: Start the Terminal User Interface (requires dev server).
- `npm test`: Run the full Vitest test suite.
- `npm run lint`: Run ESLint.

## Development Conventions & Guidelines

### 1. Decoupled Logic
- All business logic, prompts, and AI orchestration MUST live in `lib/`.
- Frontends (Web, TUI, etc.) should be thin layers that consume `lib/` or the API.
- Any change to the wizard flow (e.g., adding a step) should start in `lib/steps/`.

### 2. API & AI
- **Edge Runtime**: All routes in `app/api/` must use `export const runtime = "edge"`.
- **Streaming**: All AI interactions use streaming for better UX.
- **Custom Base URLs**: Support for OpenRouter, Ollama, etc., via `OPENAI_BASE_URL`.

### 3. Established Patterns & Pitfalls
- **Hydration Safety**: 
  - Avoid whitespace text nodes in the `<head>` of `app/layout.tsx`.
  - Place `next/script` with `strategy="afterInteractive"` in the `<body>` rather than the `<head>`.
- **Browser Compatibility**: Use a fallback for `crypto.randomUUID()` (see `app/utils/analytics.ts`).
- **File Naming**: Generated documents follow `ALL_CAPS_UNDERSCORES.md` (e.g., `ONE_PAGER.md`).

### 4. Testing Requirements
- **Mandatory Testing**: Run `npm test` before committing.
- **Mocks**: Use global mocks in `tests/setup.ts`. Mock `@ai-sdk/openai` using the `createOpenAI` pattern.

## Directory Structure Highlights

- `app/api/`: Streaming AI endpoints (thin wrappers).
- `app/wizard/`: React wizard UI components.
- `lib/`: Core "Wizard Engine" (Logic, Steps, Types).
- `tui/`: Go-based Terminal User Interface.
- `tests/`: Comprehensive test suite.
