package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"vibe-scaffold-tui2/api"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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
	
	APIClient      *api.Client
	IsLoading      bool
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

func main() {
	// Setup logging
	f, _ := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if f != nil {
		defer f.Close()
		log.SetOutput(f)
	}
	log.Println("Starting TUI 2.0...")

	steps := []Step{
		NewStep("One Pager", 1, "", ""),
		NewStep("Developer Spec", 2, "", ""),
		NewStep("Prompt Plan", 3, "", ""),
		NewStep("AGENTS", 4, "", ""),
	}

	// Initialize Prompts
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

	app := tview.NewApplication()
	
	model := &AppModel{
		App:         app,
		Steps:       steps,
		CurrentStep: 0,
		StepState:   ChatState,
		APIClient:   api.NewClient("http://localhost:3000"),
	}

	// 1. Sidebar
	sidebar := tview.NewList()
	sidebar.SetBorder(true).SetTitle(" PROGRESS ").SetTitleAlign(tview.AlignLeft)
	sidebar.SetBorderColor(tcell.ColorSlateGray)
	sidebar.SetMainTextColor(tcell.ColorGray)
	sidebar.SetSelectedTextColor(tcell.ColorDeepSkyBlue)
	sidebar.SetSelectedBackgroundColor(tcell.ColorDarkSlateGray)
	
	for i, step := range steps {
		sidebar.AddItem(fmt.Sprintf("%d. %s", i+1, step.Name), "", 0, nil)
	}
	sidebar.SetCurrentItem(0)
	model.Sidebar = sidebar

	// 2. Chat View
	chatView := tview.NewTextView()
	chatView.SetDynamicColors(true).SetRegions(true).SetWordWrap(true).SetScrollable(true)
	chatView.SetBorder(true).SetTitle(" CHAT ").SetTitleAlign(tview.AlignLeft)
	chatView.SetBorderColor(tcell.ColorSlateGray)
	model.ChatView = chatView

	// 3. Input Area (TextArea for wordwrap)
	inputArea := tview.NewTextArea()
	inputArea.SetPlaceholder("Describe your idea...")
	inputArea.SetBorder(true).SetTitle(" INPUT ").SetTitleAlign(tview.AlignLeft)
	inputArea.SetBorderColor(tcell.ColorSlateGray)
	model.InputArea = inputArea

	// 4. Status View
	statusView := tview.NewTextView()
	statusView.SetDynamicColors(true).SetTextAlign(tview.AlignLeft)
	statusView.SetTextColor(tcell.ColorGray)
	model.StatusView = statusView
	statusView.SetText(" [blue]Tab[-] Focus | [blue]Enter[-] Send | [blue]Ctrl+G[-] Generate | [blue]Ctrl+C[-] Quit")

	// 5. Option List (for multiple choice)
	optionList := tview.NewList()
	optionList.SetBorder(true).SetTitle(" SELECT OPTION ").SetTitleAlign(tview.AlignLeft)
	optionList.SetBorderColor(tcell.ColorDeepSkyBlue)
	optionList.ShowSecondaryText(false)
	model.OptionList = optionList

	// Layout
	rightFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	rightFlex.AddItem(chatView, 0, 1, false)
	rightFlex.AddItem(inputArea, 5, 1, true)
	rightFlex.AddItem(statusView, 1, 1, false)
	model.MainFlex = rightFlex

	mainFlex := tview.NewFlex().
		AddItem(sidebar, 25, 1, false).
		AddItem(rightFlex, 0, 4, true)

	model.Pages = tview.NewPages().
		AddPage("main", mainFlex, true, true)

	// Input handling for TextArea
	inputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter && event.Modifiers() != tcell.ModShift {
			text := inputArea.GetText()
			if strings.TrimSpace(text) != "" {
				log.Printf("Sending input: %s", text)
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
		return event
	})

	model.updateSidebar()
	model.updateChat()

	if err := app.SetRoot(model.Pages, true).Run(); err != nil {
		panic(err)
	}
}

func (m *AppModel) updateSidebar() {
	m.App.QueueUpdateDraw(func() {
		m.Sidebar.Clear()
		for i, step := range m.Steps {
			prefix := "  "
			if i == m.CurrentStep {
				prefix = "[blue]● [-]"
			} else if step.Approved {
				prefix = "[green]✓ [-]"
			}
			m.Sidebar.AddItem(prefix+step.Name, "", 0, nil)
		}
		m.Sidebar.SetCurrentItem(m.CurrentStep)
	})
}

func (m *AppModel) updateChat() {
	m.App.QueueUpdateDraw(func() {
		m.ChatView.Clear()
		step := m.Steps[m.CurrentStep]
		for _, msg := range step.ChatHistory {
			roleLabel := "[yellow]YOU[-]"
			if msg.Role == "assistant" {
				roleLabel = "[cyan]AI[-]"
			}
			fmt.Fprintf(m.ChatView, "%s\n%s\n\n", roleLabel, msg.Content)
		}
		if m.IsLoading {
			fmt.Fprintf(m.ChatView, "[blue]AI is thinking...[-]\n")
		}
		m.ChatView.ScrollToEnd()
	})
}

