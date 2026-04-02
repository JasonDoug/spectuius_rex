package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"vibe-scaffold-tui/api"
	"vibe-scaffold-tui/ui"

	"github.com/76creates/stickers/flexbox"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

type StepState int

const (
	ChatState StepState = iota
	GeneratingState
	PreviewState
	AnsweringState
)

type Step struct {
	Name             string
	Number           int
	SystemPrompt     string
	GenerationPrompt string
	ChatHistory      []api.Message
	GeneratedDoc     string
	Approved         bool
}

type optionItem struct {
	label, desc string
}

func (i optionItem) Title() string       { return i.label }
func (i optionItem) Description() string { return i.desc }
func (i optionItem) FilterValue() string { return i.label }

type MainModel struct {
	CurrentStep    int
	Steps          []Step
	StepState      StepState
	
	TextArea       textarea.Model
	Viewport       viewport.Model // For Docs
	ChatViewport   viewport.Model // For Chat history
	Progress       progress.Model
	Spinner        spinner.Model
	OptionList     list.Model
	FlexBox        *flexbox.FlexBox
	
	APIClient      *api.Client
	IsLoading      bool
	ErrorMessage   string
	SuccessMessage string
	
	StreamChan     chan tea.Msg
	
	Ready          bool
	Height         int
	Width          int

	pendingQuestion string
}

func NewStep(name string, number int, sysPrompt, genPrompt string) Step {
	return Step{
		Name:             name,
		Number:           number,
		SystemPrompt:     sysPrompt,
		GenerationPrompt: genPrompt,
		ChatHistory:      []api.Message{},
	}
}

type errMsg error

type chunkMsg struct {
	content string
	done    bool
}

type docChunkMsg struct {
	content string
	done    bool
}

type optionsMsg struct {
	res *api.GenerateOptionsResponse
}

type compactDelegate struct{}

