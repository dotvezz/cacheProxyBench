package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/benchmark"
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

func menu() {
	menuItems := []ui.Item{
		{Name: "(R)un", HotKey: 'r', Value: profilesMenu},
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

func main() {
	target := flag.String("profile", "", "profile to use")
	flag.Parse()

	if len(os.Args) >= 2 && os.Args[1] == "menu" {
		menu()
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

	runner := ui.NewRunner(profile, ui.WithNext(func(p profiles.Profile, t *benchmark.Timeline) tea.Model {
		return ui.NewResults(p, t)
	}))

	if _, err := tea.NewProgram(runner).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
