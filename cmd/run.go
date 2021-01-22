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
	Runner "apigo/runner"
	"time"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runner is a http load testing tool",
	Long: `Runner is a http load testing tool provides meaningful 
			statistic information on the test.`,
	Args: cobra.MinimumNArgs(1),
	Run:  run,
}

var (
	duration     string
	headers      string
	runnerConfig = Runner.RunnerConfig{}
)

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().IntVarP(&runnerConfig.Workers, "Concurrent Connections", "c", 10, "Number of concurrent connections")
	runCmd.PersistentFlags().StringVar(&runnerConfig.OutputCSVFilename, "csv", "", "Output metrics to CSV file")
	runCmd.PersistentFlags().StringVarP(&duration, "duration", "t", "1m", "Test duration")
	runCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Method, "method", "x", "GET", "Request method")
	runCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Body, "body", "d", "", "Request body")
	runCmd.PersistentFlags().StringVarP(&headers, "headers", "H", "", "Request Headers")
	runnerConfig.Request.Headers = Runner.StrToHeaders(headers)
}

func run(_ *cobra.Command, args []string) {
	runnerConfig.Duration, _ = time.ParseDuration(duration)
	runnerConfig.Request.URL = args[0]
	//bar.Reset()
	runner := Runner.NewRunner(runnerConfig)
	//fmt.Printf("%v\n", runnerConfig)

	Runner.WorkerRun(*runner)
}
