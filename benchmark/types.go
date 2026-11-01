package benchmark

import (
	"sync"
	"time"
)

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
	Hit       bool
	Collapsed bool
	Error     error
}
