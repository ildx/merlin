package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ildx/merlin/internal/models"
	"github.com/ildx/merlin/internal/scripts"
)

// ScriptStatus represents the execution status of a script
type ScriptStatus int

const (
	StatusPending ScriptStatus = iota
	StatusRunning
	StatusSuccess
	StatusFailed
)

// ScriptExecution tracks a single script's execution
type ScriptExecution struct {
	Script   models.ScriptItem
	Status   ScriptStatus
	Duration time.Duration
	Error    error
	Output   string
}

// ScriptRunnerModel handles batch script execution with progress display
type ScriptRunnerModel struct {
	toolName   string
	toolRoot   string
	scriptDir  string
	scripts    []models.ScriptItem
	executions []ScriptExecution
	current    int
	runner     *scripts.ScriptRunner
	done       bool
	err        error
	width      int
	height     int
}

// scriptFinishedMsg is sent when a script completes
type scriptFinishedMsg struct {
	index    int
	result   *scripts.ScriptResult
	duration time.Duration
}

// allScriptsFinishedMsg is sent when all scripts complete
type allScriptsFinishedMsg struct{}

// NewScriptRunnerModel creates a new script runner model
func NewScriptRunnerModel(toolName, toolRoot, scriptDir string, scriptsToRun []models.ScriptItem, runner *scripts.ScriptRunner) ScriptRunnerModel {
	executions := make([]ScriptExecution, len(scriptsToRun))
	for i, script := range scriptsToRun {
		executions[i] = ScriptExecution{
			Script: script,
			Status: StatusPending,
		}
	}

	return ScriptRunnerModel{
		toolName:   toolName,
		toolRoot:   toolRoot,
		scriptDir:  scriptDir,
		scripts:    scriptsToRun,
		executions: executions,
		runner:     runner,
	}
}

func (m ScriptRunnerModel) Init() tea.Cmd {
	// Start running scripts immediately
	return m.runNextScript()
}

func (m ScriptRunnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "q", "esc", "enter", " ":
				return m, tea.Quit
			}
		}
		// Don't allow quitting during execution
		return m, nil

	case scriptFinishedMsg:
		// Update execution status
		if msg.index < len(m.executions) {
			exec := &m.executions[msg.index]
			exec.Duration = msg.duration
			if msg.result.Success {
				exec.Status = StatusSuccess
			} else {
				exec.Status = StatusFailed
				exec.Error = msg.result.Error
				exec.Output = msg.result.Output
			}
		}

		// Move to next script
		m.current++
		if m.current < len(m.scripts) {
			return m, m.runNextScript()
		}

		// All done
		m.done = true
		return m, tea.Batch(
			tea.Printf(""), // Trigger a render
			func() tea.Msg { return allScriptsFinishedMsg{} },
		)

	case allScriptsFinishedMsg:
		return m, nil
	}

	return m, nil
}

func (m ScriptRunnerModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(fmt.Sprintf("📜 Running Scripts: %s", m.toolName)) + "\n\n")

	// Show execution progress for each script
	for _, exec := range m.executions {
		var icon, status string
		var style lipgloss.Style

		switch exec.Status {
		case StatusPending:
			icon = "⏳"
			status = "Pending"
			style = dimStyle
		case StatusRunning:
			icon = "▶"
			status = "Running..."
			style = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
		case StatusSuccess:
			icon = "✓"
			duration := fmt.Sprintf("(%.2fs)", exec.Duration.Seconds())
			status = lipgloss.NewStyle().Foreground(successColor).Render(duration)
			style = successStyle
		case StatusFailed:
			icon = "✗"
			status = lipgloss.NewStyle().Foreground(errorColor).Render("Failed")
			style = errorStyle
		}

		line := fmt.Sprintf("%s %s", icon, exec.Script.File)
		s.WriteString(style.Render(line))

		if exec.Status != StatusPending {
			s.WriteString(" " + status)
		}
		s.WriteString("\n")

		// Show error details for failed scripts
		if exec.Status == StatusFailed && exec.Error != nil {
			errMsg := lipgloss.NewStyle().
				Foreground(errorColor).
				PaddingLeft(4).
				Render(fmt.Sprintf("└─ %v", exec.Error))
			s.WriteString(errMsg + "\n")
		}
	}

	// Summary at the bottom
	if m.done {
		s.WriteString("\n")
		successCount := 0
		failCount := 0
		for _, exec := range m.executions {
			if exec.Status == StatusSuccess {
				successCount++
			} else if exec.Status == StatusFailed {
				failCount++
			}
		}

		if failCount == 0 {
			s.WriteString(successStyle.Render(fmt.Sprintf("✓ All %d scripts completed successfully!", successCount)) + "\n")
		} else {
			s.WriteString(warningStyle.Render(fmt.Sprintf("⚠ %d succeeded, %d failed", successCount, failCount)) + "\n")
		}

		s.WriteString(helpStyle.Render("\nPress any key to continue..."))
	} else {
		progress := fmt.Sprintf("Progress: %d/%d", m.current, len(m.scripts))
		s.WriteString("\n" + dimStyle.Render(progress))
	}

	return boxStyle.Render(s.String())
}

// runNextScript executes the next script in the queue
func (m *ScriptRunnerModel) runNextScript() tea.Cmd {
	if m.current >= len(m.scripts) {
		return nil
	}

	// Mark current as running
	m.executions[m.current].Status = StatusRunning

	index := m.current
	script := m.scripts[index]

	return func() tea.Msg {
		start := time.Now()
		scriptPath := fmt.Sprintf("%s/%s", m.scriptDir, script.File)
		result := m.runner.RunScript(scriptPath)
		duration := time.Since(start)

		return scriptFinishedMsg{
			index:    index,
			result:   result,
			duration: duration,
		}
	}
}

// HasFailures returns true if any scripts failed
func (m ScriptRunnerModel) HasFailures() bool {
	for _, exec := range m.executions {
		if exec.Status == StatusFailed {
			return true
		}
	}
	return false
}

// GetFailedScripts returns a list of failed script executions
func (m ScriptRunnerModel) GetFailedScripts() []ScriptExecution {
	var failed []ScriptExecution
	for _, exec := range m.executions {
		if exec.Status == StatusFailed {
			failed = append(failed, exec)
		}
	}
	return failed
}