func (d compactDelegate) Height() int                               { return 1 }
func (d compactDelegate) Spacing() int                              { return 0 }
func (d compactDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d compactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(optionItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("  %s", i.label)
	if index == m.Index() {
		str = ui.AssistantLabelStyle.Render("● " + i.label)
	} else {
		str = "  " + i.label
	}

	fmt.Fprint(w, str)
}

func main() {
	steps := []Step{
		NewStep("One Pager", 1, "", ""),
		NewStep("Developer Spec", 2, "", ""),
		NewStep("Prompt Plan", 3, "", ""),
		NewStep("AGENTS", 4, "", ""),
	}

	steps[0].SystemPrompt = `Ask me questions so that we can develop a product specification document for this idea.

The resulting document should answer at least (but not limited to) this set of questions:
* What problem does the app solve?
* Who is the ideal user for this app?
* What platform(s) does it live on (mobile web, mobile app, web, CLI)?
* Describe the core user experience, step-by-step.
* What are the must-have features for the MVP?
* What data will the app need to persist?
* Will it need user accounts, and will there be access controls?

The user will provide an initial description of their app.

Before you begin asking questions, plan your questions out to meet the following guidelines:
* If you can infer the answer from the initial idea input, no need to ask a question about it.
* Each set of questions builds on the questions before it.
* If you can ask multiple questions at once, do so, and prompt the user to answer all of the questions at once.
* For each question, provide your recommendation and a brief explanation.

We are building an MVP - bias your choices towards simplicity, ease of implementation, and speed.`
	steps[0].GenerationPrompt = "Now that we've wrapped up the brainstorming process, can you compile our findings into a clean, comprehensive one-pager? Include the problem, audience, ideal customer, platform, and flow information, such that we could start talking with product & engineering leadership about how this could be built."

	steps[1].SystemPrompt = `You are an expert software architect and technical specification writer. You will receive a product specification document as input. Parse it thoroughly before asking clarifying questions. Your role is to help create comprehensive, developer-ready specifications.

The technical specification must include these sections, if applicable:
- Architecture Overview
- Data Models
- API/Interface Contracts
- State Management
- Dependencies & Libraries
- Edge Cases & Boundary Conditions
- Implementation Sequence

Bias your choices towards simplicity, ease of implementation, and speed.`
	steps[1].GenerationPrompt = "Now that we’ve wrapped up the brainstorming process, can you compile our findings into a comprehensive, developer-ready specification? Include all relevant requirements, architecture choices, data handling details, error handling strategies, and a testing plan so a developer can immediately begin implementation."

	steps[2].SystemPrompt = `On this step, we're going to generate a step-by-step prompt plan. The user will now optionally provide some new details about their product. Note them, provide feedback, and wait for them to move on to the prompt plan stage.`
	steps[2].GenerationPrompt = `Draft a detailed, step-by-step blueprint for building this project. Break it down into small, iterative chunks that build on each other. Provide a series of prompts for a code-generation LLM that will implement each step in a test-driven manner. Include, after each prompt, a set of todo checkboxes that the AI agents can check off.`

	steps[3].SystemPrompt = `On this step, we're going to generate an AGENTS.md file. The user will now optionally provide some new details about what they want to be in that file. Note them, provide feedback, and wait for them to move on to the generate stage.`
	steps[3].GenerationPrompt = `Using the information in the prompt_plan & spec attached here, write a minimal AGENTS.md file to include in the repository so that agents like Codex & Claude Code can interact well. Include Repository docs, Agent responsibility, Guardrails for agents, and Testing policy.`

	ta := textarea.New()
	ta.Placeholder = "Describe your idea..."
	ta.Focus()
	ta.SetWidth(80)
	ta.SetHeight(3)

	prog := progress.New(progress.WithDefaultGradient())
	
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.LoadingStyle

	streamChan := make(chan tea.Msg)

	l := list.New([]list.Item{}, compactDelegate{}, 0, 0)
	l.Title = "Select an option"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Styles.Title = ui.AssistantLabelStyle.MarginLeft(2)

	m := MainModel{
		CurrentStep: 0,
		Steps:       steps,
		StepState:   ChatState,
		TextArea:    ta,
		Progress:    prog,
		Spinner:     s,
		OptionList:  l,
		FlexBox:     flexbox.New(0, 0),
		APIClient:   api.NewClient("http://localhost:3000"),
		StreamChan:  streamChan,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.Spinner.Tick, m.waitForActivity())
}

func (m MainModel) waitForActivity() tea.Cmd {
	return func() tea.Msg {
		return <-m.StreamChan
	}
}

func (m *MainModel) updateChatViewport() {
	sidebarWidth := 25
	contentWidth := m.Width - sidebarWidth - 6
	if contentWidth < 10 {
		contentWidth = 10
	}
	
	contentHeight := m.Height - 10
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Set OptionList size here instead of View to avoid resets
	if m.StepState == AnsweringState {
		m.OptionList.SetSize(contentWidth, contentHeight/2)
	}

	var historyStr strings.Builder
	step := m.Steps[m.CurrentStep]
	for _, msg := range step.ChatHistory {
		label := ui.UserLabelStyle.Render("YOU")
		if msg.Role == "assistant" {
			label = ui.AssistantLabelStyle.Render("AI")
		}
		// Use reflow/wordwrap for robust wrapping
		wrapped := wordwrap.String(msg.Content, contentWidth-4)
		historyStr.WriteString(label + "\n" + wrapped + "\n\n")
	}

	if m.IsLoading {
		historyStr.WriteString(m.Spinner.View() + " AI is thinking...\n\n")
	}

	m.ChatViewport.SetContent(historyStr.String())
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd      tea.Cmd
		vpCmd      tea.Cmd
		cvpCmd     tea.Cmd
		spinnerCmd tea.Cmd
		listCmd    tea.Cmd
	)

	// 1. Priority handling for KeyMsgs to prevent double-processing and manage state transitions
	if key, ok := msg.(tea.KeyMsg); ok {
		m.ErrorMessage = ""
		m.SuccessMessage = ""

		switch key.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			if m.StepState == ChatState {
				if m.TextArea.Focused() {
					m.TextArea.Blur()
				} else {
					m.TextArea.Focus()
				}
				return m, nil
			}
		case tea.KeyEsc:
			if m.StepState == PreviewState || m.StepState == AnsweringState {
				m.StepState = ChatState
				m.TextArea.Focus()
				m.updateChatViewport()
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
			if m.StepState == AnsweringState {
				m.OptionList, listCmd = m.OptionList.Update(msg)
				return m, listCmd
			}
			if m.StepState == ChatState && !m.TextArea.Focused() {
				m.ChatViewport, cvpCmd = m.ChatViewport.Update(msg)
				return m, cvpCmd
			}
			if m.StepState == PreviewState {
				m.Viewport, vpCmd = m.Viewport.Update(msg)
				return m, vpCmd
			}
		case tea.KeyEnter:
			if m.StepState == AnsweringState {
				if selected, ok := m.OptionList.SelectedItem().(optionItem); ok {
					m.submitAnswer(selected.label)
					m.StepState = ChatState
					m.TextArea.Focus()
					m.updateChatViewport()
					return m, tea.Batch(m.sendChatCmd(), m.waitForActivity())
				}
			}
			if m.StepState == ChatState && !m.IsLoading && m.TextArea.Focused() {
				input := m.TextArea.Value()
				if strings.TrimSpace(input) == "" {
					break
				}
				m.Steps[m.CurrentStep].ChatHistory = append(m.Steps[m.CurrentStep].ChatHistory, api.Message{
					Role:    "user",
					Content: input,
				})
				m.TextArea.Reset()
				m.IsLoading = true
				m.updateChatViewport()
				m.ChatViewport.GotoBottom()
				return m, tea.Batch(m.sendChatCmd(), m.waitForActivity())
			}
		case tea.KeyCtrlG:
			if m.StepState == ChatState && !m.IsLoading {
				m.StepState = GeneratingState
				m.IsLoading = true
				m.TextArea.Blur()
				m.Steps[m.CurrentStep].GeneratedDoc = ""
				return m, m.generateDocCmd()
			}
		case tea.KeyCtrlA:
			if m.StepState == PreviewState {
				m.Steps[m.CurrentStep].Approved = true
				if m.CurrentStep < len(m.Steps)-1 {
					m.CurrentStep++
					m.StepState = ChatState
					m.Ready = false
					m.SuccessMessage = "Step approved! Moving to " + m.Steps[m.CurrentStep].Name
				} else {
					m.SuccessMessage = "All steps complete! Press Ctrl+C to exit."
				}
				m.updateChatViewport()
				return m, nil
			}
		}
	}

	// 2. Handle other non-terminal messages
	switch msg := msg.(type) {
	case optionsMsg:
		if msg.res != nil && len(msg.res.Options) > 0 {
			items := make([]list.Item, len(msg.res.Options))
			for i, o := range msg.res.Options {
				items[i] = optionItem{label: o, desc: ""}
			}
			m.OptionList.SetItems(items)
			m.OptionList.Title = msg.res.Question
			m.StepState = AnsweringState
			m.IsLoading = false
			m.updateChatViewport()
			return m, m.waitForActivity()
		}
		m.IsLoading = false
		m.updateChatViewport()
		return m, m.waitForActivity()

	case chunkMsg:
		if msg.done {
			m.IsLoading = false
			m.checkForQuestions()
			m.updateChatViewport()
			m.ChatViewport.GotoBottom()
			return m, m.waitForActivity()
		}
		idx := len(m.Steps[m.CurrentStep].ChatHistory) - 1
		if idx >= 0 && m.Steps[m.CurrentStep].ChatHistory[idx].Role == "assistant" {
			m.Steps[m.CurrentStep].ChatHistory[idx].Content += msg.content
		} else {
			m.Steps[m.CurrentStep].ChatHistory = append(m.Steps[m.CurrentStep].ChatHistory, api.Message{
				Role:    "assistant",
				Content: msg.content,
			})
		}
		m.updateChatViewport()
		m.ChatViewport.GotoBottom()
		return m, m.waitForActivity()

	case docChunkMsg:
		if msg.done {
			m.IsLoading = false
			m.StepState = PreviewState
			m.Ready = false
			return m, m.waitForActivity()
		}
		m.Steps[m.CurrentStep].GeneratedDoc += msg.content
		return m, m.waitForActivity()

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.FlexBox.SetWidth(msg.Width)
		m.FlexBox.SetHeight(msg.Height)
		if !m.Ready {
			m.Viewport = viewport.New(msg.Width, msg.Height)
			m.ChatViewport = viewport.New(msg.Width, msg.Height)
			m.Ready = true
		}
		m.updateChatViewport()
		return m, nil

	case errMsg:
		m.ErrorMessage = msg.Error()
		m.IsLoading = false
		m.updateChatViewport()
		return m, m.waitForActivity()
	}

	// 3. Fallback updates for background component state (like spinner or text cursor)
	m.TextArea, taCmd = m.TextArea.Update(msg)
	m.Viewport, vpCmd = m.Viewport.Update(msg)
	m.ChatViewport, cvpCmd = m.ChatViewport.Update(msg)
	m.Spinner, spinnerCmd = m.Spinner.Update(msg)
	if m.StepState == AnsweringState {
		m.OptionList, listCmd = m.OptionList.Update(msg)
	}
	
	return m, tea.Batch(taCmd, vpCmd, cvpCmd, spinnerCmd, listCmd)
}

