package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/dotvezz/cacheProxyBench/profiles"
)

func TestRun(t *testing.T) {
	type args struct {
		p profiles.Profile
	}
	tests := []struct {
		name         string
		args         args
		wantTimeline []Result
		wantErr      bool
	}{
		{
			name: "simple two requests",
			args: args{
				p: profiles.Profile{
					Name: "just two requests",
					Tracks: []profiles.Track{
						{
							Name: "two requests",
							From: 0,
							To:   time.Second,
							RPS:  2,
							Paths: []string{
								"/small_css",
							},
						},
					},
				},
			},
		},
		{
			name: "375_000 small requests at 75_000 requests per second",
			args: args{
				p: profiles.Profile{
					Tracks: []profiles.Track{
						{
							From: 0,
							To:   time.Second * 5,
							RPS:  75_000,
							Paths: []string{
								"/small_css",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeline := NewTimeline(tt.args.p.PlannedRequestCount())
			Run(tt.args.p, timeline).Wait()
			fmt.Println(timeline.Len())
		})
	}
}
