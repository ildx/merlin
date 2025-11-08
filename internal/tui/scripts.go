package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ildx/merlin/internal/models"
)

// ToolScriptItem represents a tool with its scripts for selection
type ToolScriptItem struct {
	ToolName    string
	Description string
	Scripts     []models.ScriptItem
	Selected    bool
}

// ToolScriptSelectorModel allows selecting a tool to run scripts from
type ToolScriptSelectorModel struct {
	title     string
	tools     []ToolScriptItem
	cursor    int
	confirmed bool
	cancelled bool
	width     int
	height    int
}

// NewToolScriptSelectorModel creates a new tool script selector
func NewToolScriptSelectorModel(title string, tools []ToolScriptItem) ToolScriptSelectorModel {
	return ToolScriptSelectorModel{
		title: title,
		tools: tools,
	}
}

func (m ToolScriptSelectorModel) Init() tea.Cmd {
	return nil
}

func (m ToolScriptSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tools)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.tools) > 0 {
				m.confirmed = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m ToolScriptSelectorModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(m.title) + "\n\n")

	if len(m.tools) == 0 {
		s.WriteString(dimStyle.Render("No tools with scripts found.") + "\n")
		s.WriteString(helpStyle.Render("\nesc: back"))
		return boxStyle.Render(s.String())
	}

	// Render tool list
	for i, tool := range m.tools {
		cursor := "  "
		style := normalItemStyle

		if i == m.cursor {
			cursor = "▸ "
			style = selectedItemStyle
		}

		scriptCount := len(tool.Scripts)
		line := fmt.Sprintf("%s%s (%d script%s)", cursor, tool.ToolName, scriptCount, pluralize(scriptCount))
		s.WriteString(style.Render(line) + "\n")

		// Show description for selected item
		if i == m.cursor && tool.Description != "" {
			desc := lipgloss.NewStyle().
				Foreground(mutedColor).
				PaddingLeft(4).
				Italic(true).
				Render(tool.Description)
			s.WriteString(desc + "\n")
		}
	}

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select scripts • esc: cancel"))

	return boxStyle.Render(s.String())
}

// IsConfirmed returns true if user confirmed selection
func (m ToolScriptSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if user cancelled
func (m ToolScriptSelectorModel) IsCancelled() bool {
	return m.cancelled
}

// GetSelectedTool returns the currently selected tool
func (m ToolScriptSelectorModel) GetSelectedTool() *ToolScriptItem {
	if m.cursor >= 0 && m.cursor < len(m.tools) {
		return &m.tools[m.cursor]
	}
	return nil
}

// ScriptSelectorModel allows multi-selecting scripts to run
type ScriptSelectorModel struct {
	title     string
	toolName  string
	scripts   []models.ScriptItem
	selected  map[int]bool
	cursor    int
	confirmed bool
	cancelled bool
	width     int
	height    int
}

// NewScriptSelectorModel creates a new script selector
func NewScriptSelectorModel(title, toolName string, scripts []models.ScriptItem) ScriptSelectorModel {
	// Pre-select all scripts by default
	selected := make(map[int]bool)
	for i := range scripts {
		selected[i] = true
	}

	return ScriptSelectorModel{
		title:    title,
		toolName: toolName,
		scripts:  scripts,
		selected: selected,
	}
}

func (m ScriptSelectorModel) Init() tea.Cmd {
	return nil
}

func (m ScriptSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit

		case "q":
			if len(m.GetSelectedScripts()) == 0 {
				m.cancelled = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.scripts)-1 {
				m.cursor++
			}

		case " ":
			// Toggle selection
			m.selected[m.cursor] = !m.selected[m.cursor]

		case "a":
			// Select all
			for i := range m.scripts {
				m.selected[i] = true
			}

		case "n":
			// Select none
			for i := range m.scripts {
				m.selected[i] = false
			}

		case "enter":
			if len(m.GetSelectedScripts()) > 0 {
				m.confirmed = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m ScriptSelectorModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render(m.title) + "\n")
	s.WriteString(subtitleStyle.Render(fmt.Sprintf("Tool: %s", m.toolName)) + "\n\n")

	if len(m.scripts) == 0 {
		s.WriteString(dimStyle.Render("No scripts found.") + "\n")
		s.WriteString(helpStyle.Render("\nesc: back"))
		return boxStyle.Render(s.String())
	}

	// Render script list
	for i, script := range m.scripts {
		cursor := "  "
		checkbox := "☐"
		style := normalItemStyle

		if i == m.cursor {
			cursor = "▸ "
		}

		if m.selected[i] {
			checkbox = "☑"
			if i == m.cursor {
				style = selectedItemStyle
			} else {
				style = dimStyle
			}
		} else {
			if i == m.cursor {
				style = selectedItemStyle
			} else {
				style = dimStyle
			}
		}

		scriptName := script.File
		tagInfo := ""
		if len(script.Tags) > 0 {
			tags := strings.Join(script.Tags, ", ")
			tagInfo = lipgloss.NewStyle().
				Foreground(accentColor).
				Render(fmt.Sprintf(" [%s]", tags))
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, scriptName, tagInfo)
		s.WriteString(style.Render(line) + "\n")
	}

	selectedCount := len(m.GetSelectedScripts())
	s.WriteString(fmt.Sprintf("\nSelected: %d/%d\n", selectedCount, len(m.scripts)))

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • space: toggle • a: all • n: none • enter: run • esc: cancel"))

	return boxStyle.Render(s.String())
}

// IsConfirmed returns true if user confirmed selection
func (m ScriptSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if user cancelled
func (m ScriptSelectorModel) IsCancelled() bool {
	return m.cancelled
}

// GetSelectedScripts returns the selected scripts
func (m ScriptSelectorModel) GetSelectedScripts() []models.ScriptItem {
	var selected []models.ScriptItem
	for i, script := range m.scripts {
		if m.selected[i] {
			selected = append(selected, script)
		}
	}
	return selected
}

// Helper function for pluralization
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
