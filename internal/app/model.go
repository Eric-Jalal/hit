package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elisa-content-delivery/hit/internal/git"
	gh "github.com/elisa-content-delivery/hit/internal/github"
	"github.com/elisa-content-delivery/hit/internal/styles"
	"github.com/elisa-content-delivery/hit/internal/ui/auth"
	"github.com/elisa-content-delivery/hit/internal/ui/branches"
	"github.com/elisa-content-delivery/hit/internal/ui/ci"
	"github.com/elisa-content-delivery/hit/internal/ui/pr"
	"github.com/elisa-content-delivery/hit/internal/ui/review"
)

type View int

const (
	ViewAuth View = iota
	ViewBranches
	ViewCI
	ViewPR
	ViewReview
)

var tabNames = []string{"Branches", "CI", "PRs", "Reviews"}

type Model struct {
	repo        *git.Repo
	owner       string
	repoName    string
	ghClient    *gh.Client
	token       string
	currentView View
	authModel   auth.Model
	branchModel branches.Model
	ciModel     ci.Model
	prModel     pr.Model
	reviewModel review.Model
	width       int
	height      int
	ready       bool
}

func NewModel(repo *git.Repo, owner, repoName string) Model {
	return Model{
		repo:        repo,
		owner:       owner,
		repoName:    repoName,
		currentView: ViewAuth,
		authModel:   auth.New(),
		branchModel: branches.New(repo),
		prModel:     pr.New(),
		reviewModel: review.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.authModel.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, m.propagateSize(msg)

	case tea.KeyMsg:
		if m.currentView != ViewAuth {
			if cmd, handled := HandleGlobalKeys(msg); handled {
				return m, cmd
			}
		}

	case SwitchViewMsg:
		return m, m.handleViewSwitch(msg)

	case auth.AuthDoneMsg:
		m.token = msg.Token
		client, err := gh.NewClient(m.owner, m.repoName, msg.Token)
		if err == nil {
			m.ghClient = client
			m.ciModel = ci.New(client, m.repo.CurrentBranch())
		}
		m.currentView = ViewBranches
		cmds := []tea.Cmd{m.branchModel.Init()}
		if err == nil {
			cmds = append(cmds, m.ciModel.Init())
		}
		return m, tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	switch m.currentView {
	case ViewAuth:
		m.authModel, cmd = m.authModel.Update(msg)
	case ViewBranches:
		m.branchModel, cmd = m.branchModel.Update(msg)
	case ViewCI:
		m.ciModel, cmd = m.ciModel.Update(msg)
	case ViewPR:
		m.prModel, cmd = m.prModel.Update(msg)
	case ViewReview:
		m.reviewModel, cmd = m.reviewModel.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if !m.ready {
		return "Starting..."
	}

	if m.currentView == ViewAuth {
		return m.authModel.View()
	}

	tabBar := m.renderTabBar()
	var content string
	switch m.currentView {
	case ViewBranches:
		content = m.branchModel.View()
	case ViewCI:
		content = m.ciModel.View()
	case ViewPR:
		content = m.prModel.View()
	case ViewReview:
		content = m.reviewModel.View()
	}

	footer := styles.StatusBarStyle.Render("q: quit | tab: next view | esc: back | ?: help")

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content, footer)
}

func (m Model) renderTabBar() string {
	var tabs []string
	for i, name := range tabNames {
		v := View(i + 1)
		if v == m.currentView {
			tabs = append(tabs, styles.ActiveTabStyle.Render(name))
		} else {
			tabs = append(tabs, styles.InactiveTabStyle.Render(name))
		}
	}
	return styles.TabBarStyle.Render(strings.Join(tabs, " "))
}

func (m *Model) handleViewSwitch(msg SwitchViewMsg) tea.Cmd {
	if msg.View == -1 {
		next := int(m.currentView) + 1
		if next > int(ViewReview) {
			next = int(ViewBranches)
		}
		m.currentView = View(next)
	} else if msg.View == -2 {
		prev := int(m.currentView) - 1
		if prev < int(ViewBranches) {
			prev = int(ViewReview)
		}
		m.currentView = View(prev)
	} else {
		m.currentView = msg.View
	}

	switch m.currentView {
	case ViewBranches:
		return m.branchModel.Init()
	case ViewCI:
		if m.ghClient != nil {
			m.ciModel = ci.New(m.ghClient, m.repo.CurrentBranch())
			return m.ciModel.Init()
		}
	}
	return nil
}

func (m Model) propagateSize(msg tea.WindowSizeMsg) tea.Cmd {
	contentMsg := tea.WindowSizeMsg{
		Width:  msg.Width,
		Height: msg.Height - 4,
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.authModel, cmd = m.authModel.Update(contentMsg)
	cmds = append(cmds, cmd)
	m.branchModel, cmd = m.branchModel.Update(contentMsg)
	cmds = append(cmds, cmd)
	if m.ghClient != nil {
		m.ciModel, cmd = m.ciModel.Update(contentMsg)
		cmds = append(cmds, cmd)
	}
	m.prModel, cmd = m.prModel.Update(contentMsg)
	cmds = append(cmds, cmd)
	m.reviewModel, cmd = m.reviewModel.Update(contentMsg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}
