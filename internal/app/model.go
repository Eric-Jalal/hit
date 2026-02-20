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
	"github.com/elisa-content-delivery/hit/internal/ui/org"
	"github.com/elisa-content-delivery/hit/internal/ui/pr"
	"github.com/elisa-content-delivery/hit/internal/ui/reflog"
	"github.com/elisa-content-delivery/hit/internal/ui/review"
)

type View int

const (
	ViewAuth View = iota
	ViewBranches
	ViewCI
	ViewPR
	ViewReview
	ViewOrg
)

const reflogPaneWidth = 100

var tabNames = []string{
	styles.IconBranch + " Branches",
	styles.IconGear + " CI",
	styles.IconPR + "  PRs",
	styles.IconEye + "  Reviews",
	styles.IconOrg + "  Org",
}

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
	orgModel    org.Model
	reflogModel reflog.Model
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
		reflogModel: reflog.New(repo),
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
		cmd := m.propagateSize(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.currentView != ViewAuth {
			if m.currentView == ViewOrg && m.orgModel.IsOverlayActive() {
				// let the org model handle all keys when clone overlay is open
			} else if m.currentView == ViewBranches && m.branchModel.IsInputActive() {
				// let the branch model handle all keys when creating a branch
			} else if cmd, handled := HandleGlobalKeys(msg); handled {
				return m, cmd
			}
		}

	case SwitchViewMsg:
		return m, m.handleViewSwitch(msg)

	case reflog.RefreshReflogMsg:
		var cmd tea.Cmd
		m.reflogModel, cmd = m.reflogModel.Update(msg)
		return m, cmd

	case auth.AuthDoneMsg:
		m.token = msg.Token
		client, err := gh.NewClient(m.owner, m.repoName, msg.Token)
		if err == nil {
			m.ghClient = client
			m.ciModel = ci.New(client, m.repo.CurrentBranch())
			m.orgModel = org.New(client)
		}
		m.currentView = ViewBranches
		cmds := []tea.Cmd{m.branchModel.Init(), m.reflogModel.Init()}
		if err == nil {
			cmds = append(cmds, m.ciModel.Init())
		}
		return m, tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch m.currentView {
	case ViewAuth:
		m.authModel, cmd = m.authModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewBranches:
		m.branchModel, cmd = m.branchModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewCI:
		m.ciModel, cmd = m.ciModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewPR:
		m.prModel, cmd = m.prModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewReview:
		m.reviewModel, cmd = m.reviewModel.Update(msg)
		cmds = append(cmds, cmd)
	case ViewOrg:
		m.orgModel, cmd = m.orgModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Forward non-key messages to reflog pane
	if _, isKey := msg.(tea.KeyMsg); !isKey && m.currentView != ViewAuth {
		m.reflogModel, cmd = m.reflogModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if !m.ready {
		return "Starting..."
	}

	if m.currentView == ViewAuth {
		return m.authModel.View()
	}

	tabBar := m.renderTabBar()
	repoInfo := m.renderRepoInfo()
	var content string
	var hints string
	switch m.currentView {
	case ViewBranches:
		content = m.branchModel.View()
		if m.branchModel.IsConfirming() {
			hints = formatHints([][]string{{"y", "rename remote"}, {"n", "local only"}, {"esc", "cancel"}})
		} else if m.branchModel.IsInputActive() {
			hints = formatHints([][]string{{"enter", "confirm"}, {"esc", "cancel"}})
		} else {
			hints = formatHints([][]string{{"enter", "checkout"}, {"p", "push"}, {"a", "new branch"}, {"R", "rename"}, {"r", "refresh"}, {"/", "filter"}, {"tab", "next view"}, {"q", "quit"}})
		}
	case ViewCI:
		content = m.ciModel.View()
		hints = formatHints([][]string{{"enter", "details"}, {"esc", "back"}, {"r", "refresh"}, {"tab", "next view"}, {"q", "quit"}})
	case ViewPR:
		content = m.prModel.View()
		hints = formatHints([][]string{{"tab", "next view"}, {"q", "quit"}})
	case ViewReview:
		content = m.reviewModel.View()
		hints = formatHints([][]string{{"tab", "next view"}, {"q", "quit"}})
	case ViewOrg:
		content = m.orgModel.View()
		if m.orgModel.IsOverlayActive() {
			hints = formatHints([][]string{{"enter", "clone"}, {"esc", "cancel"}})
		} else {
			hints = formatHints([][]string{{"enter", "select"}, {"esc", "back"}, {"r", "refresh"}, {"tab", "next view"}, {"q", "quit"}})
		}
	}
	footer := styles.StatusBarStyle.Render(hints)

	contentHeight := m.height - lipgloss.Height(tabBar) - lipgloss.Height(repoInfo) - lipgloss.Height(footer)

	showReflog := m.width >= 200
	if showReflog {
		mainWidth := m.width - reflogPaneWidth
		mainContent := lipgloss.NewStyle().Width(mainWidth).Height(contentHeight).Render(content)
		reflogContent := m.reflogModel.View()
		content = lipgloss.JoinHorizontal(lipgloss.Top, mainContent, reflogContent)
	} else {
		content = lipgloss.NewStyle().Width(m.width).Height(contentHeight).Render(content)
	}

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, repoInfo, content, footer)
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

func (m Model) renderRepoInfo() string {
	name := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorSecondary).Render(m.owner + "/" + m.repoName)
	path := styles.SubtitleStyle.Render(m.repo.Path())

	parts := []string{name, path}

	branch := m.repo.CurrentBranch()
	if branch != "" && m.repo.HasUpstream(branch) {
		remote := styles.SubtitleStyle.Render(m.repo.RemoteURL())
		parts = append(parts, remote)
	}

	sep := styles.SubtitleStyle.Render("  ")
	return lipgloss.NewStyle().MarginLeft(1).Render(strings.Join(parts, sep))
}

func (m *Model) handleViewSwitch(msg SwitchViewMsg) tea.Cmd {
	if msg.View == -1 {
		next := int(m.currentView) + 1
		if next > int(ViewOrg) {
			next = int(ViewBranches)
		}
		m.currentView = View(next)
	} else if msg.View == -2 {
		prev := int(m.currentView) - 1
		if prev < int(ViewBranches) {
			prev = int(ViewOrg)
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
	case ViewOrg:
		if m.ghClient != nil {
			return m.orgModel.Init()
		}
	}
	return nil
}

func formatHints(pairs [][]string) string {
	key := lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true)
	desc := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	sep := desc.Render(" · ")
	var parts []string
	for _, p := range pairs {
		parts = append(parts, key.Render(p[0])+" "+desc.Render(p[1]))
	}
	return strings.Join(parts, sep)
}

func (m *Model) propagateSize(msg tea.WindowSizeMsg) tea.Cmd {
	tabBar := m.renderTabBar()
	repoInfo := m.renderRepoInfo()
	footer := styles.StatusBarStyle.Render("q: quit | tab: next view | esc: back | ?: help")
	contentHeight := msg.Height - lipgloss.Height(tabBar) - lipgloss.Height(repoInfo) - lipgloss.Height(footer)

	showReflog := msg.Width >= 160
	mainWidth := msg.Width
	if showReflog {
		mainWidth = msg.Width - reflogPaneWidth
	}

	contentMsg := tea.WindowSizeMsg{
		Width:  mainWidth,
		Height: contentHeight,
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
	if m.ghClient != nil {
		m.orgModel, cmd = m.orgModel.Update(contentMsg)
		cmds = append(cmds, cmd)
	}

	if showReflog {
		reflogMsg := tea.WindowSizeMsg{
			Width:  reflogPaneWidth,
			Height: contentHeight,
		}
		m.reflogModel, cmd = m.reflogModel.Update(reflogMsg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}
