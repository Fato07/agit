package interactive

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	prompt     string
	defaultYes bool
	result     bool
	done       bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.result = true
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.result = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.result = m.defaultYes
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "q":
			m.result = false
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.done {
		return ""
	}
	hint := "[y/N]"
	if m.defaultYes {
		hint = "[Y/n]"
	}
	return fmt.Sprintf("  %s %s ", m.prompt, hint)
}

// Confirm asks the user a yes/no question.
func Confirm(prompt string, defaultYes bool) (bool, error) {
	m := confirmModel{
		prompt:     strings.TrimSpace(prompt),
		defaultYes: defaultYes,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirm error: %w", err)
	}

	return result.(confirmModel).result, nil
}
