package runner

import (
	"net/http"
	"net/url"
	"time"
)

// APIRequest api request containing necessary data to initiate a request
type APIRequest struct {
	Method      string
	URL         string
	Arguments   url.Values
	Body        string
	ContentType string
	Headers     http.Header
}

// APIResponse holds response from an api request
type APIResponse struct {
	Body        []byte
	Status      int
	ContentType string
	Headers     http.Header
}

// NewAPIRequest creates a new APIRequest
func NewAPIRequest(method string, url string, arguments url.Values) APIRequest {
	req := APIRequest{
		Method:      method,
		URL:         url,
		Body:        "",
		ContentType: "application/json; charset=UTF-8",
		Arguments:   arguments,
		Headers:     make(map[string][]string),
	}
	return req
}

type JobProviderFunc func(runner *Runner) APIRequest
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

type RunnerConfig struct {
	// End time
	Duration time.Duration
	// How many concurrent workers
	Workers int
	// Keep response
	NeedResponse      bool
	Request           APIRequest
	OutputCSVFilename string
	CountRequestSize  bool
	CountResponseSize bool
}

type EventFunc func(runner *Runner)
type JobResponseFunc func(runner *Runner, response *http.Response)

// Worker - HTTP Job Worker
type Runner struct {
	// Start time
	Start          time.Time
	Cancelled      time.Time
	Config         RunnerConfig
	JobsCreated    int
	JobsProcessed  int
	JobsFailed     int // Network Failed, or Status returns not 200-299
	JobsSuccessful int // HTTP Status return 200-299
	StatusCodes    map[int]uint32
	Metrics        []Metric
	OnJobStart     EventFunc
	OnJobComplete  EventFunc
	OnJobRequest   JobProviderFunc
	OnJobResponse  JobResponseFunc
}

// NewRunner creates a new runner
func NewRunner(runnerConfig RunnerConfig) *Runner {
	return &Runner{
		Config:         runnerConfig,
		Cancelled:      time.Time{},
		JobsCreated:    0,
		JobsProcessed:  0,
		JobsFailed:     0,
		JobsSuccessful: 0,
		StatusCodes:    make(map[int]uint32),
		OnJobRequest:   DefaultJobRequest,
	}
}

func DefaultJobRequest(runner *Runner) APIRequest {
	return runner.Config.Request
}

func (r *Runner) GetProgress() float64 {
	return time.Now().Sub(r.Start).Seconds() / r.Config.Duration.Seconds()
}
