package auth

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gh "github.com/elisa-content-delivery/hit/internal/github"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type AuthDoneMsg struct {
	Token string
}

type Model struct {
	input  textinput.Model
	status gh.AuthStatus
	width  int
	height int
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	ti.EchoMode = textinput.EchoPassword
	ti.CharLimit = 100
	ti.Width = 50
	return Model{input: ti}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		status := gh.DetectAuth()
		if status.Authenticated {
			return AuthDoneMsg{Token: status.Token}
		}
		return status
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case gh.AuthStatus:
		m.status = msg
		m.input.Focus()
		return m, textinput.Blink

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			token := m.input.Value()
			if token != "" {
				return m, func() tea.Msg {
					return AuthDoneMsg{Token: token}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	title := styles.TitleStyle.Render("GitHub Authentication")

	body := lipgloss.NewStyle().MarginLeft(2).Render(
		"No GitHub token detected.\n\n" +
			"Option 1: Run " + styles.HighlightStyle.Render("gh auth login") + " in another terminal\n" +
			"Option 2: Paste a Personal Access Token below\n\n" +
			m.input.View() + "\n\n" +
			styles.HelpStyle.Render("Press enter to authenticate"),
	)

	return lipgloss.JoinVertical(lipgloss.Left, title, "", body)
}
