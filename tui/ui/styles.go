package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Cyan    = lipgloss.Color("#06b6d4")
	Amber   = lipgloss.Color("#f59e0b")
	Purple  = lipgloss.Color("#8b5cf6")
	Green   = lipgloss.Color("#10b981")
	Red     = lipgloss.Color("#ef4444")
	Gray    = lipgloss.Color("#71717a")
	Zinc    = lipgloss.Color("#18181b")
	White   = lipgloss.Color("#fafafa")

	// Base Styles
	MainStyle = lipgloss.NewStyle().
			Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Purple).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	FooterStyle = lipgloss.NewStyle().
			Foreground(Gray).
			MarginTop(1).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(Zinc)

	// Sidebar / Progress Styles
	SidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(Zinc).
			PaddingRight(2).
			MarginRight(2).
			Width(25)

	StepActiveStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true).
			PaddingLeft(2)

	StepInactiveStyle = lipgloss.NewStyle().
			Foreground(Gray).
			PaddingLeft(2)

	StepCompletedStyle = lipgloss.NewStyle().
			Foreground(Green).
			PaddingLeft(2)

	// Chat Styles
	ChatContainerStyle = lipgloss.NewStyle()

	UserLabelStyle = lipgloss.NewStyle().
			Foreground(Amber).
			Bold(true)

	AssistantLabelStyle = lipgloss.NewStyle().
				Foreground(Cyan).
				Bold(true)

	MessageStyle = lipgloss.NewStyle().
			MarginBottom(1)

	// Document Preview Styles
	DocContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Purple).
				Padding(1).
				MarginBottom(1)

	// Feedback Styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Red).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Green).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(Cyan).
			Italic(true)

	// Help / Shortcuts
	ShortcutStyle = lipgloss.NewStyle().
			Foreground(White).
			Background(Zinc).
			Padding(0, 1)

	ShortcutDescStyle = lipgloss.NewStyle().
				Foreground(Gray)
)
