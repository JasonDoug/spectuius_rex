# CLI & TUI User Guide

Vibe Scaffold now supports a full suite of terminal-based interfaces, allowing you to generate technical specifications without ever leaving your console.

## 🎛 The TUI (Terminal User Interface)

The TUI is a professional, interactive dashboard built with Go and the Charmbracelet (Bubble Tea) ecosystem. It provides the full 4-step wizard experience.

### Getting Started
1. Start the backend: `npm run dev`
2. Run the TUI:
   ```bash
   cd tui
   go run main.go
   ```

### Navigation & Shortcuts
- **`Enter`**: Send a message to the AI.
- **`Ctrl + G`**: Trigger document generation for the current step.
- **`Ctrl + A`**: Approve the current document and move to the next step.
- **`Tab`**: Toggle focus between the Chat input and the Document Preview.
- **`↑ / ↓`**: Scroll through the generated document or chat history.
- **`Ctrl + C`**: Quit the application.

---

## ⚡️ The Demo CLI

For developers looking for a "one-shot" generation or a reference implementation in TypeScript, we provide a minimal CLI.

### Usage
```bash
npx tsx lib/cli.ts "Your app idea here"
```
This will immediately call the `AIService` and stream a Step 1 One-Pager directly to your standard output.

---

## ⚙️ Configuration

Both terminal interfaces respect the settings in your `.env.local` file.

### Custom API Providers
To use a local provider (Ollama) or a different remote (OpenRouter), update your `.env.local`:

```bash
# For Ollama
OPENAI_BASE_URL=http://localhost:11434/v1
OPENAI_MODEL=llama3 # or your preferred model

# For OpenRouter
OPENAI_BASE_URL=https://openrouter.ai/api/v1
OPENAI_API_KEY=your_openrouter_key
```

## 🛠 Troubleshooting

- **"Connection Refused"**: Ensure `npm run dev` is running. The TUI and CLI connect to the local API at `http://localhost:3000`.
- **Formatting Issues**: Ensure your terminal emulator supports TrueColor and UTF-8 for the best experience with Lipgloss and Glamour.
