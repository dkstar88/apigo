package main

import (
	"apigo/grpc/httprunner"
	Runner "apigo/runner"
	"apigo/utils"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"log"
	"net"
	"net/http"
	"time"
	"google.golang.org/grpc"
	apigoGrpc "apigo/grpc"
)


var ticker = time.NewTicker(500 * time.Millisecond)
var done = make(chan bool)

func ConsoleRunnerOnJobComplete(runner *Runner.Runner) {
	bar.Finish()
	println()
	if len(runner.Config.OutputCSVFilename) > 0 {
		utils.MetricsToCsv(runner.Metrics, runner.Config.OutputCSVFilename)
	}
	utils.ConsoleOutput(runner)
	ticker.Stop()
	done <- true
}

func ConsoleRunnerOnJobStart(runner *Runner.Runner) {
	utils.ColorPrintSummary("URL", color.FgGreen, runner.Config.Request.URL)
	utils.ColorPrintSummary("Workers", color.FgGreen, fmt.Sprintf("%d", runner.Config.Workers))
	utils.ColorPrintSummary("Time Started", color.FgGreen, runner.Start.String())
	bar.Set(0)
	go func(runner *Runner.Runner) {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				progress := time.Now().Sub(runner.Start).Seconds() / runner.Config.Duration.Seconds() * 10000
				//fmt.Printf("%f, %f Duration Seconds\n", time.Now().Sub(runner.Start).Seconds(), runner.Config.Duration.Seconds())
				bar.Describe(fmt.Sprintf("%d/%d Jobs Completed", runner.JobsProcessed, runner.JobsCreated))
				bar.Set(int(progress))
			}
		}
	} (runner)
}

var bar = progressbar.NewOptions(10000,
	progressbar.OptionEnableColorCodes(true),
	progressbar.OptionSetWidth(30),
	progressbar.OptionSetDescription("[cyan][reset] Running..."),
	progressbar.OptionSetTheme(progressbar.Theme{
		Saucer:        "[green]=[reset]",
		SaucerHead:    "[green]>[reset]",
		SaucerPadding: " ",
		BarStart:      "[",
		BarEnd:        "]",
	}))

func OnRunnerJobResponse (runner *Runner.Runner, response *http.Response) {
	// Calc progress
	progress := time.Now().Sub(runner.Start).Seconds() / runner.Config.Duration.Seconds() * 10000
	bar.Set(int(progress))

}

var runnerConfig = Runner.RunnerConfig {

}


func main() {
	var duration string = ""
	var rootCmd = &cobra.Command{
		Use:   "runner",
		Short: "Runner is a http load testing tool",
		Long: `Runner is a http load testing tool provides meaningful 
			statistic information on the test.`,
	}
	headers := ""
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Runner is a http load testing tool",
		Long: `Runner is a http load testing tool provides meaningful 
			statistic information on the test.`,
		Args: cobra.MinimumNArgs(1),
	}

	runCmd.PersistentFlags().IntVarP(&runnerConfig.Workers, "Concurrent Connections", "c", 10, "Number of concurrent connections")
	runCmd.PersistentFlags().StringVar(&runnerConfig.OutputCSVFilename, "csv", "", "Output metrics to CSV file")
	runCmd.PersistentFlags().StringVarP(&duration, "duration", "t", "1m", "Test duration")
	runCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Method, "method", "x", "GET", "Request method")
	runCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Body, "body", "d", "", "Request body")
	runCmd.PersistentFlags().StringVarP(&headers, "headers", "H", "", "Request Headers")
	runnerConfig.Request.Headers = Runner.StrToHeaders(headers)
	runCmd.Run = func(cmd *cobra.Command, args []string) {

		runnerConfig.Duration, _ = time.ParseDuration(duration)
		runnerConfig.Request.URL = args[0]
		//bar.Reset()
		runner := Runner.NewRunner(runnerConfig)
		//fmt.Printf("%v\n", runnerConfig)
		runner.OnJobStart = ConsoleRunnerOnJobStart
		runner.OnJobComplete = ConsoleRunnerOnJobComplete
		runner.OnJobResponse = OnRunnerJobResponse

		Runner.WorkerRun(*runner)

	}
	serverCmd := &cobra.Command{
		Use:   "server",
		Aliases: []string {"s"},
		Short: "Starts a HTTP runner server",
		Long: `Starts a HTTP runner server`,
	}
	serverCmd.PersistentFlags().String("Host", "127.0.0.1", "gRPC Host address")
	serverCmd.PersistentFlags().UintP("Port", "p", 7000, "gRPC Host Port")
	serverCmd.Run = func(cmd *cobra.Command, args []string) {
		address, _ := cmd.PersistentFlags().GetString("Host")
		port, _ := cmd.PersistentFlags().GetUint("Port")
		startGrpcServer(address, port)
		println("Server started.")
	}

	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Starts a HTTP runner server",
		Long: `Starts a HTTP runner server`,
	}
	clientCmd.PersistentFlags().IntVarP(&runnerConfig.Workers, "Concurrent Connections", "c", 10, "Number of concurrent connections")
	clientCmd.PersistentFlags().StringVarP(&duration, "duration", "t", "1m", "Test duration")
	clientCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Method, "method", "x", "GET", "Request method")
	clientCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Body, "body", "d", "", "Request body")
	clientCmd.PersistentFlags().StringVarP(&headers, "headers", "H", "", "Request Headers")
	clientCmd.PersistentFlags().String("Host", "127.0.0.1", "gRPC Host address")
	clientCmd.PersistentFlags().UintP("Port", "p", 7000, "gRPC Host Port")
	clientCmd.Run = func(cmd *cobra.Command, args []string) {
		runnerConfig.Request.Headers = Runner.StrToHeaders(headers)
		runnerConfig.Duration, _ = time.ParseDuration(duration)
		runnerConfig.Request.URL = args[0]
		address, _ := cmd.PersistentFlags().GetString("Host")
		port, _ := cmd.PersistentFlags().GetUint("Port")

		serverAddress := fmt.Sprintf("%s:%d", address, port)
		conn, e := grpc.Dial(serverAddress, grpc.WithInsecure())
		if e != nil {
			log.Fatalf("failed to connect to NewHttpRunner server: %s", e)
		}
		defer conn.Close()
		client := httprunner.NewHttpRunnerClient(conn)
		rc := httprunner.RunnerConfig{
			RunnerId:     0,
			Duration:     duration,
			Workers:      int32(runnerConfig.Workers),
			NeedResponse: runnerConfig.NeedResponse,
			Url:          &httprunner.Url{
				Url:     runnerConfig.Request.URL,
				Method:  runnerConfig.Request.Method,
				Body:    runnerConfig.Request.Body,
				Headers: headers,
			},
			Status:       0,
		}
		client.Enqueue(context.Background(), &rc)
	}

	rootCmd.AddCommand(runCmd, serverCmd, clientCmd)

	rootCmd.Execute()
}

func startGrpcServer(address string, port uint) {
	netListener := getNetListener(address, port)
	gRPCServer := grpc.NewServer()
	grpcRunnerImpl := apigoGrpc.NewHttpRunnerServer()
	//gRPCServer.RegisterService(&grpcRunnerImpl, &apigoGrpc.HttpRunnerServer{})
	httprunner.RegisterHttpRunnerServer(gRPCServer, grpcRunnerImpl)

	println(fmt.Sprintf("Starting gRPC server at %s:%d", address, port))
	// start the server
	if err := gRPCServer.Serve(netListener); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

}

func getNetListener(address string, port uint) net.Listener {
	lis, err := net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d", address, port),
	)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	return lis
}