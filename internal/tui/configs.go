package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigItem represents a configuration tool with link status
type ConfigItem struct {
	Name        string
	Description string
	IsLinked    bool
	HasConflict bool
	Selected    bool
}

// ConfigSelectorModel allows selecting configs to link/unlink
type ConfigSelectorModel struct {
	title      string
	items      []ConfigItem
	cursor     int
	selected   map[int]bool
	action     string // "link" or "unlink"
	confirmed  bool
	cancelled  bool
	width      int
	height     int
	viewOffset int
}

// NewConfigSelectorModel creates a new config selector
func NewConfigSelectorModel(title string, configs []ConfigItem, action string) ConfigSelectorModel {
	return ConfigSelectorModel{
		title:    title,
		items:    configs,
		action:   action,
		selected: make(map[int]bool),
	}
}

func (m ConfigSelectorModel) Init() tea.Cmd {
	return nil
}

func (m ConfigSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.viewOffset {
					m.viewOffset = m.cursor
				}
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				maxVisible := m.height - 10
				if maxVisible < 5 {
					maxVisible = 5
				}
				if m.cursor >= m.viewOffset+maxVisible {
					m.viewOffset = m.cursor - maxVisible + 1
				}
			}

		case " ", "x":
			m.selected[m.cursor] = !m.selected[m.cursor]
			m.items[m.cursor].Selected = m.selected[m.cursor]

		case "a":
			for i := range m.items {
				m.selected[i] = true
				m.items[i].Selected = true
			}

		case "n":
			m.selected = make(map[int]bool)
			for i := range m.items {
				m.items[i].Selected = false
			}

		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ConfigSelectorModel) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render(m.title) + "\n\n")

	// Items
	maxVisible := m.height - 10
	if maxVisible < 5 {
		maxVisible = 10
	}

	start := m.viewOffset
	end := m.viewOffset + maxVisible
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := start; i < end; i++ {
		item := m.items[i]

		// Status icon
		statusIcon := "○"
		statusColor := mutedColor
		if item.IsLinked {
			statusIcon = "✓"
			statusColor = successColor
		} else if item.HasConflict {
			statusIcon = "⚠"
			statusColor = warningColor
		}

		// Checkbox
		checkbox := "☐"
		if m.selected[i] {
			checkbox = "☑"
		}

		cursor := "  "
		style := normalItemStyle

		if i == m.cursor {
			cursor = "▸ "
			style = selectedItemStyle
		}

		statusStyle := lipgloss.NewStyle().Foreground(statusColor)
		line := fmt.Sprintf("%s%s %s %s", cursor, checkbox, statusStyle.Render(statusIcon), item.Name)
		s.WriteString(style.Render(line) + "\n")

		// Show description for selected item
		if i == m.cursor && item.Description != "" {
			desc := lipgloss.NewStyle().
				Foreground(mutedColor).
				PaddingLeft(8).
				Italic(true).
				Render(item.Description)
			s.WriteString(desc + "\n")
		}
	}

	// Scroll indicator
	if len(m.items) > maxVisible {
		scrollInfo := lipgloss.NewStyle().Foreground(mutedColor).Render(
			fmt.Sprintf("\n  (showing %d-%d of %d)", start+1, end, len(m.items)))
		s.WriteString(scrollInfo)
	}

	// Stats
	selectedCount := len(m.selected)
	stats := fmt.Sprintf("\nSelected: %d/%d", selectedCount, len(m.items))
	s.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(stats) + "\n")

	// Legend
	legend := lipgloss.NewStyle().Foreground(mutedColor).Render(
		"\n✓ linked  ⚠ conflict  ○ not linked")
	s.WriteString(legend + "\n")

	// Help
	actionText := m.action
	help := helpStyle.Render(fmt.Sprintf("\n↑/↓: navigate • space: toggle • a: all • n: none • enter: %s • esc: cancel", actionText))
	s.WriteString(help)

	return boxStyle.Render(s.String())
}

// GetSelectedConfigs returns the selected config names
func (m ConfigSelectorModel) GetSelectedConfigs() []string {
	var selected []string
	for idx := range m.selected {
		if m.selected[idx] {
			selected = append(selected, m.items[idx].Name)
		}
	}
	return selected
}

// IsConfirmed returns true if user confirmed
func (m ConfigSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsCancelled returns true if user cancelled
func (m ConfigSelectorModel) IsCancelled() bool {
	return m.cancelled
}
