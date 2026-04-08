package main

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"vibe-scaffold-tui2/api"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type StepState int

const (
	StrategyState StepState = iota
	ChatState
	GeneratingState
	PreviewState
	AnsweringState
)

type QuestionStrategy int

const (
	StrategyHybrid QuestionStrategy = iota
	StrategyMultipleChoice
	StrategyFreeForm
)

type Step struct {
	Name             string
	Filename         string
	Number           int
	SystemPrompt     string
	GenerationPrompt string
	ChatHistory      []api.Message
	GeneratedDoc     string
	Approved         bool
}

type QuestionTask struct {
	Question string
	Options  []string
	IsFree   bool
}

type AppModel struct {
	App            *tview.Application
	Pages          *tview.Pages
	Sidebar        *tview.List
	ChatView       *tview.TextView
	InputArea      *tview.TextArea
	StatusView     *tview.TextView
	OptionList     *tview.List
	MainFlex       *tview.Flex
	
	CurrentStep    int
	Steps          []Step
	StepState      StepState
	Strategy       QuestionStrategy
	
	APIClient      *api.Client
	IsLoading      bool

	// Question Queue
	pendingTasks   []QuestionTask
	taskIndex      int
	answers        []string
}

func NewStep(name, filename string, number int, sysPrompt, genPrompt string) Step {
	return Step{
		Name:             name,
		Filename:         filename,
		Number:           number,
		SystemPrompt:     sysPrompt,
		GenerationPrompt: genPrompt,
		ChatHistory:      []api.Message{},
	}
}

func main() {
	// Setup logging
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}
	log.Println("--- TUI 2.0 STARTING ---")

	steps := []Step{
		NewStep("One Pager", "ONE_PAGER.md", 1, "", ""),
		NewStep("Developer Spec", "DEV_SPEC.md", 2, "", ""),
		NewStep("Prompt Plan", "PROMPT_PLAN.md", 3, "", ""),
		NewStep("AGENTS", "AGENTS.md", 4, "", ""),
	}

	app := tview.NewApplication()
	
	model := &AppModel{
		App:         app,
		Steps:       steps,
		CurrentStep: 0,
		StepState:   StrategyState,
		APIClient:   api.NewClient("http://localhost:3000"),
	}

	// 0. Strategy Selector
	strategyList := tview.NewList().
		AddItem("Hybrid (Recommended)", "Use Multiple Choice where possible, Free-form otherwise.", '1', nil).
		AddItem("Prefer Multiple Choice", "Bias AI to ask questions with clear options.", '2', nil).
		AddItem("Free-form Only", "Bias AI to have a natural open-ended conversation.", '3', nil)
	strategyList.SetBorder(true).SetTitle(" SELECT INTERVIEW STRATEGY ").SetTitleAlign(tview.AlignCenter)
	
	strategyList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		model.Strategy = QuestionStrategy(index)
		log.Printf("[UI] Strategy selected: %d", model.Strategy)
		model.initializePrompts()
		model.Pages.SwitchToPage("main")
		model.StepState = ChatState
		model.App.SetFocus(model.InputArea)
	})

	// 1. Sidebar
	sidebar := tview.NewList().
		SetSelectedTextColor(tcell.ColorDeepSkyBlue).
		SetSelectedBackgroundColor(tcell.ColorDarkSlateGray)
	sidebar.SetBorder(true).SetTitle(" PROGRESS ")
	sidebar.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if model.Steps[index].Approved {
			model.saveDocument(index)
		}
	})
	model.Sidebar = sidebar

	// 2. Chat View
	chatView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetScrollable(true)
	chatView.SetBorder(true).SetTitle(" CHAT ")
	model.ChatView = chatView

	// 3. Input Area
	inputArea := tview.NewTextArea().
		SetPlaceholder("Describe your idea...")
	inputArea.SetBorder(true).SetTitle(" INPUT ")
	model.InputArea = inputArea

	// 4. Status View
	statusView := tview.NewTextView().
		SetDynamicColors(true)
	statusView.SetText(" [blue]Tab[-] Focus | [blue]Enter[-] Send | [blue]Ctrl+G[-] Generate | [blue]Ctrl+C[-] Quit")
	model.StatusView = statusView

	// 5. Option List (for multiple choice)
	optionList := tview.NewList().
		ShowSecondaryText(false)
	optionList.SetBorder(true).SetTitle(" SELECT OPTION ")
	model.OptionList = optionList

	// Layouts
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex.AddItem(chatView, 0, 1, false)
	rightFlex.AddItem(inputArea, 5, 1, true)
	rightFlex.AddItem(statusView, 1, 1, false)
	
	mainFlex := tview.NewFlex().
		AddItem(sidebar, 25, 1, false).
		AddItem(rightFlex, 0, 4, true)

	// Pages
	model.Pages = tview.NewPages().
		AddPage("strategy", tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(strategyList, 10, 1, true).
				AddItem(nil, 0, 1, false), 60, 1, true).
			AddItem(nil, 0, 1, false), true, true).
		AddPage("main", mainFlex, true, false)

	// Input handling for TextArea
	inputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if model.StepState != ChatState {
			return event
		}
		if event.Key() == tcell.KeyEnter && event.Modifiers() != tcell.ModShift {
			text := inputArea.GetText()
			if strings.TrimSpace(text) != "" {
				log.Printf("[UI] User hit Enter. Sending: %q", text)
				model.handleUserInput(text)
				inputArea.SetText("", false)
			}
			return nil
		}
		return event
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			return nil
		}
		if model.StepState == StrategyState {
			return event
		}
		if event.Key() == tcell.KeyTab {
			if app.GetFocus() == inputArea {
				app.SetFocus(sidebar)
			} else {
				app.SetFocus(inputArea)
			}
			return nil
		}
		if event.Key() == tcell.KeyCtrlG {
			model.generateDoc()
			return nil
		}
		if event.Key() == tcell.KeyCtrlZ {
			model.downloadAll()
			return nil
		}
		return event
	})

	model.syncSidebar()
	model.syncChat()

	log.Println("[UI] Entering main loop...")
	if err := app.SetRoot(model.Pages, true).SetFocus(strategyList).EnableMouse(true).Run(); err != nil {
		log.Fatalf("[FATAL] App crashed: %v", err)
	}
}

