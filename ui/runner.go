package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/profiles"
)

func NewRunner(profile profiles.Profile) *Runner {
	return &Runner{
		profile: profile,
	}
}

type Runner struct {
	profile profiles.Profile

	cancelCh chan struct{}
}

func (r Runner) Init() tea.Cmd {
	return nil
}

func (r Runner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

}

func (r Runner) View() tea.View {
	//TODO implement me
	panic("implement me")
}
