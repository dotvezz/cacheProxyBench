package benchmark

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/dotvezz/cacheProxyBench/profiles"
)

func TestTrack_NextPath(t1 *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		repeatTimes int
		want        []string
	}{
		{
			name:        "empty",
			paths:       []string{},
			repeatTimes: 1,
			want:        []string{""},
		},
		{
			name:        "nil",
			paths:       nil,
			repeatTimes: 1,
			want:        []string{""},
		},
		{
			name:        "no overflow",
			paths:       []string{"a", "b", "c"},
			repeatTimes: 2,
			want:        []string{"a", "b"},
		},
		{
			name:        "overflow",
			paths:       []string{"a", "b", "c"},
			repeatTimes: 4,
			want:        []string{"a", "b", "c", "a"},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &profiles.Track{
				Paths: tt.paths,
			}

			got := make([]string, 0)
			for range tt.repeatTimes {
				got = append(got, t.NextPath())
			}

			if !slices.Equal(tt.want, got) {
				t1.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeline_Results(t1 *testing.T) {
	type fields struct {
		results []Result
		cursor  int
		mutex   *sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   []Result
	}{
		{
			name: "empty",
			fields: fields{
				results: []Result{},
			},
			want: []Result{},
		},
		{
			name: "1 result in 1-len slice",
			fields: fields{
				results: []Result{{}},
				cursor:  1,
			},
			want: []Result{{}},
		},
		{
			name: "1 result in 2-len slice",
			fields: fields{
				results: []Result{{}, {}},
				cursor:  1,
			},
			want: []Result{{}},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Timeline{
				results: tt.fields.results,
				cursor:  tt.fields.cursor,
				mutex:   tt.fields.mutex,
			}
			if got := t.Results(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Results() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeline_Aggregations(t1 *testing.T) {
	var nilTimeline *Timeline
	if nilTimeline.Hits() != 0 || nilTimeline.Collapsed() != 0 || nilTimeline.Errors() != 0 ||
		nilTimeline.AvgFirstByte() != 0 || nilTimeline.AvgCompleted() != 0 {
		t1.Fatalf("expected all aggregations on nil Timeline to be 0")
	}

	empty := NewTimeline(2)
	if empty.Hits() != 0 || empty.Collapsed() != 0 || empty.Errors() != 0 ||
		empty.AvgFirstByte() != 0 || empty.AvgCompleted() != 0 {
		t1.Fatalf("expected all aggregations on empty Timeline to be 0")
	}

	tl := NewTimeline(4)
	tl.Append(Result{Hit: true, FirstByte: 10 * time.Millisecond, Completed: 20 * time.Millisecond})
	tl.Append(Result{Hit: true, Collapsed: true, FirstByte: 20 * time.Millisecond, Completed: 40 * time.Millisecond})
	tl.Append(Result{Error: errors.New("boom"), FirstByte: 30 * time.Millisecond, Completed: 60 * time.Millisecond})

	if got := tl.Hits(); got != 2 {
		t1.Errorf("Hits() = %d, want 2", got)
	}
	if got := tl.Collapsed(); got != 1 {
		t1.Errorf("Collapsed() = %d, want 1", got)
	}
	if got := tl.Errors(); got != 1 {
		t1.Errorf("Errors() = %d, want 1", got)
	}
	if got := tl.AvgFirstByte(); got != 20*time.Millisecond {
		t1.Errorf("AvgFirstByte() = %v, want 20ms", got)
	}
	if got := tl.AvgCompleted(); got != 40*time.Millisecond {
		t1.Errorf("AvgCompleted() = %v, want 40ms", got)
	}
}
