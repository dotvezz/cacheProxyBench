package ui

import "strings"

import tea "charm.land/bubbletea/v2"

type ModelFunc func() tea.Model

type MenuOption struct {
	Key   string
	Value ModelFunc
}

type Menu struct {
	cursor  int
	options []MenuOption
}

func (m *Menu) Init() tea.Cmd {
	return nil
}

func (m *Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.options) - 1
			}
		case tea.KeyDown:
			m.cursor++
			if m.cursor >= len(m.options) {
				m.cursor = 0
			}
		case tea.KeyEnter:
			if m.options != nil {
				return m.options[m.cursor].Value(), nil
			}
		case 'q':
			return nil, tea.Quit
		}

	}

	return m, nil
}

func (m *Menu) View() tea.View {
	var s strings.Builder
	for i := range m.options {
		if i == m.cursor {
			s.WriteString("-> ")
		} else {
			s.WriteString("   ")
		}
		s.WriteString(m.options[i].Key + "\n")
	}

	return tea.NewView(s.String())
}

func NewMenu(options []MenuOption) *Menu {
	return &Menu{
		options: options,
	}
}