func (m *MainModel) checkForQuestions() {
	history := m.Steps[m.CurrentStep].ChatHistory
	if len(history) == 0 {
		return
	}
	
	lastMsg := history[len(history)-1]
	if lastMsg.Role != "assistant" {
		return
	}

	lines := strings.Split(lastMsg.Content, "\n")
	var question string
	
	// Find the LAST line that contains a question mark, 
	// then take a small window around it for context
	lastQIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "?") {
			lastQIdx = i
			break
		}
	}

	if lastQIdx != -1 {
		// Take a few lines before it if they look like they are part of the same thought
		startLine := lastQIdx
		for i := lastQIdx - 1; i >= 0 && i >= lastQIdx-2; i-- {
			if strings.TrimSpace(lines[i]) == "" {
				break
			}
			startLine = i
		}
		
		for i := startLine; i <= lastQIdx; i++ {
			question += lines[i] + " "
		}
		
		// Clean up markdown and extra whitespace
		question = strings.ReplaceAll(question, "**", "")
		question = strings.ReplaceAll(question, "__", "")
		question = strings.TrimSpace(question)
	}

	if question != "" {
		m.pendingQuestion = question
		m.IsLoading = true
		go func(q string, h []api.Message) {
			res, err := m.APIClient.GenerateOptions(api.GenerateOptionsRequest{
				QuestionText: q,
				ChatHistory:  h,
			})
			if err == nil && res != nil && len(res.Options) > 0 {
				m.StreamChan <- optionsMsg{res: res}
			} else {
				// If no options or error, ensure we stop loading so user can type
				m.StreamChan <- optionsMsg{res: nil}
			}
		}(question, history)
	}
}

