package branches

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elisa-content-delivery/hit/internal/git"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type checkoutDoneMsg struct {
	branch string
	err    error
}

type branchesLoadedMsg struct {
	branches []git.Branch
	err      error
}

type branchCreatedMsg struct {
	branch string
	err    error
}

type branchRenamedMsg struct {
	oldName     string
	newName     string
	renamedRemote bool
	err         error
}

type Model struct {
	repo              *git.Repo
	list              list.Model
	nameInput         textinput.Model
	creating          bool
	renaming          bool
	renamingFrom      string
	confirmRemote     bool
	pendingOldName    string
	pendingNewName    string
	width             int
	height            int
	status            string
}

func New(repo *git.Repo) Model {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.SetStatusBarItemName("branch", "branches")
	l.Styles.Title = styles.TitleStyle

	ti := textinput.New()
	ti.Prompt = "New branch: "
	ti.CharLimit = 128

	return Model{repo: repo, list: l, nameInput: ti}
}

func (m Model) IsInputActive() bool {
	return m.creating || m.renaming || m.confirmRemote
}

func (m Model) IsConfirming() bool {
	return m.confirmRemote
}

func (m Model) Init() tea.Cmd {
	return m.loadBranches
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case branchesLoadedMsg:
		if msg.err != nil {
			m.status = styles.ErrorLineStyle.Render("Error: ") + styles.SubtitleStyle.Render(msg.err.Error())
			return m, nil
		}
		items := make([]list.Item, len(msg.branches))
		for i, b := range msg.branches {
			items[i] = branchItem{branch: b}
		}
		m.list.SetItems(items)
		m.status = ""
		return m, nil

	case checkoutDoneMsg:
		if msg.err != nil {
			m.status = styles.ErrorLineStyle.Render("Checkout failed: ") + styles.SubtitleStyle.Render(msg.err.Error())
			return m, nil
		}
		m.status = styles.BadgeSuccess.Render("Switched to ") + styles.HighlightStyle.Render(msg.branch)
		return m, m.loadBranches

	case branchCreatedMsg:
		if msg.err != nil {
			m.status = styles.ErrorLineStyle.Render("Create failed: ") + styles.SubtitleStyle.Render(msg.err.Error())
			return m, nil
		}
		m.status = styles.BadgeSuccess.Render("Created and switched to ") + styles.HighlightStyle.Render(msg.branch)
		return m, m.loadBranches

	case branchRenamedMsg:
		if msg.err != nil {
			m.status = styles.ErrorLineStyle.Render("Rename failed: ") + styles.SubtitleStyle.Render(msg.err.Error())
			return m, nil
		}
		status := styles.BadgeSuccess.Render("Renamed ") + styles.HighlightStyle.Render(msg.oldName) + styles.BadgeSuccess.Render(" to ") + styles.HighlightStyle.Render(msg.newName)
		if msg.renamedRemote {
			status += styles.BadgeSuccess.Render(" (local + remote)")
		}
		m.status = status
		return m, m.loadBranches

	case tea.KeyMsg:
		if m.confirmRemote {
			return m.handleConfirmRemote(msg)
		}
		if m.creating {
			return m.handleCreateInput(msg)
		}
		if m.renaming {
			return m.handleRenameInput(msg)
		}

		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "enter":
			selected, ok := m.list.SelectedItem().(branchItem)
			if !ok {
				return m, nil
			}
			if selected.branch.IsCurrent {
				m.status = styles.BadgeNeutral.Render("Already on ") + styles.HighlightStyle.Render(selected.branch.Name)
				return m, nil
			}
			m.status = styles.BadgePending.Render("Checking out ") + styles.HighlightStyle.Render(selected.branch.Name) + styles.BadgePending.Render("...")
			return m, m.checkout(selected.branch.Name)

		case "r":
			m.status = styles.BadgePending.Render("Refreshing...")
			return m, m.loadBranches

		case "a":
			m.creating = true
			m.nameInput.SetValue("")
			m.nameInput.Prompt = "New branch: "
			m.nameInput.Focus()
			m.status = ""
			return m, m.nameInput.Cursor.BlinkCmd()

		case "R":
			selected, ok := m.list.SelectedItem().(branchItem)
			if !ok {
				return m, nil
			}
			m.renaming = true
			m.renamingFrom = selected.branch.Name
			m.nameInput.Prompt = "Rename to: "
			m.nameInput.SetValue(selected.branch.Name)
			m.nameInput.Focus()
			m.nameInput.CursorEnd()
			m.status = ""
			return m, m.nameInput.Cursor.BlinkCmd()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	content := m.list.View()
	if m.creating || m.renaming {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.nameInput.View())
	}
	if m.status != "" {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.status)
	}
	if m.confirmRemote {
		content = m.renderConfirmOverlay(content)
	}
	return content
}

