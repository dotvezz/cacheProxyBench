package menu

import (
	"strings"
)

import tea "charm.land/bubbletea/v2"

type ModelFunc func(parent tea.Model) tea.Model

type Item struct {
	Name   string
	Value  ModelFunc
	HotKey rune
}

type Menu struct {
	cursor int
	items  []Item
	title  string
	footer string
	parent tea.Model
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
				m.cursor = len(m.items) - 1
			}
		case tea.KeyDown:
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
		case tea.KeyEnter:
			if m.items != nil {
				return m.items[m.cursor].Value(m), nil
			}
		case 'q':
			if m.parent == nil {
				return m, tea.Quit
			}
		case 'b':
			if m.parent != nil {
				return m.parent, nil
			}
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		for i := range m.items {
			if m.items[i].HotKey == msg.Code {
				return m.items[i].Value(m), nil
			}
		}

	}

	return m, nil
}

func (m *Menu) View() tea.View {
	var s strings.Builder

	if m.title != "" {
		s.WriteString("   " + m.title + "\n\n")
	}

	for i := range m.items {
		if i == m.cursor {
			s.WriteString("-> ")
		} else {
			s.WriteString("   ")
		}
		s.WriteString(m.items[i].Name + "\n")
	}

	if m.parent != nil {
		s.WriteString("\n\n   Press B to go back")
	} else {
		s.WriteString("\n\n   Press Q to quit")
	}

	return tea.NewView(s.String())
}

type Option func(*Menu)

func WithItems(is []Item) Option {
	return func(m *Menu) {
		m.items = is
	}
}

func WithTitle(t string) Option {
	return func(m *Menu) {
		m.title = t
	}
}

func WithParent(p tea.Model) Option {
	return func(m *Menu) {
		m.parent = p
	}
}

func New(options ...Option) *Menu {
	m := new(Menu)
	for i := range options {
		options[i](m)
	}
	return m
}
