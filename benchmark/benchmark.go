package benchmark

import (
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync"
	"time"

	"github.com/dotvezz/cacheProxyBench/profiles"
	"github.com/dotvezz/cacheProxyBench/requests"
)

var (
	ErrNilResponse = errors.New("nil response")

	host = flag.String("host", "http://127.0.0.1:8080", "host:port")
)

func runTrack(wg *sync.WaitGroup, track profiles.Track, startTime time.Time, client *http.Client, timeline *Timeline, cancelCh <-chan struct{}) {
	stopAfterN := track.PlannedRequestCount()
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
			uri := *host + track.NextPath()
			go doRequest(wg, uri, startTime, client, timeline)
		case <-time.After(track.To - track.From):
			break
		case <-cancelCh:
			return
		}
	}
}

func doRequest(wg *sync.WaitGroup, uri string, startTime time.Time, client *http.Client, timeline *Timeline) {
	func() {
		req := requests.Get(http.MethodGet, uri)
		defer func() {
			wg.Done()
			requests.Put(req)
		}()
		result := Result{}

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

		result.Completed = time.Since(startTime)
		if resp != nil {
			if _, err = io.Copy(io.Discard, resp.Body); err != nil {
				panic(err)
			}

			result.Hit = isHit(resp)
			result.Collapsed = isCollapsed(resp)
		}

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
