package org

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/elisa-content-delivery/hit/internal/github"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type pane int

const (
	paneOrgs pane = iota
	paneRepos
)

type orgsLoadedMsg struct {
	orgs []gh.Org
	err  error
}

type reposLoadedMsg struct {
	repos []gh.OrgRepo
	err   error
}

type cloneDoneMsg struct {
	repoName string
	err      error
}

type Model struct {
	client           *gh.Client
	currentPane      pane
	orgsList         list.Model
	reposList        list.Model
	spinner          spinner.Model
	loading          bool
	selectedOrg      *gh.Org
	showCloneOverlay bool
	cloneInput       textinput.Model
	cloneSSHURL      string
	cloneRepoName    string
	cloning          bool
	width            int
	height           int
	status           string
}

func New(client *gh.Client) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	makeList := func(title string) list.Model {
		l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
		l.Title = title
		l.SetShowHelp(false)
		l.SetFilteringEnabled(false)
		l.Styles.Title = styles.TitleStyle
		return l
	}

	ti := textinput.New()
	ti.Prompt = "Path: "
	ti.CharLimit = 256

	return Model{
		client:      client,
		currentPane: paneOrgs,
		orgsList:    makeList("Organizations"),
		reposList:   makeList("Repos"),
		spinner:     s,
		cloneInput:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadOrgs)
}

func (m Model) IsOverlayActive() bool {
	return m.showCloneOverlay
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := msg.Height - 4
		m.orgsList.SetSize(msg.Width, listHeight)
		m.reposList.SetSize(msg.Width, listHeight)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case orgsLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.err)
			return m, nil
		}
		items := make([]list.Item, len(msg.orgs))
		for i, o := range msg.orgs {
			items[i] = orgItem{org: o}
		}
		m.orgsList.SetItems(items)
		return m, nil

	case reposLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Error: %s", msg.err)
			m.currentPane = paneOrgs
			return m, nil
		}
		items := make([]list.Item, len(msg.repos))
		for i, r := range msg.repos {
			items[i] = repoItem{repo: r}
		}
		m.reposList.SetItems(items)
		m.currentPane = paneRepos
		return m, nil

	case cloneDoneMsg:
		m.cloning = false
		m.showCloneOverlay = false
		if msg.err != nil {
			m.status = fmt.Sprintf("Clone failed: %s", msg.err)
		} else {
			m.status = fmt.Sprintf("Cloned %s successfully", msg.repoName)
		}
		return m, nil

	case tea.KeyMsg:
		if m.showCloneOverlay {
			return m.handleOverlayKey(msg)
		}

		switch msg.String() {
		case "esc":
			return m.goBack()
		case "enter":
			return m.drillDown()
		case "r":
			if m.currentPane == paneOrgs {
				m.loading = true
				m.status = ""
				return m, tea.Batch(m.spinner.Tick, m.loadOrgs)
			}
		}
	}

	var cmd tea.Cmd
	switch m.currentPane {
	case paneOrgs:
		m.orgsList, cmd = m.orgsList.Update(msg)
	case paneRepos:
		m.reposList, cmd = m.reposList.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.loading && !m.cloning {
		return m.spinner.View() + " Loading..."
	}

	var content string
	switch m.currentPane {
	case paneOrgs:
		content = m.orgsList.View()
	case paneRepos:
		content = m.reposList.View()
	}

	if m.status != "" {
		content += "\n" + lipgloss.NewStyle().MarginLeft(2).Render(m.status)
	}

	nav := m.breadcrumb()
	view := nav + "\n" + content

	if m.showCloneOverlay {
		view = m.renderOverlay(view)
	}

	return view
}

func (m Model) breadcrumb() string {
	parts := []string{styles.SubtitleStyle.Render("Org")}
	if m.selectedOrg != nil {
		parts = append(parts, styles.HighlightStyle.Render(m.selectedOrg.Login))
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += styles.SubtitleStyle.Render(" > ") + p
	}
	return result
}

func (m Model) goBack() (Model, tea.Cmd) {
	if m.currentPane == paneRepos {
		m.currentPane = paneOrgs
		m.selectedOrg = nil
		m.status = ""
	}
	return m, nil
}

func (m Model) drillDown() (Model, tea.Cmd) {
	switch m.currentPane {
	case paneOrgs:
		selected, ok := m.orgsList.SelectedItem().(orgItem)
		if !ok {
			return m, nil
		}
		m.selectedOrg = &selected.org
		m.loading = true
		return m, tea.Batch(m.spinner.Tick, m.loadRepos(selected.org.Login))

	case paneRepos:
		selected, ok := m.reposList.SelectedItem().(repoItem)
		if !ok {
			return m, nil
		}
		m.cloneSSHURL = selected.repo.SSHURL
		m.cloneRepoName = selected.repo.Name
		m.showCloneOverlay = true
		home, _ := os.UserHomeDir()
		m.cloneInput.SetValue(filepath.Join(home, selected.repo.Name))
		m.cloneInput.Focus()
		m.cloneInput.CursorEnd()
		return m, m.cloneInput.Cursor.BlinkCmd()
	}
	return m, nil
}

func (m Model) handleOverlayKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showCloneOverlay = false
		return m, nil
	case "enter":
		target := m.cloneInput.Value()
		if target == "" {
			return m, nil
		}
		m.cloning = true
		return m, m.cloneRepo(m.cloneSSHURL, target, m.cloneRepoName)
	}

	var cmd tea.Cmd
	m.cloneInput, cmd = m.cloneInput.Update(msg)
	return m, cmd
}

func (m Model) renderOverlay(bg string) string {
	var body string
	if m.cloning {
		body = m.spinner.View() + " Cloning..."
	} else {
		title := styles.TitleStyle.Render("Clone " + m.cloneRepoName)
		input := m.cloneInput.View()
		hint := styles.SubtitleStyle.Render("enter: clone  esc: cancel")
		body = title + "\n\n" + input + "\n\n" + hint
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorPrimary).
		Padding(1, 2).
		Width(60).
		Render(body)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func (m Model) loadOrgs() tea.Msg {
	orgs, err := m.client.GetUserOrgs()
	return orgsLoadedMsg{orgs: orgs, err: err}
}

func (m Model) loadRepos(orgLogin string) tea.Cmd {
	return func() tea.Msg {
		repos, err := m.client.GetOrgRepos(orgLogin)
		return reposLoadedMsg{repos: repos, err: err}
	}
}

func (m Model) cloneRepo(sshURL, targetPath, repoName string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "clone", sshURL, targetPath)
		err := cmd.Run()
		return cloneDoneMsg{repoName: repoName, err: err}
	}
}
