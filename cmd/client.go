/*
Copyright © 2021 daochun.zhao <daochun.zhao@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"apigo/grpc/httprunner"
	Runner "apigo/runner"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Starts a HTTP runner server",
	Long:  `Starts a HTTP runner server`,
	Run:   client,
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.PersistentFlags().IntVarP(&runnerConfig.Workers, "Concurrent Connections", "c", 10, "Number of concurrent connections")
	clientCmd.PersistentFlags().StringVarP(&duration, "duration", "t", "1m", "Test duration")
	clientCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Method, "method", "x", "GET", "Request method")
	clientCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Body, "body", "d", "", "Request body")
	clientCmd.PersistentFlags().StringVarP(&headers, "headers", "H", "", "Request Headers")
	clientCmd.PersistentFlags().String("Host", "127.0.0.1", "gRPC Host address")
	clientCmd.PersistentFlags().UintP("Port", "p", 7000, "gRPC Host Port")
}

func client(cmd *cobra.Command, args []string) {
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
	defer func() {
		_ = conn.Close()
	}()
	client := httprunner.NewHttpRunnerClient(conn)
	rc := httprunner.RunnerConfig{
		Duration:     duration,
		Workers:      int32(runnerConfig.Workers),
		NeedResponse: runnerConfig.NeedResponse,
		Url: &httprunner.Url{
			Url:     runnerConfig.Request.URL,
			Method:  runnerConfig.Request.Method,
			Body:    runnerConfig.Request.Body,
			Headers: headers,
		},
	}
	enqueueRes, err := client.Enqueue(context.Background(), &rc)
	if err != nil {
		log.Printf("error enqueue request %v", err)
	}
	println("Runner Enqueued ID: ", enqueueRes.Runner.RunnerId)
	waitUtil := runnerConfig.Duration
	idReq := httprunner.IdRunnerRequest{RunnerId: enqueueRes.Runner.RunnerId}
	for {
		if waitUtil > 0 {
			time.Sleep(time.Millisecond * 500)
			waitUtil -= time.Millisecond * 500
		}
		runnerRes, err := client.GetRunner(context.Background(), &idReq)
		if err != nil {
			log.Printf("error GetRunner %v", err)
		}
		strStatus := ""
		switch runnerRes.Runner.Status {
		case 1:
			strStatus = "QUEUED"
		case 2:
			strStatus = "RUNNING"
		case 3:
			strStatus = "DONE"
		case 9:
			strStatus = "ERROR"
		default:
			strStatus = "Unknown"
		}
		println("Runner Status: ", runnerRes.Runner.Status, strStatus)
		if runnerRes.Runner.Status == 3 {
			break
		}
	}
}
