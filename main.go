package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/profiles"
	"github.com/dotvezz/cacheProxyBench/ui"
	"github.com/dotvezz/cacheProxyBench/ui/menu"
)

func profilesMenu(m tea.Model) tea.Model {
	ps := profiles.GetAll()

	items := make([]menu.Item, len(ps))
	for i, p := range ps {
		items[i] = menu.Item{Name: p.Name, Value: func(parent tea.Model) tea.Model {
			runner := ui.NewRunner(p)
			return runner
		}}
	}

	return menu.New(menu.WithParent(m), menu.WithTitle("Select a profile"), menu.WithItems(items))
}

func showMenu() {
	menuItems := []menu.Item{
		{Name: "(R)un", HotKey: 'r', Value: profilesMenu},
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

func main() {
	target := flag.String("profile", "", "profile to use")
	flag.Parse()

	if len(os.Args) >= 2 && os.Args[1] == "menu" {
		showMenu()
		return
	}

	var (
		profile profiles.Profile
		err     error
	)

	if target != nil && *target != "" {
		profile, err = profiles.GetTarget(*target)
	} else {
		profile = profiles.GetAll()[0]
	}

	if err != nil {
		fmt.Println("Error getting profile:", err)
	}

	runner := ui.NewRunner(profile)

	if _, err := tea.NewProgram(runner).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