func (m *MainModel) submitAnswer(answer string) {
	m.Steps[m.CurrentStep].ChatHistory = append(m.Steps[m.CurrentStep].ChatHistory, api.Message{
		Role:    "user",
		Content: answer,
	})
	m.IsLoading = true
}

func (m MainModel) View() string {
	if !m.Ready {
		return "\n  Initializing..."
	}

	// 1. Header
	appName := ui.HeaderStyle.Render(" VIBE SCAFFOLD ")
	subtitle := lipgloss.NewStyle().Foreground(ui.Gray).PaddingLeft(1).Render("AI Technical Spec Generator")
	header := lipgloss.JoinHorizontal(lipgloss.Center, appName, subtitle)

	// 2. Sidebar (Progress)
	var sidebarBuilder strings.Builder
	sidebarBuilder.WriteString(lipgloss.NewStyle().Bold(true).PaddingBottom(1).Render("PROGRESS"))
	sidebarBuilder.WriteString("\n")
	for i, step := range m.Steps {
		prefix := "  "
		style := ui.StepInactiveStyle
		if i == m.CurrentStep {
			prefix = "● "
			style = ui.StepActiveStyle
		} else if i < m.CurrentStep || step.Approved {
			prefix = "✓ "
			style = ui.StepCompletedStyle
		}
		sidebarBuilder.WriteString(style.Render(prefix + step.Name))
		sidebarBuilder.WriteString("\n")
	}
	sidebar := ui.SidebarStyle.Render(sidebarBuilder.String())

	// 3. Main Content
	var mainContent string
	step := m.Steps[m.CurrentStep]

	sidebarWidth := 25
	contentWidth := m.Width - sidebarWidth - 6
	contentHeight := m.Height - 10

	switch m.StepState {
	case ChatState, AnsweringState:
		var historyStr strings.Builder
		
		for _, msg := range step.ChatHistory {
			label := ui.UserLabelStyle.Render("YOU")
			if msg.Role == "assistant" {
				label = ui.AssistantLabelStyle.Render("AI")
			}
			// Use reflow/wordwrap for robust wrapping
			wrapped := wordwrap.String(msg.Content, contentWidth - 4)
			historyStr.WriteString(label + "\n" + wrapped + "\n\n")
		}
		
		if m.IsLoading {
			historyStr.WriteString(m.Spinner.View() + " AI is thinking...\n\n")
		}

		m.ChatViewport.Width = contentWidth
		if m.StepState == ChatState {
			m.ChatViewport.Height = contentHeight - 10
		} else {
			m.ChatViewport.Height = contentHeight / 2
		}
		
		var viewportView string
		borderStyle := lipgloss.NewStyle().
			Width(contentWidth).
			Height(m.ChatViewport.Height).
			Padding(0, 1)

		if !m.TextArea.Focused() && m.StepState == ChatState {
			viewportView = borderStyle.
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ui.Cyan).
				Render(m.ChatViewport.View())
		} else {
			viewportView = borderStyle.
				Border(lipgloss.NormalBorder()).
				BorderForeground(ui.Zinc).
				Render(m.ChatViewport.View())
		}

		if m.StepState == ChatState {
			m.TextArea.SetWidth(contentWidth - 4)
			taView := m.TextArea.View()
			taStyle := lipgloss.NewStyle().Width(contentWidth).Padding(0, 1)
			if m.TextArea.Focused() {
				taView = taStyle.Border(lipgloss.RoundedBorder()).BorderForeground(ui.Amber).Render(m.TextArea.View())
			} else {
				taView = taStyle.Border(lipgloss.NormalBorder()).BorderForeground(ui.Zinc).Render(m.TextArea.View())
			}
			mainContent = lipgloss.JoinVertical(lipgloss.Left, viewportView, taView)
		} else {
			m.OptionList.SetSize(contentWidth, contentHeight / 2)
			listView := ui.DocContainerStyle.Width(contentWidth).Render(m.OptionList.View())
			mainContent = lipgloss.JoinVertical(lipgloss.Left, viewportView, listView)
		}

	case GeneratingState:
		var genBuilder strings.Builder
		genBuilder.WriteString(ui.AssistantLabelStyle.Render("GENERATING: " + step.Name) + "\n\n")
		genBuilder.WriteString(m.Spinner.View() + " AI is writing...\n\n")
		lines := strings.Split(step.GeneratedDoc, "\n")
		var lastLines []string
		if len(lines) > 15 {
			lastLines = lines[len(lines)-15:]
		} else {
			lastLines = lines
		}
		genBuilder.WriteString(strings.Join(lastLines, "\n"))
		mainContent = genBuilder.String()

	case PreviewState:
		renderer, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(contentWidth-10),
		)
		out, _ := renderer.Render(step.GeneratedDoc)
		m.Viewport.Width = contentWidth
		m.Viewport.Height = contentHeight - 4
		m.Viewport.SetContent(out)
		mainContent = lipgloss.JoinVertical(lipgloss.Left, 
			ui.AssistantLabelStyle.Render("PREVIEW: "+step.Name),
			ui.DocContainerStyle.Width(contentWidth).Render(m.Viewport.View()),
		)
	}

	// 4. Banner
	var banner string
	if m.ErrorMessage != "" {
		banner = ui.ErrorStyle.Render("ERROR: " + m.ErrorMessage)
	} else if m.SuccessMessage != "" {
		banner = ui.SuccessStyle.Render("SUCCESS: " + m.SuccessMessage)
	}

	// 5. Footer
	var shortcuts []string
	shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Ctrl+C")+" "+ui.ShortcutDescStyle.Render("Quit"))
	if m.StepState == ChatState {
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Tab")+" "+ui.ShortcutDescStyle.Render("Focus"))
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Enter")+" "+ui.ShortcutDescStyle.Render("Send"))
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Ctrl+G")+" "+ui.ShortcutDescStyle.Render("Generate"))
	} else if m.StepState == AnsweringState {
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Enter")+" "+ui.ShortcutDescStyle.Render("Select"))
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Esc")+" "+ui.ShortcutDescStyle.Render("Back to Chat"))
	} else if m.StepState == PreviewState {
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Ctrl+A")+" "+ui.ShortcutDescStyle.Render("Approve"))
		shortcuts = append(shortcuts, ui.ShortcutStyle.Render("Esc")+" "+ui.ShortcutDescStyle.Render("Chat"))
	}
	footer := ui.FooterStyle.Render(strings.Join(shortcuts, "  "))

	// Assembly using Flexbox
	m.FlexBox.SetRows([]*flexbox.Row{
		m.FlexBox.NewRow().AddCells(
			flexbox.NewCell(1, 2).SetContent(header),
		),
		m.FlexBox.NewRow().AddCells(
			flexbox.NewCell(1, contentHeight).SetContent(sidebar),
			flexbox.NewCell(4, contentHeight).SetContent(mainContent),
		),
		m.FlexBox.NewRow().AddCells(
			flexbox.NewCell(1, 1).SetContent(banner),
		),
		m.FlexBox.NewRow().AddCells(
			flexbox.NewCell(1, 2).SetContent(footer),
		),
	})

	return ui.MainStyle.Render(m.FlexBox.Render())
}