func (m *AppModel) initializePrompts() {
	bias := ""
	switch m.Strategy {
	case StrategyMultipleChoice:
		bias = "\n* CRITICAL: Phrase your questions as clear, distinct choices that can be easily turned into a multiple-choice list."
	case StrategyFreeForm:
		bias = "\n* CRITICAL: Phrase your questions as open-ended conversation. Do not use numbered lists or highly structured options."
	case StrategyHybrid:
		bias = "\n* Phrase your questions naturally. Use numbered lists for complex options, and open-ended text for creative ones."
	}

	m.Steps[0].SystemPrompt = `Ask me questions so that we can develop a product specification document for this idea.

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
* For each question, provide your recommendation and a brief explanation.` + bias + `

We are building an MVP - bias your choices towards simplicity, ease of implementation, and speed.`
	
	m.Steps[0].GenerationPrompt = "Now that we've wrapped up the brainstorming process, can you compile our findings into a clean, comprehensive one-pager? Include the problem, audience, ideal customer, platform, and flow information, such that we could start talking with product & engineering leadership about how this could be built."

	m.Steps[1].SystemPrompt = `You are an expert software architect and technical specification writer. You will receive a product specification document as input. Parse it thoroughly before asking clarifying questions. Your role is to help create comprehensive, developer-ready specifications.` + bias + `

Bias your choices towards simplicity, ease of implementation, and speed.`
	m.Steps[1].GenerationPrompt = "Now that we’ve wrapped up the brainstorming process, can you compile our findings into a comprehensive, developer-ready specification? Include all relevant requirements, architecture choices, data handling details, error handling strategies, and a testing plan so a developer can immediately begin implementation."

	m.Steps[2].SystemPrompt = `On this step, we're going to generate a step-by-step prompt plan. The user will now optionally provide some new details about their product. Note them, provide feedback, and wait for them to move on to the prompt plan stage.` + bias
	m.Steps[2].GenerationPrompt = `Draft a detailed, step-by-step blueprint for building this project. Break it down into small, iterative chunks that build on each other. Provide a series of prompts for a code-generation LLM that will implement each step in a test-driven manner. Include, after each prompt, a set of todo checkboxes that the AI agents can check off.`

	m.Steps[3].SystemPrompt = `On this step, we're going to generate an AGENTS.md file. The user will now optionally provide some new details about what they want to be in that file. Note them, provide feedback, and wait for them to move on to the generate stage.` + bias
	m.Steps[3].GenerationPrompt = `Using the information in the prompt_plan & spec attached here, write a minimal AGENTS.md file to include in the repository so that agents like Codex & Claude Code can interact well. Include Repository docs, Agent responsibility, Guardrails for agents, and Testing policy.`
}

