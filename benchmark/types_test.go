package benchmark

import (
	"reflect"
	"slices"
	"sync"
	"testing"
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
			t := &Track{
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
