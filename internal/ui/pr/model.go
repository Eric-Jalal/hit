package pr

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type Model struct {
	width  int
	height int
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	return styles.TitleStyle.Render("Pull Requests") + "\n\n" +
		styles.SubtitleStyle.Render("  Coming soon")
}
