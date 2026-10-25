package benchmark

import (
	"fmt"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	type args struct {
		host string
		p    Profile
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
				host: "http://127.0.0.1:8005",
				p: Profile{
					Name: "just two requests",
					Tracks: []Track{
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
				host: "http://127.0.0.1:8005",
				p: Profile{
					Tracks: []Track{
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
			timeline := NewTimeline(PlannedRequestCount(tt.args.p.Tracks...))
			err := Run(tt.args.host, tt.args.p, timeline)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(timeline.Len())
		})
	}
}
