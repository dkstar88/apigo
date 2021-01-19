package runner

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var DefaultRunner = Runner{
	OnJobStart:    nil,
	OnJobComplete: nil,
	OnJobRequest:  nil,
	OnJobResponse: nil,
}

type Worker struct {
	runner *Runner
}

func WorkerRun(runner Runner) {
	worker := Worker{runner: &runner}
	worker.Run()
}

// Run from runner configuration
func (worker *Worker) Run() {
	var wg sync.WaitGroup
	var metricLock sync.Mutex

	//fmt.Println("Run >>>")
	workers := worker.runner.Config.Workers
	results := make(chan APIResponse, workers)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	worker.runner.Start = time.Now()
	worker.OnJobStart()
	for w := 1; w <= workers; w++ {
		go worker.worker(&wg, ctx, results, &metricLock)
	}
	wg.Add(workers)
	for {
		timeSince := time.Since(worker.runner.Start)
		// fmt.Printf("Time since: %v\n", timeSince)
		if timeSince < worker.runner.Config.Duration {
			// Still running
			//fmt.Println("Still running")
		} else {
			//fmt.Println("Exiting")
			cancel()
			break
		}
		for a := 1; a <= len(results); a++ {
			atomic.AddInt64(&worker.runner.JobsProcessed, 1)
			r := <-results
			if val, ok := worker.runner.StatusCodes[r.Status]; ok {
				worker.runner.StatusCodes[r.Status] = val + 1
			} else {
				worker.runner.StatusCodes[r.Status] = 1
			}
			if r.Status >= 200 && r.Status <= 299 {
				atomic.AddInt64(&worker.runner.JobsSuccessful, 1)
			} else {
				atomic.AddInt64(&worker.runner.JobsFailed, 1)
			}
		}
	}
	cancel()
	go func() {
		wg.Wait()
		close(results)
	}()

	worker.OnJobComplete()
}

// Min returns the smaller of x or y.
func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// worker processes jobs channel and sends http request
func (worker *Worker) worker(waiter *sync.WaitGroup, ctx context.Context, results chan<- APIResponse, metricLock *sync.Mutex) {
	defer waiter.Done()
	// Init with no error wait time
	errorWaitTime := time.Duration(0)
	var errorTimes int = 0
	for {
		if !worker.runner.Cancelled.IsZero() {
			break
		}
		if errorWaitTime > 0 {
			time.Sleep(time.Millisecond * 100)
			errorWaitTime -= time.Millisecond * 100
			if errorWaitTime > 0 {
				continue
			}
		}
		var j Job
		if worker.runner.OnJobRequest != nil {
			j = worker.runner.OnJobRequest(worker.runner)
		} else if DefaultRunner.OnJobRequest != nil {
			j = DefaultRunner.OnJobRequest(worker.runner)
		} else {
			j = DefaultJobRequest(worker.runner)
		}
		atomic.AddInt64(&worker.runner.JobsCreated, 1)
		//fmt.Printf("started job: %v\n", j)
		r, err := worker.MakeRequest(ctx, j, metricLock)
		//fmt.Println("worker", id, "finished job", j)
		if err != nil {
			atomic.AddInt64(&worker.runner.JobsFailed, 1)
			errorWaitTime = (100 * time.Millisecond) * time.Duration(min(errorTimes, 10))
		}
		select {
		case <-ctx.Done():

			break
		case results <- r:
		}
	}
}

// MakeRequest - initiate a request using HTTP
func (worker *Worker) MakeRequest(ctx context.Context, job Job, metricLock *sync.Mutex) (APIResponse, error) {

	u, err := url.Parse(job.URL)
	if err != nil {
		log.Fatalf("url.Parse: %v", err)
	}
	query := u.Query()
	for k, v := range job.Arguments {
		for _, i := range v {
			query.Add(k, i)
		}
	}
	requestStart := time.Now()
	// create a request object
	req, err := http.NewRequestWithContext(ctx, job.Method, u.String(), nil)
	if err != nil {
		log.Printf("Error: NewRequestWithContext: %v", err)
		return APIResponse{}, err
	}
	req.Header.Add("Content-Type", job.ContentType)

	var dnsStart, dnsDone, connStart, connDone, requestDone, responseStart, tlsStart, tlsDone time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart: func(_, _ string) {
			if dnsDone.IsZero() {
				// connecting to IP
				dnsDone = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			connStart = time.Now()
		},
		GotConn:              func(_ httptrace.GotConnInfo) { connDone = time.Now() },
		GotFirstResponseByte: func() { responseStart = time.Now() },
		TLSHandshakeStart:    func() { tlsStart = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { tlsDone = time.Now() },
		WroteRequest:         func(_ httptrace.WroteRequestInfo) { requestDone = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 30 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects, visit does that
			// manually if required.
			return http.ErrUseLastResponse
		},
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return APIResponse{}, err
	}

	responseDone := time.Now() // after read body
	if dnsStart.IsZero() {
		// we skipped DNS
		dnsStart = dnsDone
	}

	apiResponse := APIResponse{Headers: res.Header, Status: res.StatusCode}
	apiResponse.ContentType = res.Header.Get("Content-Type")
	dataReceived := 0
	if worker.runner.Config.CountResponseSize {
		dataReceived = CountResponseSize(res)
	}
	dataSent := 0
	if worker.runner.Config.CountRequestSize {
		dataSent = CountRequestSize(req)
	}
	if worker.runner.Config.NeedResponse {
		if worker.runner.OnJobResponse != nil {
			worker.runner.OnJobResponse(worker.runner, res)
		} else if DefaultRunner.OnJobResponse != nil {
			DefaultRunner.OnJobResponse(worker.runner, res)
		}

	}

	metric := Metric{
		DataSent:       dataSent,
		DataReceived:   dataReceived,
		HTTPDNS:        dnsDone.Sub(dnsStart),
		HTTPConnecting: connStart.Sub(dnsDone),
		HTTPReceiving:  responseDone.Sub(responseStart),
		HTTPBlocked:    dnsStart.Sub(requestStart),
		HTTPWaiting:    responseStart.Sub(connDone),
		HTTPSending:    requestDone.Sub(connDone),
		HTTPTotal:      responseDone.Sub(requestStart),
	}
	switch u.Scheme {
	case "https":
		metric.HTTPTls = tlsDone.Sub(tlsStart)
		metric.HTTPConnecting = connStart.Sub(dnsDone)
	case "http":
		metric.HTTPConnecting = connDone.Sub(dnsDone)
	}

	metricLock.Lock()
	worker.runner.Metrics = append(worker.runner.Metrics, metric)
	metricLock.Unlock()

	return apiResponse, nil
}

func (worker *Worker) OnJobStart() {
	// TODO: Move runner setup code here
	if worker.runner.OnJobStart != nil {
		worker.runner.OnJobStart(worker.runner)
	} else if DefaultRunner.OnJobComplete != nil {
		DefaultRunner.OnJobStart(worker.runner)
	}
}

func (worker *Worker) OnJobComplete() {

	if worker.runner.OnJobComplete != nil {
		worker.runner.OnJobComplete(worker.runner)
	} else if DefaultRunner.OnJobComplete != nil {
		DefaultRunner.OnJobComplete(worker.runner)
	}
}
