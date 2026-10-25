package benchmark

import (
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/dotvezz/cacheProxyBench/requests"
)

func Run(host string, p Profile, timeline *Timeline) (err error) {
	requestCount := PlannedRequestCount(p.Tracks...)
	startTime := time.Now()
	client := requests.NewClient()

	wg := &sync.WaitGroup{}
	wg.Add(requestCount)

	for _, t := range p.Tracks {
		_ = time.AfterFunc(t.From, func() {
			runTrack(wg, host, t, startTime, client, timeline)
		})
	}

	wg.Wait()

	return err
}

func PlannedRequestCount(tracks ...Track) (count int) {
	for _, t := range tracks {
		count += int((t.To - t.From).Seconds() * float64(t.RPS))
	}

	return count
}

func runTrack(wg *sync.WaitGroup, host string, track Track, startTime time.Time, client *http.Client, timeline *Timeline) {
	stopAfterN := PlannedRequestCount(track)
	countStarted := 0
	ticker := time.NewTicker(time.Second / time.Duration(track.RPS))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if countStarted >= stopAfterN {
				break
			}
			countStarted++
			uri := host + track.NextPath()
			go doRequest(wg, uri, startTime, client, timeline)
		case <-time.After(track.To - track.From):
			break
			// TODO: Cancel
		}
	}
}

func doRequest(wg *sync.WaitGroup, uri string, startTime time.Time, client *http.Client, timeline *Timeline) {
	func() {
		defer func() {
			wg.Done()
		}()
		result := Result{}
		req := requests.Get(http.MethodGet, uri)

		trace := &httptrace.ClientTrace{
			WroteRequest: func(info httptrace.WroteRequestInfo) {
				result.Sent = time.Since(startTime)
			},
			GotFirstResponseByte: func() {
				result.FirstByte = time.Since(startTime)
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		resp, err := client.Do(req)
		if err != nil {
			// TODO: handle this case better
			panic(err)
		}
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			panic(err)
		}
		result.Completed = time.Since(startTime)

		timeline.Append(result)
	}()
}
