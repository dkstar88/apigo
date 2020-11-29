package models

import (
	"fmt"
	"net/url"
	"time"
)

type JobProviderFunc func() APIRequest
type Job = APIRequest

// Metric tracks each http request timing metrics
type Metric struct {
	DataSent       int
	DataReceived   int
	HTTPDNS        time.Duration
	HTTPBlocked    time.Duration
	HTTPConnecting time.Duration
	HTTPDuration   time.Duration
	HTTPReceiving  time.Duration
	HTTPSending    time.Duration
	HTTPTls        time.Duration
	HTTPWaiting    time.Duration
	HTTPTotal      time.Duration
}

// Runner - HTTP Job Runner
type Runner struct {
	// Start time
	Start time.Time
	// End time
	Duration time.Duration
	// How many concurrent workers
	Workers int
	// Keep response
	NeedResponse bool
	// stages
	// stages []Stage
	// Logger
	// Job Provider
	JobProvider       JobProviderFunc
	JobsCreated       int
	JobsProcessed     int
	Metrics           []Metric
	OutputCSVFilename string
	CountRequestSize  bool
	CountResponseSize bool
}

// NewRunner creates a new runner
func NewRunner(seconds int, workers int) *Runner {
	return &Runner{
		Duration:          time.Duration(seconds) * time.Second,
		Workers:           workers,
		JobsCreated:       0,
		JobsProcessed:     0,
		JobProvider:       DefaultJobProvider,
		CountRequestSize:  false,
		CountResponseSize: false,
	}
}

// DefaultJobProvider to provide sample job request
func DefaultJobProvider() Job {
	j := 1
	return NewAPIRequest("GET", fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", j), make(url.Values))
}
