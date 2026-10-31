package output

import (
	"time"

	"github.com/dotvezz/cacheProxyBench/benchmark"
)

type measurement int

func AverageTTFB(rs []benchmark.Result) time.Duration {
	var sum time.Duration
	for _, r := range rs {
		sum += r.FirstByte - r.Sent
	}
	return sum / time.Duration(len(rs))
}

func AverageTotal(rs []benchmark.Result) time.Duration {
	var sum time.Duration
	for _, r := range rs {
		sum += r.Completed - r.Sent
	}
	return sum / time.Duration(len(rs))
}

func ErrorRate(rs []benchmark.Result) float64 {
	var sum float64
	for _, r := range rs {
		if r.Error != nil {
			sum++
		}
	}
	return sum / float64(len(rs))
}

func HitRate(rs []benchmark.Result) float64 {
	var sum float64
	for _, r := range rs {
		if r.Error == nil {
			sum++
		}
	}
	return sum / float64(len(rs))
}

func CollapseRate(rs []benchmark.Result) float64 {
	var sum float64
	for _, r := range rs {
		if r.Collapsed {
			sum++
		}
	}
	return sum / float64(len(rs))
}
