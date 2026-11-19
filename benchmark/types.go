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

func (t *Timeline) count(pred func(Result) bool) int {
	if t == nil {
		return 0
	}
	n := 0
	for _, r := range t.Results() {
		if pred(r) {
			n++
		}
	}
	return n
}

// Hits returns the number of cache-hit results in the timeline.
func (t *Timeline) Hits() int {
	return t.count(func(r Result) bool { return r.Hit })
}

// Misses returns the number of cache-hit results in the timeline.
func (t *Timeline) Misses() int {
	return t.count(func(r Result) bool { return !r.Hit })
}

// Collapsed returns the number of collapsed results in the timeline.
func (t *Timeline) Collapsed() int {
	return t.count(func(r Result) bool { return r.Collapsed })
}

// Errors returns the number of results with a non-nil error in the timeline.
func (t *Timeline) Errors() int {
	return t.count(func(r Result) bool { return r.Error != nil })
}

func (t *Timeline) avgDuration(extract func(Result) time.Duration) time.Duration {
	if t == nil {
		return 0
	}
	results := t.Results()
	if len(results) == 0 {
		return 0
	}
	var sum time.Duration
	for _, r := range results {
		sum += extract(r)
	}
	return sum / time.Duration(len(results))
}

// AvgFirstByte returns the average time to first byte across all results.
func (t *Timeline) AvgFirstByte() time.Duration {
	return t.avgDuration(func(r Result) time.Duration {
		return r.FirstByte - r.Sent
	})
}

// AvgCompleted returns the average completion duration across all results.
func (t *Timeline) AvgCompleted() time.Duration {
	return t.avgDuration(func(r Result) time.Duration { return r.Completed - r.Sent })
}

type Result struct {
	Sent      time.Duration
	FirstByte time.Duration
	Completed time.Duration
	Hit       bool
	Collapsed bool
	Error     error
}
