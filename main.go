package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/profiles"
	"github.com/dotvezz/cacheProxyBench/ui/menu"
)

func profilesMenu(m tea.Model) tea.Model {
	ps := profiles.GetAll()

	items := make([]menu.Item, len(ps))
	for i, p := range ps {
		items[i] = menu.Item{Name: p.Name, Value: nil}
	}

	return menu.New(menu.WithParent(m), menu.WithTitle("Select a profile"), menu.WithItems(items))
}

func main() {
	menuItems := []menu.Item{
		{Name: "(R)un", HotKey: 'r', Value: profilesMenu},
		{Name: "", Value: nil},
	}

	m := menu.New(
		menu.WithItems(menuItems),
		menu.WithTitle("Select an option"),
	)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
