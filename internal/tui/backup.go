package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ildx/merlin/internal/backup"
)

// BackupItem represents a backup in the list
type BackupItem struct {
	manifest *backup.BackupManifest
}

func (i BackupItem) FilterValue() string {
	return i.manifest.ID + " " + i.manifest.Reason
}

func (i BackupItem) Title() string {
	return fmt.Sprintf("%s - %s", i.manifest.ID, i.manifest.Timestamp.Format("2006-01-02 15:04"))
}

func (i BackupItem) Description() string {
	return fmt.Sprintf("%s (%d files)", i.manifest.Reason, len(i.manifest.Files))
}

// BackupListModel shows available backups
type BackupListModel struct {
	list     list.Model
	selected *backup.BackupManifest
	quitting bool
	width    int
	height   int
}

// NewBackupListModel creates a new backup list model
func NewBackupListModel() (BackupListModel, error) {
	backups, err := backup.ListBackups()
	if err != nil {
		return BackupListModel{}, err
	}

	items := make([]list.Item, len(backups))
	for i, b := range backups {
		items[i] = BackupItem{manifest: b}
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Configuration Backups"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "restore"),
			),
			key.NewBinding(
				key.WithKeys("d"),
				key.WithHelp("d", "delete"),
			),
		}
	}

	return BackupListModel{
		list: l,
	}, nil
}

func (m BackupListModel) Init() tea.Cmd {
	return nil
}

func (m BackupListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if item, ok := m.list.SelectedItem().(BackupItem); ok {
				m.selected = item.manifest
				return m, tea.Quit
			}
			return m, nil

		case "d":
			if item, ok := m.list.SelectedItem().(BackupItem); ok {
				// Delete backup
				if err := backup.DeleteBackup(item.manifest.ID); err == nil {
					// Remove from list
					m.list.RemoveItem(m.list.Index())
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m BackupListModel) View() string {
	if m.quitting && m.selected == nil {
		return ""
	}

	return docStyle.Render(m.list.View())
}

// BackupDetailsModel shows backup file details before restore
type BackupDetailsModel struct {
	manifest *backup.BackupManifest
	cursor   int
	selected []bool // Track which files to restore
	confirm  bool
	width    int
	height   int
}

// NewBackupDetailsModel creates a new backup details model
func NewBackupDetailsModel(manifest *backup.BackupManifest) BackupDetailsModel {
	selected := make([]bool, len(manifest.Files))
	// Select all by default
	for i := range selected {
		selected[i] = true
	}

	return BackupDetailsModel{
		manifest: manifest,
		selected: selected,
	}
}

func (m BackupDetailsModel) Init() tea.Cmd {
	return nil
}

func (m BackupDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.manifest.Files)-1 {
				m.cursor++
			}
			return m, nil

		case " ":
			// Toggle selection
			m.selected[m.cursor] = !m.selected[m.cursor]
			return m, nil

		case "a":
			// Select all
			for i := range m.selected {
				m.selected[i] = true
			}
			return m, nil

		case "n":
			// Deselect all
			for i := range m.selected {
				m.selected[i] = false
			}
			return m, nil

		case "enter":
			// Proceed to restore
			m.confirm = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m BackupDetailsModel) View() string {
	if m.confirm {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render(fmt.Sprintf("Restore Backup: %s", m.manifest.ID)))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Created: %s\n", m.manifest.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Reason: %s\n", m.manifest.Reason))
	b.WriteString("\n")

	// File list
	b.WriteString(subtitleStyle.Render("Select files to restore:"))
	b.WriteString("\n\n")

	for i, entry := range m.manifest.Files {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[✓]"
		}

		style := lipgloss.NewStyle()
		if m.cursor == i {
			style = style.Foreground(lipgloss.Color("212"))
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, entry.OriginalPath)
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	helpText := dimStyle.Render("space: toggle • a: select all • n: none • enter: restore • q: cancel")
	b.WriteString(helpText)

	return docStyle.Render(b.String())
}

// BackupRestoreModel handles the restore operation
type BackupRestoreModel struct {
	manifest      *backup.BackupManifest
	selectedFiles []string
	status        string
	done          bool
	err           error
}

// NewBackupRestoreModel creates a new restore model
func NewBackupRestoreModel(manifest *backup.BackupManifest, selectedFiles []string) BackupRestoreModel {
	return BackupRestoreModel{
		manifest:      manifest,
		selectedFiles: selectedFiles,
	}
}

func (m BackupRestoreModel) Init() tea.Cmd {
	return m.restore()
}

func (m BackupRestoreModel) restore() tea.Cmd {
	return func() tea.Msg {
		err := backup.RestoreBackup(m.manifest.ID, m.selectedFiles)
		return backupRestoreDoneMsg{err: err}
	}
}

type backupRestoreDoneMsg struct {
	err error
}

func (m BackupRestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case backupRestoreDoneMsg:
		m.done = true
		m.err = msg.err
		if msg.err == nil {
			m.status = "✅ Backup restored successfully"
		} else {
			m.status = fmt.Sprintf("❌ Restore failed: %v", msg.err)
		}
		return m, tea.Quit

	case tea.KeyMsg:
		if m.done {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m BackupRestoreModel) View() string {
	if !m.done {
		return progressStyle.Render("Restoring files...")
	}

	return docStyle.Render(m.status)
}