func (m *AppModel) syncSidebar() {
	m.Sidebar.Clear()
	for i, step := range m.Steps {
		prefix := "  "
		suffix := ""
		if i == m.CurrentStep {
			prefix = "[blue]● [-]"
		} else if step.Approved {
			prefix = "[green]✓ [-]"
			suffix = " [yellow]⬇[-]"
		}
		m.Sidebar.AddItem(prefix+step.Name+suffix, "", 0, nil)
	}
	m.Sidebar.SetCurrentItem(m.CurrentStep)
}

func (m *AppModel) syncChat() {
	m.ChatView.Clear()
	step := m.Steps[m.CurrentStep]
	for _, msg := range step.ChatHistory {
		roleLabel := "[yellow]YOU[-]"
		if msg.Role == "assistant" {
			roleLabel = "[cyan]AI[-]"
		}
		// Escape content to prevent tview tag parsing
		fmt.Fprintf(m.ChatView, "%s\n%s\n\n", roleLabel, tview.Escape(msg.Content))
	}

	if !m.IsLoading && len(step.ChatHistory) > 0 {
		lastMsg := step.ChatHistory[len(step.ChatHistory)-1]
		if lastMsg.Role == "assistant" && !strings.Contains(lastMsg.Content, "?") {
			fmt.Fprintf(m.ChatView, "[green:black:b] REQUIREMENTS COMPLETE [-]\n[gray]Press [white:blue] Ctrl+G [-][gray] to generate the %s.[-]\n\n", step.Name)
			m.StatusView.SetText(" [blue]Tab[-] Focus | [blue]Enter[-] Send | [yellow:black:b]Ctrl+G[-] Generate | [blue]Ctrl+C[-] Quit")
		} else {
			m.StatusView.SetText(" [blue]Tab[-] Focus | [blue]Enter[-] Send | [blue]Ctrl+G[-] Generate | [blue]Ctrl+C[-] Quit")
		}
	}

	if m.IsLoading {
		fmt.Fprintf(m.ChatView, "[blue]AI is thinking...[-]\n")
	}
	m.ChatView.ScrollToEnd()
}

func (m *AppModel) handleUserInput(text string) {
	log.Printf("[CORE] handleUserInput: %s", text)
	step := &m.Steps[m.CurrentStep]
	step.ChatHistory = append(step.ChatHistory, api.Message{
		Role:    "user",
		Content: text,
	})
	
	m.IsLoading = true
	m.StatusView.SetText(" [yellow]Sending request to AI...[-]")
	m.syncChat()
	
	go func() {
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

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		var lastUpdate time.Time
		receivingStarted := false

		err := m.APIClient.StreamChat(req, func(chunk string) {
			if !receivingStarted {
				receivingStarted = true
				m.App.QueueUpdateDraw(func() {
					m.IsLoading = false
					m.StatusView.SetText(" [green]Receiving AI response...[-]")
				})
			}
			history := m.Steps[m.CurrentStep].ChatHistory
			if len(history) > 0 && history[len(history)-1].Role == "assistant" {
				m.Steps[m.CurrentStep].ChatHistory[len(history)-1].Content += chunk
			} else {
				m.Steps[m.CurrentStep].ChatHistory = append(m.Steps[m.CurrentStep].ChatHistory, api.Message{
					Role:    "assistant",
					Content: chunk,
				})
			}
			if time.Since(lastUpdate) > 50*time.Millisecond {
				m.App.QueueUpdateDraw(func() { m.syncChat() })
				lastUpdate = time.Now()
			}
		})

		m.App.QueueUpdateDraw(func() {
			m.IsLoading = false
			if err != nil {
				m.StatusView.SetText(fmt.Sprintf("[red]ERROR: %v[-]", err))
			} else {
				m.StatusView.SetText(" [blue]Tab[-] Focus | [blue]Enter[-] Send | [blue]Ctrl+G[-] Generate | [blue]Ctrl+C[-] Quit")
			}
			m.syncChat()
			m.checkForQuestions()
		})
	}()
}

