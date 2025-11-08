package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel shows progress for long-running operations
type ProgressModel struct {
	spinner spinner.Model
	message string
	done    bool
	err     error
	width   int
	height  int
}

// NewProgressModel creates a new progress indicator model
func NewProgressModel(message string) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return ProgressModel{
		spinner: s,
		message: message,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ProgressCompleteMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	}

	return m, nil
}

func (m ProgressModel) View() string {
	if m.done {
		if m.err != nil {
			return errorStyle.Render(fmt.Sprintf("✗ %s: %v", m.message, m.err))
		}
		return successStyle.Render(fmt.Sprintf("✓ %s", m.message))
	}

	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

// ProgressCompleteMsg signals operation completion
type ProgressCompleteMsg struct {
	err error
}

// BatchProgressModel shows progress for multiple items
type BatchProgressModel struct {
	items       []string
	current     int
	total       int
	currentItem string
	done        bool
	errors      []error
	spinner     spinner.Model
	width       int
	height      int
}

// NewBatchProgressModel creates a progress model for batch operations
func NewBatchProgressModel(items []string) BatchProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return BatchProgressModel{
		items:   items,
		total:   len(items),
		spinner: s,
	}
}

func (m BatchProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m BatchProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if !m.done {
				return m, tea.Quit
			}
		case "enter", " ", "q":
			if m.done {
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		if !m.done {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case BatchProgressMsg:
		m.current = msg.current
		m.currentItem = msg.item
		if msg.err != nil {
			m.errors = append(m.errors, msg.err)
		}
		if m.current >= m.total {
			m.done = true
			return m, nil
		}

	case BatchCompleteMsg:
		m.done = true
		return m, nil
	}

	return m, nil
}

func (m BatchProgressModel) View() string {
	var s strings.Builder

	if m.done {
		// Summary
		s.WriteString(successStyle.Render("✓ Installation complete\n\n"))
		s.WriteString(fmt.Sprintf("Processed: %d/%d\n", m.total, m.total))

		if len(m.errors) > 0 {
			s.WriteString(fmt.Sprintf("Errors: %d\n\n", len(m.errors)))
			s.WriteString(errorStyle.Render("Failed items:\n"))
			for _, err := range m.errors {
				s.WriteString(fmt.Sprintf("  • %v\n", err))
			}
		} else {
			s.WriteString(successStyle.Render("\nAll items installed successfully!"))
		}

		s.WriteString("\n\n" + helpStyle.Render("Press enter to continue"))
	} else {
		// Progress
		s.WriteString(titleStyle.Render("Installing Packages\n\n"))

		// Progress bar
		percent := float64(m.current) / float64(m.total)
		barWidth := 40
		filled := int(percent * float64(barWidth))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		s.WriteString(fmt.Sprintf("[%s] %d/%d\n\n", bar, m.current, m.total))

		// Current item
		if m.currentItem != "" {
			s.WriteString(fmt.Sprintf("%s Installing: %s\n", m.spinner.View(), m.currentItem))
		}

		s.WriteString("\n" + helpStyle.Render("Press Ctrl+C to cancel"))
	}

	return boxStyle.Render(s.String())
}

// BatchProgressMsg updates batch progress
type BatchProgressMsg struct {
	current int
	item    string
	err     error
}

// BatchCompleteMsg signals batch completion
type BatchCompleteMsg struct{}
