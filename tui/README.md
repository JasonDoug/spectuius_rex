# Vibe Scaffold TUI

A professional, modern Terminal User Interface for Vibe Scaffold, built with Go and the Charm Bracelet family of libraries (`bubbletea`, `bubbles`, `lipgloss`, `glamour`).

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- The Vibe Scaffold Next.js server must be running locally at `http://localhost:3000`.

## Getting Started

1. Start the Next.js development server:
   ```bash
   npm run dev
   ```

2. In a new terminal, navigate to the `tui/` directory:
   ```bash
   cd tui
   ```

3. Run the TUI:
   ```bash
   go run main.go
   ```

## Controls

- **[Enter]**: Send message to the AI.
- **[Ctrl+G]**: Generate the document for the current step.
- **[Ctrl+A]**: Approve the generated document and move to the next step.
- **[Esc]**: Go back to chat from preview, or quit the application.
- **[Ctrl+C]**: Quit the application.

## Architecture

- **`main.go`**: Entry point and main Bubble Tea loop.
- **`api/`**: HTTP client for interacting with the Next.js API.
- **`model/`**: State management and data structures.
- **`ui/`**: Lipgloss styles and UI components.

The TUI proxies all AI operations through the local Next.js API routes (`/api/chat` and `/api/generate-doc`), ensuring consistency with the web version while providing a high-performance terminal experience.