func (m *AppModel) checkForQuestions() {
	log.Println("[CORE] checkForQuestions starting...")
	step := m.Steps[m.CurrentStep]
	if len(step.ChatHistory) == 0 {
		return
	}
	lastMsg := step.ChatHistory[len(step.ChatHistory)-1]
	if lastMsg.Role != "assistant" {
		return
	}

	lines := strings.Split(lastMsg.Content, "\n")
	var questions []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// More robust cleaning: strip markdown first, then check for ?
		cleaned := strings.ReplaceAll(trimmed, "**", "")
		cleaned = strings.ReplaceAll(cleaned, "__", "")
		cleaned = strings.TrimSpace(cleaned)
		
		if strings.Contains(cleaned, "?") {
			// Extract up to the question mark to avoid trailing noise
			idx := strings.Index(cleaned, "?")
			question := cleaned[:idx+1]
			if len(question) > 5 { // Avoid tiny fragments
				questions = append(questions, question)
				log.Printf("[CORE] Detected question: %q", question)
			}
		}
	}

	if len(questions) > 0 {
		m.pendingTasks = []QuestionTask{}
		m.taskIndex = 0
		m.answers = make([]string, len(questions))
		
		log.Printf("[CORE] Processing %d questions...", len(questions))
		m.IsLoading = true
		m.syncChat()

		go func() {
			for i, q := range questions {
				m.App.QueueUpdateDraw(func() {
					m.StatusView.SetText(fmt.Sprintf(" [yellow]Analyzing Question %d/%d...[-]", i+1, len(questions)))
				})
				
				res, err := m.APIClient.GenerateOptions(api.GenerateOptionsRequest{
					QuestionText: q,
					ChatHistory:  step.ChatHistory,
				})
				
				task := QuestionTask{Question: q}
				if err == nil && res != nil && len(res.Options) > 0 {
					log.Printf("[API] Options found for %q", q)
					task.Options = res.Options
				} else {
					log.Printf("[API] No options for %q (Error: %v)", q, err)
					task.IsFree = true
				}
				m.pendingTasks = append(m.pendingTasks, task)
			}

			m.App.QueueUpdateDraw(func() {
				m.IsLoading = false
				m.syncChat()
				m.showNextQuestion()
			})
		}()
	} else {
		log.Println("[CORE] No questions detected in the last response.")
	}
}

func (m *AppModel) showNextQuestion() {
	if m.taskIndex >= len(m.pendingTasks) {
		var combinedAnswer strings.Builder
		for i, q := range m.pendingTasks {
			combinedAnswer.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, q.Question, m.answers[i]))
		}
		m.handleUserInput(combinedAnswer.String())
		return
	}

	task := m.pendingTasks[m.taskIndex]
	m.StepState = AnsweringState

	if task.IsFree {
		m.showFreeformInput(task)
	} else {
		m.showMultipleChoice(task)
	}
}

func (m *AppModel) showMultipleChoice(task QuestionTask) {
	m.OptionList.Clear()
	for _, opt := range task.Options {
		m.OptionList.AddItem(opt, "", 0, nil)
	}
	
	m.OptionList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		m.answers[m.taskIndex] = mainText
		m.Pages.RemovePage("question")
		m.taskIndex++
		m.showNextQuestion()
	})

	// Add Esc handler
	m.OptionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.Pages.RemovePage("question")
			m.StepState = ChatState
			m.App.SetFocus(m.InputArea)
			m.syncChat()
			return nil
		}
		return event
	})

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetText(fmt.Sprintf("[white:blue:b] QUESTION %d/%d [-]\n\n%s", m.taskIndex+1, len(m.pendingTasks), tview.Escape(task.Question)))
	header.SetBorder(true).SetBorderColor(tcell.ColorDeepSkyBlue)

	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(header, 10, 1, false).
				AddItem(m.OptionList, 10, 1, true).
				AddItem(tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).SetText("[gray]Press Enter to Select | Esc to Cancel[-]"), 1, 1, false), 80, 1, true).
			AddItem(nil, 0, 1, false), 22, 1, true).
		AddItem(nil, 0, 1, false)

	m.Pages.AddAndSwitchToPage("question", modal, true)
	m.App.SetFocus(m.OptionList)
}

func (m *AppModel) showFreeformInput(task QuestionTask) {
	answerInput := tview.NewInputField()
	answerInput.SetLabel(" Answer: ").
		SetLabelColor(tcell.ColorDeepSkyBlue).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetBorder(true)
	
	answerInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := answerInput.GetText()
			if strings.TrimSpace(text) == "" {
				return
			}
			m.answers[m.taskIndex] = text
			m.Pages.RemovePage("question")
			m.taskIndex++
			m.showNextQuestion()
		}
	})

	// Add Esc handler
	answerInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.Pages.RemovePage("question")
			m.StepState = ChatState
			m.App.SetFocus(m.InputArea)
			m.syncChat()
			return nil
		}
		return event
	})

	header := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetTextAlign(tview.AlignLeft).
		SetText(fmt.Sprintf("[white:blue:b] QUESTION %d/%d [-]\n\n%s", m.taskIndex+1, len(m.pendingTasks), tview.Escape(task.Question)))
	header.SetBorder(true).SetBorderColor(tcell.ColorDeepSkyBlue)

	modal := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(header, 10, 1, false).
				AddItem(answerInput, 3, 1, true).
				AddItem(tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).SetText("[gray]Type answer and press Enter | Esc to Cancel[-]"), 1, 1, false), 80, 1, true).
			AddItem(nil, 0, 1, false), 16, 1, true).
		AddItem(nil, 0, 1, false)

	m.Pages.AddAndSwitchToPage("question", modal, true)
	m.App.SetFocus(answerInput)
}

