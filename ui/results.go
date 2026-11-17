package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/benchmark"
	"github.com/dotvezz/cacheProxyBench/profiles"
)

func NewResults(profile profiles.Profile, timeline *benchmark.Timeline) *Results {
	return &Results{
		profile:  profile,
		timeline: timeline,
	}
}

type Results struct {
	profile  profiles.Profile
	timeline *benchmark.Timeline
}

func (r *Results) Init() tea.Cmd {
	return nil
}

func (r *Results) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'q':
			return r, tea.Quit
		}
		switch msg.String() {
		case "ctrl+c":
			return r, tea.Quit
		}
	}

	return r, nil
}

func (r *Results) View() tea.View {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Results: %s (target: %s)\n\n", r.profile.Name, r.profile.TargetName))

	total := 0
	if r.timeline != nil {
		total = r.timeline.Len()
	}

	sb.WriteString(fmt.Sprintf("Requests:  %d / %d\n", total, r.profile.PlannedRequestCount()))
	sb.WriteString(fmt.Sprintf("Hits:      %d\n", r.timeline.Hits()))
	sb.WriteString(fmt.Sprintf("Collapsed: %d\n", r.timeline.Collapsed()))
	sb.WriteString(fmt.Sprintf("Errors:    %d\n", r.timeline.Errors()))

	if total > 0 {
		sb.WriteString(fmt.Sprintf("\nAvg First Byte: %s\n", r.timeline.AvgFirstByte().Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("Avg Completed:  %s\n", r.timeline.AvgCompleted().Round(time.Millisecond)))
	}

	sb.WriteString("\nPress Q to quit\n")

	return tea.NewView(sb.String())
}
