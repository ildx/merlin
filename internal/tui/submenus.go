package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PackageTypeMenu allows selecting package type
type PackageTypeMenu struct {
	items     []string
	cursor    int
	selected  string
	cancelled bool
}

// NewPackageTypeMenu creates a menu for selecting package type
func NewPackageTypeMenu() PackageTypeMenu {
	return PackageTypeMenu{
		items: []string{
			"🔧 Formulae only",
			"📱 Casks only",
			"📦 Both formulae and casks",
		},
	}
}

func (m PackageTypeMenu) Init() tea.Cmd {
	return nil
}

func (m PackageTypeMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			switch m.cursor {
			case 0:
				m.selected = "formulae"
			case 1:
				m.selected = "casks"
			case 2:
				m.selected = "both"
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m PackageTypeMenu) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("📦 Select Package Type") + "\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalItemStyle

		if i == m.cursor {
			cursor = "▸ "
			style = selectedItemStyle
		}

		s.WriteString(style.Render(cursor+item) + "\n")
	}

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select • esc: cancel"))

	return boxStyle.Render(s.String())
}

// ConfigActionMenu allows selecting link or unlink action
type ConfigActionMenu struct {
	items     []string
	cursor    int
	selected  string
	cancelled bool
}

// NewConfigActionMenu creates a menu for selecting config action
func NewConfigActionMenu() ConfigActionMenu {
	return ConfigActionMenu{
		items: []string{
			"🔗 Link configs",
			"🔓 Unlink configs",
		},
	}
}

func (m ConfigActionMenu) Init() tea.Cmd {
	return nil
}

func (m ConfigActionMenu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			switch m.cursor {
			case 0:
				m.selected = "link"
			case 1:
				m.selected = "unlink"
			}
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ConfigActionMenu) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🔗 Manage Dotfiles") + "\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := normalItemStyle

		if i == m.cursor {
			cursor = "▸ "
			style = selectedItemStyle
		}

		s.WriteString(style.Render(cursor+item) + "\n")
	}

	s.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select • esc: cancel"))

	return boxStyle.Render(s.String())
}
