package benchmark

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"time"

	"github.com/dotvezz/cacheProxyBench/requests"
)

var (
	ErrNilResponse = errors.New("nil response")
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
			result.Error = err
		} else if resp == nil {
			result.Error = ErrNilResponse
		}
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			panic(err)
		}
		result.Completed = time.Since(startTime)
		result.Hit = isHit(resp)
		result.Collapsed = isCollapsed(resp)

		timeline.Append(result)
	}()
}

func isHit(resp *http.Response) bool {
	cs := resp.Header.Get("Cache-Status")
	if cs != "" {
		csParts := strings.Split(cs, ";")
		for i := range csParts {
			if strings.TrimSpace(csParts[i]) == "hit" {
				return true
			}
		}
	}

	// TODO: Other response headers as needed
	return false
}

func isCollapsed(resp *http.Response) bool {
	cs := resp.Header.Get("Cache-Status")
	if cs != "" {
		csParts := strings.Split(cs, ";")
		for i := range csParts {
			if strings.TrimSpace(csParts[i]) == "collapsed" {
				return true
			}
		}
	}

	// TODO: Other response headers as needed
	return false
}