func (m Model) renderConfirmOverlay(_ string) string {
	title := styles.TitleStyle.Render("Rename remote branch?")
	desc := styles.SubtitleStyle.Render(m.pendingOldName+" exists on remote.\nThis will delete the old remote branch and force push the new name.")
	hint := styles.HighlightStyle.Render("y") + styles.SubtitleStyle.Render(": yes, rename remote too") +
		"\n" + styles.HighlightStyle.Render("n") + styles.SubtitleStyle.Render(": no, local only") +
		"\n" + styles.HighlightStyle.Render("esc") + styles.SubtitleStyle.Render(": cancel")

	body := title + "\n\n" + desc + "\n\n" + hint

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorWarning).
		Padding(1, 2).
		Width(56).
		Render(body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func (m Model) handleCreateInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.creating = false
		return m, nil
	case "enter":
		name := m.nameInput.Value()
		if name == "" {
			return m, nil
		}
		m.creating = false
		m.status = styles.BadgePending.Render("Creating ") + styles.HighlightStyle.Render(name) + styles.BadgePending.Render("...")
		return m, m.createBranch(name)
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m Model) handleRenameInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.renaming = false
		return m, nil
	case "enter":
		newName := m.nameInput.Value()
		if newName == "" || newName == m.renamingFrom {
			m.renaming = false
			return m, nil
		}
		m.renaming = false
		if m.repo.HasUpstream(m.renamingFrom) {
			m.confirmRemote = true
			m.pendingOldName = m.renamingFrom
			m.pendingNewName = newName
			return m, nil
		}
		m.status = styles.BadgePending.Render("Renaming ") + styles.HighlightStyle.Render(m.renamingFrom) + styles.BadgePending.Render("...")
		return m, m.renameBranch(m.renamingFrom, newName, false)
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m Model) handleConfirmRemote(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.confirmRemote = false
		m.status = styles.BadgePending.Render("Renaming ") + styles.HighlightStyle.Render(m.pendingOldName) + styles.BadgePending.Render(" (local + remote)...")
		return m, m.renameBranch(m.pendingOldName, m.pendingNewName, true)
	case "n":
		m.confirmRemote = false
		m.status = styles.BadgePending.Render("Renaming ") + styles.HighlightStyle.Render(m.pendingOldName) + styles.BadgePending.Render(" (local only)...")
		return m, m.renameBranch(m.pendingOldName, m.pendingNewName, false)
	case "esc":
		m.confirmRemote = false
		return m, nil
	}
	return m, nil
}

func (m Model) loadBranches() tea.Msg {
	branches, err := m.repo.ListBranches()
	return branchesLoadedMsg{branches: branches, err: err}
}

func (m Model) checkout(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.Checkout(name)
		return checkoutDoneMsg{branch: name, err: err}
	}
}

func (m Model) createBranch(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.CreateBranch(name)
		return branchCreatedMsg{branch: name, err: err}
	}
}

func (m Model) renameBranch(oldName, newName string, renameRemote bool) tea.Cmd {
	return func() tea.Msg {
		err := m.repo.RenameBranch(oldName, newName)
		if err != nil {
			return branchRenamedMsg{oldName: oldName, newName: newName, err: err}
		}
		if renameRemote {
			err = m.repo.RenameRemoteBranch(oldName, newName)
		}
		return branchRenamedMsg{oldName: oldName, newName: newName, renamedRemote: renameRemote, err: err}
	}
}
