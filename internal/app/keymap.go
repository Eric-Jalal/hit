package app

import "github.com/charmbracelet/bubbletea"

type SwitchViewMsg struct {
	View View
}

func HandleGlobalKeys(msg tea.KeyMsg) (tea.Cmd, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return tea.Quit, true
	case "tab":
		return func() tea.Msg {
			return SwitchViewMsg{View: -1}
		}, true
	case "shift+tab":
		return func() tea.Msg {
			return SwitchViewMsg{View: -2}
		}, true
	}
	return nil, false
}