func (m MainModel) sendChatCmd() tea.Cmd {
	return func() tea.Msg {
		go func() {
			step := m.Steps[m.CurrentStep]
			docInputs := make(map[string]string)
			nameMap := map[string]string{
				"One Pager":      "onePager",
				"Developer Spec": "devSpec",
				"Prompt Plan":    "promptPlan",
				"AGENTS":         "agentsMd",
			}

			for i := 0; i < m.CurrentStep; i++ {
				apiKey := nameMap[m.Steps[i].Name]
				docInputs[apiKey] = m.Steps[i].GeneratedDoc
			}

			req := api.ChatRequest{
				Messages:       step.ChatHistory,
				SystemPrompt:   step.SystemPrompt,
				DocumentInputs: docInputs,
			}

			err := m.APIClient.StreamChat(req, func(chunk string) {
				m.StreamChan <- chunkMsg{content: chunk}
			})
			if err != nil {
				m.StreamChan <- errMsg(err)
			} else {
				m.StreamChan <- chunkMsg{done: true}
			}
		}()
		return nil
	}
}

func (m MainModel) generateDocCmd() tea.Cmd {
	return func() tea.Msg {
		go func() {
			step := m.Steps[m.CurrentStep]
			docInputs := make(map[string]string)
			nameMap := map[string]string{
				"One Pager":      "onePager",
				"Developer Spec": "devSpec",
				"Prompt Plan":    "promptPlan",
				"AGENTS":         "agentsMd",
			}
			for i := 0; i < m.CurrentStep; i++ {
				apiKey := nameMap[m.Steps[i].Name]
				docInputs[apiKey] = m.Steps[i].GeneratedDoc
			}

			apiKey := nameMap[step.Name]

			req := api.GenerateDocRequest{
				StepName:         apiKey,
				ChatHistory:      step.ChatHistory,
				DocumentInputs:   docInputs,
				GenerationPrompt: step.GenerationPrompt,
			}

			err := m.APIClient.StreamGenerateDoc(req, func(chunk string) {
				m.StreamChan <- docChunkMsg{content: chunk}
			})
			if err != nil {
				m.StreamChan <- errMsg(err)
			} else {
				m.StreamChan <- docChunkMsg{done: true}
			}
		}()
		return nil
	}
}
