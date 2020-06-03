package stats

import "time"

type statsStruct struct {
	RPS             float64 `json:"rps"`               // requests per second total averaged over the past 10 seconds
	Count           int64   `json:"count"`             // request count since process start
	Count400        int64   `json:"count_400"`         // HTTP 40x response count since process start
	Count500        int64   `json:"count_500"`         // HTTP 50x response count since process start
	ResponseTime95  int64   `json:"response_time_95"`  // 95th percentile of response times for upstream requests in the past 10sec
	ResponseTimeAvg float64 `json:"response_time_avg"` // average of response times for all upstream requests in the past 10 sec
	ResponseTimeMin int64   `json:"response_time_min"` // the smallest response time in the past 10 sec
	ResponseTimeMax int64   `json:"response_time_max"` // the largest response time in the past 10 sec
}

var statsData = statsStruct{}
var respTimes = NewTimeSeries(10 * time.Second)

// UpdateStats updates the current request/response statistics.
func UpdateStats(respTime int64, status400 bool, status500 bool) {
	respTimes.Append(respTime)

	statsData.Count = statsData.Count + 1

	if status400 {
		statsData.Count400 = statsData.Count400 + 1
	}

	if status500 {
		statsData.Count500 = statsData.Count500 + 1
	}

	statsData.RPS = respTimes.CountPerSecond()
	statsData.ResponseTime95 = respTimes.Percentile(0.95)
	statsData.ResponseTimeAvg = respTimes.Avg()
	statsData.ResponseTimeMin = respTimes.Min()
	statsData.ResponseTimeMax = respTimes.Max()
}
