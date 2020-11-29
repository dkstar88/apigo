package services

import (
	"apigo/runner/Utils"
	"apigo/runner/models"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/TylerBrock/colorjson"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	config models.Runner
}
// jobProvider provides jobs and add to jobs channel
func (runner *Runner) collectFromJobProvider(ctx context.Context, jobs chan models.Job) {
	j := 1
	for {
		select {
		case <-ctx.Done():
			break
		case jobs <- runner.config.JobProvider():
			//fmt.Println("Pushing a new job")
			// jobs <- fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", j)
			j++
			runner.config.JobsCreated = j
		}
	}
}

func RunnerRun(config models.Runner) {
	runner := Runner{ config: config}
	runner.Run()
}
// Run from runner configuration
func (runner *Runner) Run() {

	var wg sync.WaitGroup
	//fmt.Println("Run >>>")
	workers := runner.config.Workers
	jobs := make(chan models.Job, workers)
	results := make(chan models.APIResponse, workers)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runner.collectFromJobProvider(ctx, jobs)
	for w := 1; w <= workers; w++ {
		go runner.worker(wg, ctx, w, jobs, results)
	}
	wg.Add(workers)
	runner.config.Start = time.Now()
	runner.OnJobStart()
	for {
		timeSince := time.Since(runner.config.Start)
		// fmt.Printf("Time since: %v\n", timeSince)
		if timeSince < runner.config.Duration {
			// Still running
			//fmt.Println("Still running")
		} else {
			//fmt.Println("Exiting")
			cancel()
			break
		}
		for a := 1; a <= len(results); a++ {
			runner.config.JobsProcessed++
			thisResult := <-results
			if runner.config.NeedResponse {
				var obj map[string]interface{}
				e := json.Unmarshal(thisResult.Body, &obj)
				if e != nil {
					log.Fatal("JSON Unmarshal failed")
				}
				// Make a custom formatter with indent set
				f := colorjson.NewFormatter()
				f.Indent = 4

				// Marshall the Colorized JSON
				s, _ := f.Marshal(obj)
				fmt.Println(string(s))
			}

		}
	}
	cancel()
	go func() {
		wg.Wait()
		close(results)
	}()

	runner.OnJobComplete()
	//fmt.Println("Run <<<")
}

// worker processes jobs channel and sends http request
func (runner *Runner) worker(waiter sync.WaitGroup, ctx context.Context, id int, jobs <-chan models.Job, results chan<- models.APIResponse) {
	for {
		j := runner.config.JobProvider()
		//fmt.Println("worker", id, "started  job", j)
		r := runner.MakeRequest(ctx, j)
		//fmt.Println("worker", id, "finished job", j)

		select {
		case <-ctx.Done():
			waiter.Done()
			break
		case results <- r:
		}
	}
}

var (
	// Command line flags.
	insecure        bool
	clientCertFile  string
)


// readClientCert - helper function to read client certificate
// from pem formatted file
func readClientCert(filename string) []tls.Certificate {
	if filename == "" {
		return nil
	}
	var (
		pkeyPem []byte
		certPem []byte
	)

	// read client certificate file (must include client private key and certificate)
	certFileBytes, err := ioutil.ReadFile(clientCertFile)
	if err != nil {
		log.Fatalf("failed to read client certificate file: %v", err)
	}

	for {
		block, rest := pem.Decode(certFileBytes)
		if block == nil {
			break
		}
		certFileBytes = rest

		if strings.HasSuffix(block.Type, "PRIVATE KEY") {
			pkeyPem = pem.EncodeToMemory(block)
		}
		if strings.HasSuffix(block.Type, "CERTIFICATE") {
			certPem = pem.EncodeToMemory(block)
		}
	}

	cert, err := tls.X509KeyPair(certPem, pkeyPem)
	if err != nil {
		log.Fatalf("unable to load client cert and key pair: %v", err)
	}
	return []tls.Certificate{cert}
}



// MakeRequest - initiate a request using HTTP
func (runner *Runner) MakeRequest(ctx context.Context, job models.Job) models.APIResponse {

	u, err := url.Parse(job.URL)
	if err != nil {
		log.Fatal(err)
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
		log.Fatalf("%v", err)
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
		WroteRequest: func(_ httptrace.WroteRequestInfo) { requestDone = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 30 * time.Second,
	}

	switch u.Scheme {
	case "https":
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: insecure,
			Certificates:       readClientCert(clientCertFile),
		}

		// Because we create a custom TLSClientConfig, we have to opt-in to HTTP/2.
		// See https://github.com/golang/go/issues/14275
		err = http2.ConfigureTransport(tr)
		if err != nil {
			log.Fatalf("failed to prepare transport for HTTP/2: %v", err)
		}
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
		log.Fatalf("failed to read response: %v", err)
	}

	responseDone := time.Now() // after read body
	if dnsStart.IsZero() {
		// we skipped DNS
		dnsStart = dnsDone
	}

	apiResponse := models.APIResponse{Headers: res.Header, Status: res.StatusCode}
	apiResponse.ContentType = res.Header.Get("Content-Type")
	dataReceived := 0
	if runner.config.CountResponseSize {
		dataReceived = Utils.CountResponseSize(res)
	}
	dataSent := 0
	if runner.config.CountRequestSize{
		dataSent = Utils.CountRequestSize(req)
	}
	if runner.config.NeedResponse {
		// read response body
		apiResponse.Body, _ = ioutil.ReadAll(res.Body)
		dataReceived = len(apiResponse.Body)
		// close response body
		e := res.Body.Close()
		if e != nil {
			log.Fatal("Response body close failed")
		}
	}

	metric := models.Metric{
		DataSent:       dataSent,
		DataReceived:   dataReceived,
		HTTPDNS:        dnsDone.Sub(dnsStart),
		HTTPConnecting: connStart.Sub(dnsDone),
		HTTPReceiving:  responseDone.Sub(responseStart),
		HTTPBlocked: 	dnsStart.Sub(requestStart),
		HTTPWaiting:    responseStart.Sub(connDone),
		HTTPSending: requestDone.Sub(connDone),
		HTTPTotal: responseDone.Sub(requestStart),
	}
	switch u.Scheme {
	case "https":
		metric.HTTPTls = tlsDone.Sub(tlsStart)
		metric.HTTPConnecting = connStart.Sub(dnsDone)
	case "http":
		metric.HTTPConnecting = connDone.Sub(dnsDone)
	}

	runner.config.Metrics = append(runner.config.Metrics, metric)
	return apiResponse
}

func (runner *Runner) OnJobStart() {
	// TODO: Move runner setup code here
}

func (runner *Runner) OnJobComplete() {
	Utils.ConsoleOutput(&runner.config)
	if len(runner.config.OutputCSVFilename) <= 0 {
		return
	}
	Utils.MetricsToCsv(runner.config.Metrics, runner.config.OutputCSVFilename)

}
