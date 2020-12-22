package utils

import (
	"apigo/runner"
	"math"
	"reflect"
	"sort"
	"time"
	"github.com/montanaflynn/stats"
)

type MetricStat struct {
	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration
	Min time.Duration
	Max time.Duration
	Avg time.Duration
	Median time.Duration
	StdDev time.Duration
}

func avg(durations []time.Duration) time.Duration {
	total := time.Duration(0)
	for _, duration := range durations {
		total = total + duration
	}
	return time.Duration(float64(total) / float64(len(durations)))
}

func getPercentile(sortedDuration []time.Duration, percentile uint) time.Duration {

	// Find the length of items in the slice
	length := len(sortedDuration)
	if length <= 0 {
		return 0
	}
	// Return the last item
	if percentile >= 100 {
		return sortedDuration[length-1]
	} else if percentile <= 0 {
		return sortedDuration[0]
	}

	// Find ordinal ranking
	or := int(math.Ceil(float64(length) * float64(percentile) / 100))

	// Return the item that is in the place of the ordinal rank
	if or == 0 {
		return sortedDuration[0]
	}
	return sortedDuration[or-1]
}

func GetStats(durations []time.Duration) MetricStat {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	float64Data := make(stats.Float64Data, len(durations))
	for i := range durations {
		float64Data[i] = float64(durations[i])
	}
	stddev, err := float64Data.StandardDeviation()
	if err != nil {
		stddev = 0
	}
	median, err := float64Data.Median()
	if err != nil {
		median = 0
	}
	result := MetricStat{
		P50: getPercentile(durations, 50),
		P90: getPercentile(durations, 90),
		P95: getPercentile(durations, 95),
		P99: getPercentile(durations, 99),
		Min: durations[0],
		Avg: avg(durations),
		Max: durations[len(durations)-1],
		StdDev: time.Duration(stddev),
		Median: time.Duration(median),
	}
	return result
}

func MetricsExtract(metrics []runner.Metric, field string) []time.Duration {
	tmpMetric := runner.Metric{}
	vt := reflect.TypeOf(tmpMetric)
	f, _ := vt.FieldByName(field)
	result := make([]time.Duration, len(metrics))
	for i, m := range metrics {
		result[i] = time.Duration(reflect.ValueOf(m).FieldByIndex(f.Index).Int())
	}
	return result
}


func GetMetricsStat(metrics []runner.Metric) map[string] MetricStat {

	fields := []string{
		"Blocked", "DNS", "Tls", "Connection",
		"Sending", "Receiving",
		"Waiting", "Total",
	}
	result := make(map[string] MetricStat)
	for _, f := range fields {
		result[f] = GetStats(MetricsExtract(metrics, f))
	}
	return result
}

func MetricsToCsv(metrics []runner.Metric, csvFilename string) {

	headers := []string{
		"DataSent", "DataReceived",
		"Blocked", "DNS", "Tls", "Connection",
		"Sending", "Receiving",
		"Waiting", "Total",
	}
	records := make([][]string, len(metrics))
	for i, m := range metrics {
		records[i] = []string{
			ByteCountIEC(m.DataSent),
			ByteCountIEC(m.DataReceived),
			DurationToString(m.HTTPBlocked),
			DurationToString(m.HTTPDNS),
			DurationToString(m.HTTPTls),
			DurationToString(m.HTTPConnecting),
			DurationToString(m.HTTPSending),
			DurationToString(m.HTTPWaiting),
			DurationToString(m.HTTPReceiving),
			DurationToString(m.HTTPTotal),
		}
	}
	WriteCSVFile(csvFilename, headers, records)
}
