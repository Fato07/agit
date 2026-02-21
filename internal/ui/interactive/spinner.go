package interactive

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ err error }

type spinnerModel struct {
	spinner spinner.Model
	msg     string
	fn      func() error
	err     error
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		if err := m.fn(); err != nil {
			return errMsg{err}
		}
		return errMsg{}
	})
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errMsg:
		m.err = msg.err
		m.done = true
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.done = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return fmt.Sprintf("  %s %s\n", m.spinner.View(), m.msg)
}

// WithSpinner shows a spinner while fn executes.
func WithSpinner(msg string, fn func() error) error {
	s := spinner.New()
	s.Spinner = spinner.Dot

	m := spinnerModel{
		spinner: s,
		msg:     msg,
		fn:      fn,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return fmt.Errorf("spinner error: %w", err)
	}

	return result.(spinnerModel).err
}
