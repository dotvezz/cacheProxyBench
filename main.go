package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/profiles"
	"github.com/dotvezz/cacheProxyBench/ui"
)

func profilesMenu(m tea.Model) tea.Model {
	ps := profiles.GetAll()

	items := make([]ui.Item, len(ps))
	for i, p := range ps {
		items[i] = ui.Item{Name: p.Name, Value: nil}
	}

	return ui.New(ui.WithParent(m), ui.WithTitle("Select a profile"), ui.WithItems(items))
}

func main() {
	menuItems := []ui.Item{
		{Name: "(R)un", HotKey: 'r', Value: profilesMenu},
		{Name: "(Q)uit", Value: func(_ tea.Model) tea.Model { return ui.Quitter{} }},
	}

	m := ui.New(
		ui.WithItems(menuItems),
		ui.WithTitle("Select an option"),
	)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
