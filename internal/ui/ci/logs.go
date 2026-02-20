package ci

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elisa-content-delivery/hit/internal/styles"
)

type LogView struct {
	viewport viewport.Model
	rawLog   string
	ready    bool
}

func NewLogView() LogView {
	return LogView{}
}

func (l LogView) Update(msg tea.Msg) (LogView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !l.ready {
			l.viewport = viewport.New(msg.Width, msg.Height-6)
			l.viewport.HighPerformanceRendering = false
			l.ready = true
		} else {
			l.viewport.Width = msg.Width
			l.viewport.Height = msg.Height - 6
		}
		if l.rawLog != "" {
			l.viewport.SetContent(highlightErrors(l.rawLog))
		}
	}

	var cmd tea.Cmd
	l.viewport, cmd = l.viewport.Update(msg)
	return l, cmd
}

func (l *LogView) SetContent(log string) {
	l.rawLog = log
	if l.ready {
		l.viewport.SetContent(highlightErrors(log))
		l.viewport.GotoBottom()
	}
}

func (l LogView) View() string {
	if !l.ready {
		return "Loading..."
	}
	return l.viewport.View()
}

func highlightErrors(log string) string {
	var b strings.Builder
	for _, line := range strings.Split(log, "\n") {
		if IsErrorLine(line) {
			b.WriteString(styles.ErrorLineStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteByte('\n')
	}
	return b.String()
}
