package benchmark

import (
	"sync"
	"time"

	"github.com/dotvezz/cacheProxyBench/profiles"
	"github.com/dotvezz/cacheProxyBench/requests"
)

type Job struct {
	cancelOnce sync.Once
	cancelCh   chan struct{}
	doneCh     chan struct{}
	timeline   *Timeline
}

func (j *Job) Cancel() {
	j.cancelOnce.Do(func() {
		close(j.cancelCh)
	})
}

func (j *Job) Wait() {
	<-j.doneCh
}

func (j *Job) Done() bool {
	select {
	case <-j.doneCh:
		return true
	default:
		return false
	}
}

func Run(host string, p profiles.Profile, timeline *Timeline) *Job {
	job := &Job{
		cancelCh: make(chan struct{}),
		doneCh:   make(chan struct{}),
	}

	requestCount := p.PlannedRequestCount()
	startTime := time.Now()
	client := requests.NewClient()

	wg := &sync.WaitGroup{}
	wg.Add(requestCount)

	for _, t := range p.Tracks {
		_ = time.AfterFunc(t.From, func() {
			runTrack(wg, host, t, startTime, client, timeline, job.cancelCh)
		})
	}

	go func() {
		wg.Wait()
		close(job.doneCh)
	}()

	return job
}
