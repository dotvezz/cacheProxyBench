package benchmark

import (
	"sync"
	"time"
)

type Track struct {
	Name  string        `yaml:"name,omitempty"`
	Paths []string      `yaml:"paths"`
	RPS   int           `yaml:"rps,omitempty"`
	From  time.Duration `yaml:"from,omitempty"`
	To    time.Duration `yaml:"to,omitempty"`

	pathI     int
	pathMutex *sync.Mutex
}

func (t *Track) NextPath() string {
	if t.pathMutex == nil {
		t.pathMutex = &sync.Mutex{}
	}
	if len(t.Paths) == 0 {
		return ""
	}
	t.pathMutex.Lock()
	defer t.pathMutex.Unlock()
	i := (t.pathI) % len(t.Paths)
	t.pathI++

	return t.Paths[i]
}

type Profile struct {
	Name       string
	AllowError bool    `yaml:"allow_error,omitempty"`
	Tracks     []Track `yaml:"tracks"`
}

func NewTimeline(length int) *Timeline {
	t := Timeline{}
	t.results = make([]Result, length)
	t.mutex = &sync.Mutex{}
	return &t
}

type Timeline struct {
	results []Result
	cursor  int
	mutex   *sync.Mutex
}

func (t *Timeline) Len() int {
	return t.cursor
}

func (t *Timeline) Append(result Result) {
	if t.mutex == nil {
		t.mutex = &sync.Mutex{}
	}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if len(t.results) <= t.cursor {
		t.results = append(t.results, result)
	} else {
		t.results[t.cursor] = result
	}
	t.cursor++
}

func (t *Timeline) Results() []Result {
	return t.results[:t.cursor]
}

type Result struct {
	Sent      time.Duration
	FirstByte time.Duration
	Completed time.Duration
	Track     int
	Hit       bool
	Collapsed bool
	Error     error
}
