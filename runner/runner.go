package runner

import (
	"fmt"
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

type RunnerEventFunc func (runner *Runner)
type JobResponseFunc func (runner *Runner, response *http.Response)
// Worker - HTTP Job Worker
type Runner struct {
	// Start time
	Start         time.Time
	Config        RunnerConfig
	JobProvider   JobProviderFunc
	JobsCreated   int
	JobsProcessed int
	Metrics       []Metric
	OnJobStart    RunnerEventFunc
	OnJobComplete RunnerEventFunc
	OnJobRequest  JobProviderFunc
	OnJobResponse JobResponseFunc
}

// NewRunner creates a new runner
func NewRunner(seconds int, workers int) *Runner {
	return &Runner{
		Config: RunnerConfig{
			Duration:          time.Duration(seconds) * time.Second,
			Workers:           workers,
			NeedResponse:      false,
			Request:           APIRequest{},
			OutputCSVFilename: "",
			CountRequestSize:  false,
			CountResponseSize: false,
		},
		JobsCreated:   0,
		JobsProcessed: 0,
		JobProvider:   DefaultJobProvider,
	}
}

// DefaultJobProvider to provide sample job request
func DefaultJobProvider() Job {
	j := 1
	return NewAPIRequest("GET", fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", j), make(url.Values))
}
