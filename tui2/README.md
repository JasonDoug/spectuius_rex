# Vibe Scaffold TUI 2.0

A modern Terminal User Interface for the Vibe Scaffold AI Technical Spec Generator, built with the `tview` framework and modeled after the `chiko` aesthetic.

## Features

- **Component-Based UI:** Sidebar for navigation, main chat area, and dedicated input field.
- **Dynamic Streaming:** Real-time AI response streaming with visual feedback.
- **Multiple Choice Selector:** Modal-style overlay for answering clarifying questions.
- **Document Preview:** Full-screen preview mode for generated specifications with approval workflow.
- **IDE-Like Layout:** Flex-based structural grid with high-contrast borders and professional color palette.

## Prerequisites

- Go 1.25+
- Local Next.js development server running (`npm run dev`)

## Getting Started

### Build

```bash
cd tui2
go build -o vibe-scaffold-tui2 main.go
```

### Run

```bash
./vibe-scaffold-tui2
```

## Keybindings

- **Tab:** Toggle focus between Sidebar and Input Field.
- **Enter:** Send message / Select option.
- **Ctrl+G:** Generate the technical document for the current step.
- **Ctrl+A:** Approve the generated document (in Preview mode).
- **Esc:** Exit Preview mode or Options selector / Quit application.
- **Ctrl+C:** Force quit.
