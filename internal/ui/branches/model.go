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

type Model struct {
	repo       *git.Repo
	list       list.Model
	nameInput  textinput.Model
	creating   bool
	width      int
	height     int
	status     string
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
	return m.creating
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

	case tea.KeyMsg:
		if m.creating {
			return m.handleCreateInput(msg)
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
			m.nameInput.Focus()
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
	if m.creating {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.nameInput.View())
	}
	if m.status != "" {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.status)
	}
	return content
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
