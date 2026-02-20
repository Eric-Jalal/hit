package reflog

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/elisa-content-delivery/hit/internal/git"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

const reflogLimit = 50

// RefreshReflogMsg is sent by other views to trigger a reflog reload.
type RefreshReflogMsg struct{}

type reflogLoadedMsg struct {
	entries []git.ReflogEntry
	err     error
}

type Model struct {
	repo     *git.Repo
	viewport viewport.Model
	entries  []git.ReflogEntry
	ready    bool
	width    int
	height   int
}

func New(repo *git.Repo) Model {
	return Model{repo: repo}
}

func (m Model) Init() tea.Cmd {
	return m.loadReflog
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Account for border (2) + title line with newline (1)
		innerWidth := msg.Width - 2
		innerHeight := msg.Height - 3
		if innerWidth < 0 {
			innerWidth = 0
		}
		if innerHeight < 0 {
			innerHeight = 0
		}
		if !m.ready {
			m.viewport = viewport.New(innerWidth, innerHeight)
			m.viewport.HighPerformanceRendering = false
			m.ready = true
		} else {
			m.viewport.Width = innerWidth
			m.viewport.Height = innerHeight
		}
		if len(m.entries) > 0 {
			m.viewport.SetContent(m.renderEntries(innerWidth))
		}
		return m, nil

	case reflogLoadedMsg:
		if msg.err != nil {
			return m, nil
		}
		m.entries = msg.entries
		if m.ready {
			m.viewport.SetContent(m.renderEntries(m.viewport.Width))
			m.viewport.GotoTop()
		}
		return m, nil

	case RefreshReflogMsg:
		return m, m.loadReflog
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return ""
	}

	title := styles.SubtitleStyle.Render(styles.IconHistory + " Git Reflog")

	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorMuted).
		Width(m.width - 2).
		Height(m.height - 2)

	content := title + "\n" + m.viewport.View()
	return border.Render(content)
}

func (m Model) renderEntries(width int) string {
	if len(m.entries) == 0 {
		return styles.SubtitleStyle.Render("  No reflog entries")
	}

	var b strings.Builder
	for _, e := range m.entries {
		line := m.renderEntry(e, width)
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func (m Model) renderEntry(e git.ReflogEntry, width int) string {
	actionColor := actionStyle(e.Action)

	selector := styles.SubtitleStyle.Render(e.Selector)
	action := actionColor.Render(e.Action)
	timeAgo := styles.SubtitleStyle.Render(e.TimeAgo)

	// Truncate detail to fit within width
	detailMaxWidth := width - lipgloss.Width(selector) - lipgloss.Width(action) - lipgloss.Width(timeAgo) - 6 // spaces between
	detail := e.Detail
	if detailMaxWidth > 0 && len(detail) > detailMaxWidth {
		if detailMaxWidth > 3 {
			detail = detail[:detailMaxWidth-3] + "..."
		} else {
			detail = detail[:detailMaxWidth]
		}
	} else if detailMaxWidth <= 0 {
		detail = ""
	}

	detailRendered := lipgloss.NewStyle().Foreground(styles.ColorText).Render(detail)

	parts := []string{selector, action}
	if detail != "" {
		parts = append(parts, detailRendered)
	}
	parts = append(parts, timeAgo)

	return fmt.Sprintf(" %s", strings.Join(parts, " "))
}

func actionStyle(action string) lipgloss.Style {
	switch {
	case action == "checkout":
		return lipgloss.NewStyle().Foreground(styles.ColorSecondary)
	case action == "commit" || action == "commit (amend)" || action == "commit (initial)":
		return lipgloss.NewStyle().Foreground(styles.ColorSuccess)
	case action == "merge" || action == "cherry-pick":
		return lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	case action == "rebase" || action == "reset":
		return lipgloss.NewStyle().Foreground(styles.ColorWarning)
	default:
		return lipgloss.NewStyle().Foreground(styles.ColorMuted)
	}
}

func (m Model) loadReflog() tea.Msg {
	entries, err := m.repo.GetReflog(reflogLimit)
	return reflogLoadedMsg{entries: entries, err: err}
}
