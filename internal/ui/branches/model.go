package branches

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
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

type Model struct {
	repo   *git.Repo
	list   list.Model
	width  int
	height int
	status string
}

func New(repo *git.Repo) Model {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Branches"
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.TitleStyle
	return Model{repo: repo, list: l}
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
			m.status = fmt.Sprintf("Error: %s", msg.err)
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
			m.status = styles.ErrorLineStyle.Render(fmt.Sprintf("Checkout failed: %s", msg.err))
			return m, nil
		}
		m.status = styles.BadgeSuccess.Render(fmt.Sprintf("Switched to %s", msg.branch))
		return m, m.loadBranches

	case tea.KeyMsg:
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
				m.status = "Already on this branch"
				return m, nil
			}
			m.status = fmt.Sprintf("Checking out %s...", selected.branch.Name)
			return m, m.checkout(selected.branch.Name)

		case "r":
			m.status = "Refreshing..."
			return m, m.loadBranches
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	content := m.list.View()
	if m.status != "" {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.status)
	}
	return content
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