func (m *AppModel) handleUserInput(text string) {
	log.Printf("handleUserInput called with: %s", text)
	step := &m.Steps[m.CurrentStep]
	step.ChatHistory = append(step.ChatHistory, api.Message{
		Role:    "user",
		Content: text,
	})
	
	m.IsLoading = true
	m.updateChat()
	
	go func() {
		log.Println("Starting API call...")
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
			m.App.QueueUpdateDraw(func() {
				history := m.Steps[m.CurrentStep].ChatHistory
				if len(history) > 0 && history[len(history)-1].Role == "assistant" {
					m.Steps[m.CurrentStep].ChatHistory[len(history)-1].Content += chunk
				} else {
					m.Steps[m.CurrentStep].ChatHistory = append(m.Steps[m.CurrentStep].ChatHistory, api.Message{
						Role:    "assistant",
						Content: chunk,
					})
				}
				m.updateChat()
			})
		})

		m.App.QueueUpdateDraw(func() {
			m.IsLoading = false
			if err != nil {
				log.Printf("API Error: %v", err)
				m.StatusView.SetText(fmt.Sprintf("[red]ERROR: %v[-]", err))
			} else {
				log.Println("API call completed successfully")
			}
			m.updateChat()
			m.checkForQuestions()
		})
	}()
}

func (m *AppModel) checkForQuestions() {
	step := m.Steps[m.CurrentStep]
	if len(step.ChatHistory) == 0 {
		return
	}
	lastMsg := step.ChatHistory[len(step.ChatHistory)-1]
	if lastMsg.Role != "assistant" {
		return
	}

	lines := strings.Split(lastMsg.Content, "\n")
	var question string
	lastQIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.Contains(lines[i], "?") {
			lastQIdx = i
			break
		}
	}

	if lastQIdx != -1 {
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
		question = strings.ReplaceAll(question, "**", "")
		question = strings.TrimSpace(question)
	}

	if question != "" {
		log.Printf("Question detected: %s", question)
		m.IsLoading = true
		m.updateChat()
		go func() {
			res, err := m.APIClient.GenerateOptions(api.GenerateOptionsRequest{
				QuestionText: question,
				ChatHistory:  step.ChatHistory,
			})
			
			m.App.QueueUpdateDraw(func() {
				m.IsLoading = false
				if err == nil && res != nil && len(res.Options) > 0 {
					log.Printf("Options found: %d", len(res.Options))
					m.showOptions(res)
				} else {
					log.Printf("No options or error: %v", err)
				}
				m.updateChat()
			})
		}()
	}
}

func (m *AppModel) showOptions(res *api.GenerateOptionsResponse) {
	m.OptionList.Clear()
	m.OptionList.SetTitle(fmt.Sprintf(" %s ", res.Question))
	for _, opt := range res.Options {
		m.OptionList.AddItem(opt, "", 0, nil)
	}
	
	m.OptionList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		m.Pages.RemovePage("options")
		m.handleUserInput(mainText)
		m.App.SetFocus(m.InputArea)
	})

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(m.OptionList, 10, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)

	m.Pages.AddPage("options", modal, true, true)
	m.App.SetFocus(m.OptionList)
}

func (m *AppModel) generateDoc() {
	m.IsLoading = true
	m.StepState = GeneratingState
	m.updateChat()
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
	
	previewView := tview.NewTextView()
	previewView.SetDynamicColors(true).SetWordWrap(true).SetScrollable(true)
	previewView.SetBorder(true).SetTitle(fmt.Sprintf(" PREVIEW: %s ", m.Steps[m.CurrentStep].Name))
	previewView.SetBorderColor(tcell.ColorGreen)
	
	fmt.Fprintf(previewView, "%s", m.Steps[m.CurrentStep].GeneratedDoc)

	previewFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(previewView, 0, 1, true).
		AddItem(tview.NewTextView().SetDynamicColors(true).SetText(" [blue]Ctrl+A[-] Approve | [blue]Esc[-] Back to Chat"), 1, 1, false)

	m.Pages.AddPage("preview", previewFlex, true, true)
	m.App.SetFocus(previewView)

	previewView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.Pages.RemovePage("preview")
			m.StepState = ChatState
			m.App.SetFocus(m.InputArea)
			m.StatusView.SetText(" [blue]Tab[-] Focus | [blue]Enter[-] Send | [blue]Ctrl+G[-] Generate | [blue]Ctrl+C[-] Quit")
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
	
	if m.CurrentStep < len(m.Steps)-1 {
		m.CurrentStep++
		m.StepState = ChatState
		m.updateSidebar()
		m.updateChat()
		m.App.SetFocus(m.InputArea)
		m.StatusView.SetText(fmt.Sprintf("[green]Step approved! Moved to %s[-]", m.Steps[m.CurrentStep].Name))
	} else {
		m.StatusView.SetText("[green]All steps complete! Press Ctrl+C to exit.[-]")
	}
}