func (m *AppModel) generateDoc() {
	log.Println("[CORE] generateDoc requested")
	m.IsLoading = true
	m.StepState = GeneratingState
	m.syncChat()
	m.StatusView.SetText("[yellow]Generating document...[-]")

	go func() {
		step := &m.Steps[m.CurrentStep]
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

		step.GeneratedDoc = ""
		err := m.APIClient.StreamGenerateDoc(req, func(chunk string) {
			m.App.QueueUpdateDraw(func() {
				step.GeneratedDoc += chunk
				m.StatusView.SetText(fmt.Sprintf("[yellow]Generating %s... (%d bytes)[-]", step.Name, len(step.GeneratedDoc)))
			})
		})

		m.App.QueueUpdateDraw(func() {
			m.IsLoading = false
			if err != nil {
				m.StatusView.SetText(fmt.Sprintf("[red]ERROR: %v[-]", err))
				m.StepState = ChatState
			} else {
				m.showPreview()
			}
		})
	}()
}

func (m *AppModel) showPreview() {
	m.StepState = PreviewState
	
	previewView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)
	previewView.SetBorder(true).SetTitle(fmt.Sprintf(" PREVIEW: %s ", m.Steps[m.CurrentStep].Name))
	previewView.SetBorderColor(tcell.ColorGreen)
	
	fmt.Fprintf(previewView, "%s", tview.Escape(m.Steps[m.CurrentStep].GeneratedDoc))

	previewFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(previewView, 0, 1, true).
		AddItem(tview.NewTextView().SetDynamicColors(true).SetText(" [blue]Ctrl+A[-] Approve | [blue]Esc[-] Back to Chat"), 1, 1, false)

	m.Pages.AddPage("preview", previewFlex, true, true)
	m.App.SetFocus(previewView)

	previewView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			log.Println("[UI] Preview: Esc caught")
			m.Pages.RemovePage("preview")
			m.StepState = ChatState
			m.App.SetFocus(m.InputArea)
			m.syncChat()
			return nil
		}
		if event.Key() == tcell.KeyCtrlA {
			m.approveStep()
			return nil
		}
		return event
	})
}

func (m *AppModel) approveStep() {
	m.Steps[m.CurrentStep].Approved = true
	m.Pages.RemovePage("preview")
	m.saveDocument(m.CurrentStep)
	
	if m.CurrentStep < len(m.Steps)-1 {
		m.CurrentStep++
		m.StepState = ChatState
		m.syncSidebar()
		m.syncChat()
		m.App.SetFocus(m.InputArea)
		m.StatusView.SetText(fmt.Sprintf("[green]Step approved! Moved to %s[-]", m.Steps[m.CurrentStep].Name))
	} else {
		m.syncSidebar()
		m.syncChat()
		m.App.SetFocus(m.Sidebar)
		m.StatusView.SetText("[green]All steps complete! Press Ctrl+Z to Download All ZIP.[-]")
	}
}

func (m *AppModel) saveDocument(index int) {
	step := m.Steps[index]
	err := os.WriteFile(step.Filename, []byte(step.GeneratedDoc), 0644)
	if err != nil {
		m.StatusView.SetText(fmt.Sprintf("[red]Failed to save %s[-]", step.Filename))
	} else {
		m.StatusView.SetText(fmt.Sprintf("[green]Saved %s successfully![-]", step.Filename))
	}
}

func (m *AppModel) downloadAll() {
	zipName := "vibe-scaffold-docs.zip"
	newZipFile, err := os.Create(zipName)
	if err != nil {
		m.StatusView.SetText(fmt.Sprintf("[red]Failed to create ZIP: %v[-]", err))
		return
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, step := range m.Steps {
		if step.Approved && step.GeneratedDoc != "" {
			f, err := zipWriter.Create(step.Filename)
			if err != nil {
				continue
			}
			_, _ = f.Write([]byte(step.GeneratedDoc))
		}
	}

	m.StatusView.SetText(fmt.Sprintf("[green:black:b] All documents bundled into %s! [-]", zipName))
}
