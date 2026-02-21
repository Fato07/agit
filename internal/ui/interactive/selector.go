package interactive

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Item represents a selectable item in the list.
type Item struct {
	ID    string
	Label string
	Desc  string
}

// selectorModel is the bubbletea model for single-select.
type selectorModel struct {
	title    string
	items    []Item
	cursor   int
	selected int
	quit     bool
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.selected = -1
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.cursor
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectorModel) View() string {
	if m.quit {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s\n\n", m.title))

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%s", cursor, item.Label)
		if item.Desc != "" {
			line += fmt.Sprintf("  %s", item.Desc)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n  (j/k to move, enter to select, q to cancel)\n")
	return b.String()
}

// Select presents a list of items and returns the selected one.
// Returns an error if the user cancels.
func Select(title string, items []Item) (Item, error) {
	if len(items) == 0 {
		return Item{}, fmt.Errorf("no items to select from")
	}

	m := selectorModel{
		title:    title,
		items:    items,
		selected: -1,
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return Item{}, fmt.Errorf("selector error: %w", err)
	}

	final := result.(selectorModel)
	if final.selected < 0 {
		return Item{}, fmt.Errorf("selection cancelled")
	}

	return items[final.selected], nil
}

// multiSelectModel is the bubbletea model for multi-select.
type multiSelectModel struct {
	title    string
	items    []Item
	cursor   int
	selected map[int]bool
	done     bool
	quit     bool
}

func (m multiSelectModel) Init() tea.Cmd { return nil }

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if m.selected[m.cursor] {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = true
			}
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	if m.done || m.quit {
		return ""
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  %s\n\n", m.title))

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		check := "[ ]"
		if m.selected[i] {
			check = "[x]"
		}
		line := fmt.Sprintf("%s%s %s", cursor, check, item.Label)
		if item.Desc != "" {
			line += fmt.Sprintf("  %s", item.Desc)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n  (j/k to move, space to toggle, enter to confirm, q to cancel)\n")
	return b.String()
}

// MultiSelect presents a list of items and returns the selected ones.
func MultiSelect(title string, items []Item) ([]Item, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select from")
	}

	m := multiSelectModel{
		title:    title,
		items:    items,
		selected: make(map[int]bool),
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("selector error: %w", err)
	}

	final := result.(multiSelectModel)
	if final.quit {
		return nil, fmt.Errorf("selection cancelled")
	}

	var selected []Item
	for i, item := range items {
		if final.selected[i] {
			selected = append(selected, item)
		}
	}
	return selected, nil
}
