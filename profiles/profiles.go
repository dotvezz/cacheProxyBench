package profiles

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

	plannedRequestCount int
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

func (t *Track) PlannedRequestCount() (count int) {
	if t.plannedRequestCount > 0 {
		return t.plannedRequestCount
	}
	t.plannedRequestCount = int((t.To - t.From).Seconds() * float64(t.RPS))
	return t.plannedRequestCount
}

type Profile struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description,omitempty"`
	AllowError  bool    `yaml:"allow_error,omitempty"`
	Tracks      []Track `yaml:"tracks"`

	plannedRequestCount int
}

func (p *Profile) PlannedRequestCount() int {
	if p.plannedRequestCount > 0 {
		return p.plannedRequestCount
	}

	for _, track := range p.Tracks {
		p.plannedRequestCount += track.PlannedRequestCount()
	}

	return p.plannedRequestCount
}

func GetAll() []Profile {
	return GetIncluded()
}
