package grpc

import (
	pb "apigo/grpc/httprunner"
	"apigo/runner"
	"apigo/utils"
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

type HttpRunnerServer struct {
	pb.HttpRunnerServer
	runners       []*pb.Runner
	runnerChan    chan *pb.Runner
	currentRunner *pb.Runner
	runnerPairMap map[string]runnerPair
}

type runnerPair struct {
	pbRunner *pb.Runner
	runner   *runner.Runner
}

func NewHttpRunnerServer() *HttpRunnerServer {
	svr := &HttpRunnerServer{
		runners:       make([]*pb.Runner, 0),
		currentRunner: nil,
		runnerChan:    make(chan *pb.Runner, 100),
		runnerPairMap: make(map[string]runnerPair),
	}
	go svr.processRunners()
	return svr
}

func (h *HttpRunnerServer) Enqueue(_ context.Context, config *pb.RunnerConfig) (*pb.RunnerResponse, error) {

	_runner := pb.Runner{
		RunnerId:  uuid.New().String(),
		Config:    config,
		Stats:     nil,
		StartTime: nil,
		Status:    pb.Status_UNKNOWN,
		Progress:  0,
	}
	result := pb.RunnerResponse{
		Status:  0,
		Message: "",
		Runner:  &_runner,
	}
	h.runners = append(h.runners, &_runner)
	fmt.Printf("Runner Added %s: Test [%s] with %d workers for %s\n",
		_runner.RunnerId, _runner.Config.Url, _runner.Config.Workers, _runner.Config.Duration)
	h.runnerChan <- &_runner
	return &result, nil
}

func (h *HttpRunnerServer) GetRunner(_ context.Context, request *pb.IdRunnerRequest) (*pb.RunnerResponse, error) {
	pair, ok := h.runnerPairMap[request.RunnerId]
	if !ok {
		log.Printf("%s Runner ID not found", request.RunnerId)
		return &pb.RunnerResponse{
			Status:  404,
			Message: "Not Found",
		}, fmt.Errorf("runner Id not found: %s", request.RunnerId)
	}
	if pair.pbRunner == nil {
		return &pb.RunnerResponse{
			Status:  0,
			Message: "OK",
		}, nil
	}
	return &pb.RunnerResponse{
		Status:  0,
		Message: "OK",
		Runner:  pair.pbRunner,
	}, nil
}

func (h *HttpRunnerServer) GetRunners(_ context.Context, _ *emptypb.Empty) (*pb.RunnersResponse, error) {
	return &pb.RunnersResponse{
		Status:  0,
		Runners: h.runners,
		Count:   int32(len(h.runners)),
	}, nil
}

func (h *HttpRunnerServer) RemoveRunner(_ context.Context, request *pb.IdRunnerRequest) (*pb.SimpleResponse, error) {
	_, ok := h.runnerPairMap[request.RunnerId]
	if !ok {
		log.Printf("%s Runner ID not found", request.RunnerId)
		return &pb.SimpleResponse{
			Status:  404,
			Message: "Not Found",
		}, fmt.Errorf("runner Id not found: %s", request.RunnerId)
	}
	delete(h.runnerPairMap, request.RunnerId)
	idx := 0
	for i, r := range h.runners {
		if r.RunnerId == request.RunnerId {
			idx = i
			break
		}
	}
	h.runners = append(h.runners[:idx], h.runners[idx+1:]...)
	return &pb.SimpleResponse{
		Status:  0,
		Message: "OK",
	}, nil
}

func (h *HttpRunnerServer) CancelRunning(_ context.Context, _ *emptypb.Empty) (*pb.SimpleResponse, error) {

	if h.currentRunner != nil {
		log.Printf("No running job")
		return &pb.SimpleResponse{
			Status:  404,
			Message: "Not Found",
		}, errors.New("no running job")
	}
	pair, ok := h.runnerPairMap[h.currentRunner.RunnerId]
	if !ok {
		log.Printf("%s Runner ID not found", h.currentRunner.RunnerId)
		return &pb.SimpleResponse{
			Status:  404,
			Message: "Not Found",
		}, fmt.Errorf("runner Id not found: %s", h.currentRunner.RunnerId)
	}
	pair.runner.Cancelled = time.Now()
	return &pb.SimpleResponse{
		Status:  0,
		Message: "OK",
	}, nil
}

func (h *HttpRunnerServer) Listen(_ context.Context, _ *emptypb.Empty) (*pb.RunnerResponse, error) {
	return &pb.RunnerResponse{
		Status:  0,
		Message: "OK",
	}, nil
}

func (h *HttpRunnerServer) processRunners() {
	for true {
		h.currentRunner = <-h.runnerChan
		runner.WorkerRun(h.pbRunnerToRunner(h.currentRunner))
	}
}

func (h *HttpRunnerServer) pbRunnerToRunner(pbRunner *pb.Runner) runner.Runner {
	config := pbRunner.Config
	duration, err := time.ParseDuration(config.Duration)
	if err != nil {
		duration = time.Second * 10
	}
	reader := bufio.NewReader(strings.NewReader(config.Url.Headers + "\r\n"))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		log.Print(err)
	}
	_runner := runner.Runner{
		Config: runner.RunnerConfig{
			Duration:     duration,
			Workers:      int(config.Workers),
			NeedResponse: config.NeedResponse,
			Request: runner.APIRequest{
				Method:  config.Url.Method,
				URL:     config.Url.Url,
				Body:    config.Url.Body,
				Headers: http.Header(mimeHeader),
			},
			OutputCSVFilename: "",
			CountRequestSize:  false,
			CountResponseSize: false,
		},
		Metrics: nil,
		OnJobResponse: func(r *runner.Runner, response *http.Response) {
			pbRunner.StartTime, err = ptypes.TimestampProto(r.Start)
			if err != nil {
				log.Printf("Error: Timestamp conversion failed %v", err)
			}
			runner.DefaultRunner.OnJobResponse(r, response)
		},
		OnJobStart: func(r *runner.Runner) {
			pbRunner.Status = pb.Status_RUNNING
			pbRunner.StartTime, err = ptypes.TimestampProto(r.Start)
			runner.DefaultRunner.OnJobStart(r)
		},
		OnJobComplete: func(r *runner.Runner) {
			pbRunner.Status = pb.Status_DONE
			stats := convertToPbStat(utils.GetMetricsStat(r.Metrics))
			pbRunner.Stats = stats
			pbRunner.Progress = float32(r.GetProgress())
			runner.DefaultRunner.OnJobComplete(r)
		},
	}
	h.runnerPairMap[pbRunner.RunnerId] = runnerPair{
		pbRunner: pbRunner,
		runner:   &_runner,
	}
	return _runner
}

func convertToPbStat(stat map[string]utils.MetricStat) map[string]*pb.Stat {
	result := make(map[string]*pb.Stat)
	for key, item := range stat {
		pbStat := pb.Stat{
			Avg:    ptypes.DurationProto(item.Avg),
			Min:    ptypes.DurationProto(item.Min),
			Max:    ptypes.DurationProto(item.Max),
			P50:    ptypes.DurationProto(item.P50),
			P90:    ptypes.DurationProto(item.P90),
			P95:    ptypes.DurationProto(item.P95),
			P99:    ptypes.DurationProto(item.P99),
			Median: ptypes.DurationProto(item.Median),
			StdDev: ptypes.DurationProto(item.StdDev),
		}
		result[key] = &pbStat
	}
	return result
}

//func (h HttpRunnerServer) mustEmbedUnimplementedHttpRunnerServer() {
//	panic("implement me")
//}
