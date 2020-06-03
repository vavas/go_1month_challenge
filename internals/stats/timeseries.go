package stats

import (
	"math"
	"sort"
	"sync"
	"time"
)

type timeseriesItem struct {
	value     int64
	timestamp int64
}

// TimeSeries structure, it contains a list of integers and it provides Avg, Min, Max, Percentile, etc.
type TimeSeries struct {
	sync.Mutex
	duration  time.Duration
	lastPrune int64
	data      []timeseriesItem
}

// Int64Slice provides a sorting feature to int64 slice. It's an implementation of sort.Interface.
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// NewTimeSeries creates a new TimeSeries.
func NewTimeSeries(duration time.Duration) *TimeSeries {
	return &TimeSeries{duration: duration, data: []timeseriesItem{}}
}

func (ts *TimeSeries) prune() {
	ts.Lock()
	defer ts.Unlock()

	now := time.Now().Unix()
	if ts.lastPrune == now {
		return
	}

	ts.lastPrune = now

	cutOffIdx := 0
	cutOff := time.Now().Add(-ts.duration).UnixNano()

	for idx, item := range ts.data {
		if item.timestamp > cutOff {
			cutOffIdx = idx
			break
		}
	}

	ts.data = ts.data[cutOffIdx:]
}

func (ts *TimeSeries) appendValue(value int64) {
	ts.Lock()
	defer ts.Unlock()

	ts.data = append(ts.data, timeseriesItem{value, time.Now().UnixNano()})
}

// Append add a new value to the list.
func (ts *TimeSeries) Append(value int64) {
	ts.appendValue(value)
	ts.prune()
}

// Percentile returns a percentile of the list.
func (ts *TimeSeries) Percentile(percent float64) int64 {
	values := []int64{}
	for _, item := range ts.data {
		values = append(values, item.value)
	}

	sort.Sort(Int64Slice(values))

	idx := int(math.Floor((percent * float64(len(values)-1)) + 0.5))

	return values[idx]
}

// Avg returns the average of the list.
func (ts *TimeSeries) Avg() float64 {
	count := len(ts.data)

	if count == 0 {
		return 0
	}

	var sum float64
	for _, item := range ts.data {
		sum = sum + float64(item.value)
	}

	return sum / float64(count)
}

// Min returns the min value of the list.
func (ts *TimeSeries) Min() int64 {
	if len(ts.data) == 0 {
		return 0
	}

	min := ts.data[0].value

	for _, item := range ts.data {
		if item.value < min {
			min = item.value
		}
	}

	return min
}

// Max returns the max value of the list.
func (ts *TimeSeries) Max() int64 {
	if len(ts.data) == 0 {
		return 0
	}

	max := ts.data[0].value

	for _, item := range ts.data {
		if item.value > max {
			max = item.value
		}
	}

	return max
}

// CountPerSecond returns the ratio of the length of the list by the duration in second.
func (ts *TimeSeries) CountPerSecond() float64 {
	if ts.duration == 0 {
		return 0
	}

	return float64(len(ts.data)) / ts.duration.Seconds()
}
