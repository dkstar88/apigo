package Utils

import (
	"apigo/runner/models"
	"reflect"
	"sort"
	"time"
)

type MetricStat struct {
	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	Min time.Duration
	Max time.Duration
	Avg time.Duration
}

func avg(durations []time.Duration) time.Duration {
	total := time.Duration(0)
	for _, duration := range durations {
		total = total + duration
	}
	return time.Duration(float64(total) / float64(len(durations)))
}

func GetStats(durations []time.Duration) MetricStat {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	result := MetricStat{
		P50: durations[len(durations)/2],
		P90: durations[int(float64(len(durations))*0.9)],
		P95: durations[int(float64(len(durations))*0.95)],
		Min: durations[0],
		Avg: avg(durations),
		Max: durations[len(durations)-1],
	}
	return result
}

func MetricsExtract(metrics []models.Metric, field string) []time.Duration {
	tmpMetric := models.Metric{}
	vt := reflect.TypeOf(tmpMetric)
	f, _ := vt.FieldByName(field)
	result := make([]time.Duration, len(metrics))
	for i, m := range metrics {
		result[i] = time.Duration(reflect.ValueOf(m).FieldByIndex(f.Index).Int())
	}
	return result
}

func MetricsToCsv(metrics []models.Metric, csvFilename string) {

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
