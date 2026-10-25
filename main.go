package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

var (
	_ tea.Model = model{}
)

type model struct {
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) View() tea.View {
	//TODO implement me
	panic("implement me")
}

func newModel() model {
	return model{}
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
