package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a selectable option in the main menu
type MenuItem struct {
	Title       string
	Description string
	Action      string
}

// MenuModel is the main menu TUI model
type MenuModel struct {
	items    []MenuItem
	cursor   int
	selected string
	width    int
	height   int
}

// NewMenuModel creates a new main menu model
func NewMenuModel() MenuModel {
	items := []MenuItem{
		{
			Title:       "📦 Install Packages",
			Description: "Install Homebrew packages and Mac App Store apps",
			Action:      "install",
		},
		{
			Title:       "🔗 Manage Dotfiles",
			Description: "Link, unlink, and manage configuration files",
			Action:      "dotfiles",
		},
		{
			Title:       "📜 Run Scripts",
			Description: "Execute tool setup scripts",
			Action:      "scripts",
		},
		{
			Title:       "🔍 Doctor",
			Description: "Check system prerequisites",
			Action:      "doctor",
		},
		{
			Title:       "❌ Quit",
			Description: "Exit Merlin",
			Action:      "quit",
		},
	}

	return MenuModel{
		items: items,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.selected = "quit"
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.selected = m.items[m.cursor].Action
			if m.selected == "quit" {
				return m, tea.Quit
			}
			// For other actions, return to let parent handle
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	var s strings.Builder

	// Title
	title := titleStyle.Render("✨ Merlin - macOS Dotfiles Manager")
	subtitle := subtitleStyle.Render("A magical tool for managing your macOS setup")
	s.WriteString(title + "\n" + subtitle + "\n\n")

	// Menu items
	for i, item := range m.items {
		cursor := "  "
		style := normalItemStyle

		if i == m.cursor {
			cursor = "▸ "
			style = selectedItemStyle
		}

		title := style.Render(fmt.Sprintf("%s%s", cursor, item.Title))
		s.WriteString(title + "\n")

		// Show description for selected item
		if i == m.cursor {
			desc := lipgloss.NewStyle().
				Foreground(mutedColor).
				PaddingLeft(4).
				Italic(true).
				Render(item.Description)
			s.WriteString(desc + "\n")
		}
		s.WriteString("\n")
	}

	// Help text
	help := helpStyle.Render("\n↑/↓ or j/k: navigate • enter/space: select • q: quit")
	s.WriteString(help)

	return boxStyle.Render(s.String())
}

// GetSelected returns the selected action
func (m MenuModel) GetSelected() string {
	return m.selected
}
