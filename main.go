package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.NewMenu([]ui.MenuOption{
		{Key: "Option 1", Value: nil},
		{Key: "Option 2", Value: nil},
	})).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
