package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/ui/menu"
)

func main() {

	menuItems := []menu.Item{
		{Name: "(R)un", HotKey: 'r', Value: func(m tea.Model) tea.Model { return menu.New(menu.WithParent(m)) }},
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
