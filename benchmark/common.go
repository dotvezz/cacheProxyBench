package benchmark

import (
	"time"
)

type measurement int

func AverageTTFB(rs []Result) time.Duration {
	var sum time.Duration
	for _, r := range rs {
		sum += r.FirstByte - r.Sent
	}
	return sum / time.Duration(len(rs))
}

func AverageTotal(rs []Result) time.Duration {
	var sum time.Duration
	for _, r := range rs {
		sum += r.Completed - r.Sent
	}
	return sum / time.Duration(len(rs))
}

func ErrorRate(rs []Result) float64 {
	var sum float64
	for _, r := range rs {
		if r.Error != nil {
			sum++
		}
	}
	return sum / float64(len(rs))
}

func HitRate(rs []Result) float64 {
	var sum float64
	for _, r := range rs {
		if r.Error == nil {
			sum++
		}
	}
	return sum / float64(len(rs))
}

func CollapseRate(rs []Result) float64 {
	var sum float64
	for _, r := range rs {
		if r.Collapsed {
			sum++
		}
	}
	return sum / float64(len(rs))
}
