package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"
	"github.com/dotvezz/cacheProxyBench/benchmark"
	"github.com/dotvezz/cacheProxyBench/profiles"
)

func NewRunner(profile profiles.Profile) *runner {
	return &runner{
		profile: profile,
	}
}

type tickMsg time.Time
type jobDoneMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func waitJobCmd(job *benchmark.Job) tea.Cmd {
	return func() tea.Msg {
		job.Wait()
		return jobDoneMsg{}
	}
}

type runner struct {
	profile  profiles.Profile
	timeline benchmark.Timeline
	done     bool
	job      *benchmark.Job
}

func (r *runner) Init() tea.Cmd {
	r.job = benchmark.Run(r.profile, &r.timeline)

	return tea.Batch(tickCmd(), waitJobCmd(r.job))
}

func (r *runner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		var cmd = tickCmd()
		if r.done {
			cmd = nil
		}
		return r, cmd
	case jobDoneMsg:
		r.done = true
	case tea.KeyPressMsg:
		switch msg.Code {
		case 'q':
			r.job.Cancel()
			return r, tea.Quit
		}

		switch msg.String() {
		case "ctrl+c":
			r.job.Cancel()
			return r, tea.Quit
		}
	}

	return r, nil
}

func (r *runner) chart(metric, aggType string) string {
	tslc := timeserieslinechart.New(41, 10)
	for i, v := range []float64{0, 4, 8, 10, 8, 4, 0, -4, -8, -10, -8, -4, 0} {
		date := time.Now().Add(time.Hour * time.Duration(24*i))
		tslc.Push(timeserieslinechart.TimePoint{date, v})
	}

	return tslc.View()
}

func (r *runner) View() tea.View {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s (target name %s)\n\n", r.profile.Name, r.profile.TargetName))

	total := r.timeline.Len()

	sb.WriteString(fmt.Sprintf("Requests:  %d / %d\n", total, r.profile.PlannedRequestCount()))
	sb.WriteString(fmt.Sprintf("Misses:      %d\n", r.timeline.Misses()))
	sb.WriteString(fmt.Sprintf("Hits:      %d\n", r.timeline.Hits()))
	sb.WriteString(fmt.Sprintf("Collapsed: %d\n", r.timeline.Collapsed()))
	sb.WriteString(fmt.Sprintf("Errors:    %d\n", r.timeline.Errors()))

	if total > 0 {
		sb.WriteString(fmt.Sprintf("\nTime To First Byte: %s\n", r.timeline.AvgFirstByte().Round(time.Millisecond)))
		sb.WriteString(fmt.Sprintf("Avg Completed:  %s\n", r.timeline.AvgCompleted().Round(time.Millisecond)))
	}

	sb.WriteString(fmt.Sprintf("Finished %d of %d requests.\n\n", r.timeline.Len(), r.profile.PlannedRequestCount()))

	return tea.NewView(sb.String())
}
