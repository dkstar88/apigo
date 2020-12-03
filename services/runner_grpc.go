package services

import (
	"apigo/runner/models"
	pb "apigo/runner/services/httprunner"
	"bufio"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

var MaxRunner = 100

type HttpRunnerServer struct {
	pb.UnimplementedHttpRunnerServer
	runners []models.Runner
	currentRunner *models.Runner

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
	runnerConfig := models.Runner{
		Duration:          duration,
		Workers:           int(config.Workers),
		NeedResponse:      config.NeedResponse,
		Request:           models.APIRequest{
			Method:  config.Url.Method,
			URL:     config.Url.Url,
			Body:    config.Url.Body,
			Headers: http.Header(mimeHeader),
		},
		Metrics:           nil,
		CountRequestSize:  false,
		CountResponseSize: false,
	}
	h.runners = append(h.runners, runnerConfig)
	result := pb.RunnerResponse{
		Status:       0,
		Message:      "",
		RunnerConfig: &pb.RunnerConfig{
			RunnerId:     int32(len(h.runners)),
			Duration:     runnerConfig.Duration.String(),
			Workers:      int32(runnerConfig.Workers),
			NeedResponse: runnerConfig.NeedResponse,
			Url:        &pb.Url{
				Url:     runnerConfig.Request.URL,
				Method:  runnerConfig.Request.URL,
				Body:    runnerConfig.Request.URL,
				Headers: config.Url.Headers,
			},
			Status:       0,
		},
	}
	return &result, nil
}

func (h *HttpRunnerServer) GetRunner(ctx context.Context, request *pb.IdRunnerRequest) (*pb.RunnerResponse, error) {
	panic("implement me")
}

func (h *HttpRunnerServer) GetRunners(ctx context.Context, empty *emptypb.Empty) (*pb.RunnersResponse, error) {
	panic("implement me")
}

func (h *HttpRunnerServer) RemoveRunner(ctx context.Context, request *pb.IdRunnerRequest) (*pb.SimpleResponse, error) {
	panic("implement me")
}

func (h *HttpRunnerServer) CancelRunning(ctx context.Context, empty *emptypb.Empty) (*pb.SimpleResponse, error) {
	panic("implement me")
}

func (h HttpRunnerServer) Listen(ctx context.Context, empty *emptypb.Empty) (*pb.RunnerResponse, error) {
	panic("implement me")
}

func (h HttpRunnerServer) mustEmbedUnimplementedHttpRunnerServer() {
	panic("implement me")
}

