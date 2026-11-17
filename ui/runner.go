package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/dotvezz/cacheProxyBench/benchmark"
	"github.com/dotvezz/cacheProxyBench/profiles"
)

type RunnerOption func(*Runner)

func WithNext(fn func(profiles.Profile, *benchmark.Timeline) tea.Model) RunnerOption {
	return func(r *Runner) {
		r.next = fn
	}
}

func NewRunner(profile profiles.Profile, opts ...RunnerOption) *Runner {
	r := &Runner{
		profile: profile,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
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

type Runner struct {
	profile  profiles.Profile
	timeline benchmark.Timeline

	job  *benchmark.Job
	next func(profiles.Profile, *benchmark.Timeline) tea.Model
}

func (r *Runner) Init() tea.Cmd {
	r.job = benchmark.Run(r.profile, &r.timeline)

	return tea.Batch(tickCmd(), waitJobCmd(r.job))
}

func (r *Runner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		return r, tickCmd()
	case jobDoneMsg:
		if r.next != nil {
			next := r.next(r.profile, &r.timeline)
			return next, next.Init()
		}
		return r, tea.Quit
	}

	return r, nil
}

func (r *Runner) View() tea.View {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Running: %s (target name %s)\n\n", r.profile.Name, r.profile.TargetName))
	sb.WriteString(fmt.Sprintf("Finished: %d of %d requests)\n\n", r.timeline.Len(), r.profile.PlannedRequestCount()))

	return tea.NewView(sb.String())
}
