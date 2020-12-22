package grpc

import (
	pb "apigo/grpc/httprunner"
	"apigo/runner"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

var MaxRunner = 100

type HttpRunnerServer struct {
	pb.HttpRunnerServer
	runners []*runner.Runner
	runnerChan chan *runner.Runner
	currentRunner *runner.Runner

}

func NewHttpRunnerServer() *HttpRunnerServer {
	svr := &HttpRunnerServer{
		runners: make([]*runner.Runner, 0),
		currentRunner: nil,
		runnerChan: make(chan *runner.Runner, 100),
	}
	go svr.processRunners()
	return svr
}

func (h *HttpRunnerServer) Enqueue(ctx context.Context, config *pb.RunnerConfig) (*pb.RunnerResponse, error) {
	duration, err := time.ParseDuration(config.Duration)
	if err != nil {
		duration = time.Second * 10
	}
	reader := bufio.NewReader(strings.NewReader(config.Url.Headers + "\r\n"))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Fatal(err)
	}
	runnerConfig := runner.Runner{
		Config: runner.RunnerConfig{
			Duration:          duration,
			Workers:           int(config.Workers),
			NeedResponse:      config.NeedResponse,
			Request:           runner.APIRequest{
				Method:  config.Url.Method,
				URL:     config.Url.Url,
				Body:    config.Url.Body,
				Headers: http.Header(mimeHeader),
			},
			OutputCSVFilename: "",
			CountRequestSize:  false,
			CountResponseSize: false,
		},
		Metrics:           nil,
	}
	h.runners = append(h.runners, &runnerConfig)
	result := pb.RunnerResponse{
		Status:       0,
		Message:      "",
		RunnerConfig: &pb.RunnerConfig{
			RunnerId:     int32(len(h.runners)),
			Duration:     runnerConfig.Config.Duration.String(),
			Workers:      int32(runnerConfig.Config.Workers),
			NeedResponse: runnerConfig.Config.NeedResponse,
			Url:        &pb.Url{
				Url:     runnerConfig.Config.Request.URL,
				Method:  runnerConfig.Config.Request.Method,
				Body:    runnerConfig.Config.Request.Body,
				Headers: config.Url.Headers,
			},
			Status:       0,
		},
	}
	fmt.Printf("Runner Added %v\n", runnerConfig)
	h.runnerChan <- &runnerConfig
	return &result, nil
}

func (h *HttpRunnerServer) GetRunner(ctx context.Context, request *pb.IdRunnerRequest) (*pb.RunnerResponse, error) {
	return &pb.RunnerResponse{
		Status:       0,
		Message:      "OK",
		RunnerConfig: nil,
	}, nil
}

func (h *HttpRunnerServer) GetRunners(ctx context.Context, empty *emptypb.Empty) (*pb.RunnersResponse, error) {
	return &pb.RunnersResponse{
		Status:       0,
	}, nil
}

func (h *HttpRunnerServer) RemoveRunner(ctx context.Context, request *pb.IdRunnerRequest) (*pb.SimpleResponse, error) {
	return &pb.SimpleResponse{
		Status:       0,
		Message:      "OK",
	}, nil
}


func (h *HttpRunnerServer) CancelRunning(ctx context.Context, empty *emptypb.Empty) (*pb.SimpleResponse, error) {
	return &pb.SimpleResponse{
		Status:       0,
		Message:      "OK",
	}, nil
}

func (h *HttpRunnerServer) Listen(ctx context.Context, empty *emptypb.Empty) (*pb.RunnerResponse, error) {
	return &pb.RunnerResponse{
		Status:       0,
		Message:      "OK",
		RunnerConfig: nil,
	}, nil
}

func (h* HttpRunnerServer) processRunners() {
	for true {
		h.currentRunner = <- h.runnerChan
		runner.WorkerRun(*h.currentRunner)
	}
}
//func (h HttpRunnerServer) mustEmbedUnimplementedHttpRunnerServer() {
//	panic("implement me")
//}

