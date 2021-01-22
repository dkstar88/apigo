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
	"apigo/utils"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"

	"github.com/spf13/cobra"
)

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "runner",
	Short: "Runner is a http load testing tool",
	Long: `Runner is a http load testing tool provides meaningful 
			statistic information on the test.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("--help get usage")
	},
}

func Execute() {
	Runner.DefaultRunner.OnJobStart = ConsoleRunnerOnJobStart
	Runner.DefaultRunner.OnJobComplete = ConsoleRunnerOnJobComplete
	Runner.DefaultRunner.OnJobResponse = OnRunnerJobResponse

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var ticker = time.NewTicker(500 * time.Millisecond)
var done = make(chan bool)

func ConsoleRunnerOnJobComplete(runner *Runner.Runner) {
	_ = bar.Finish()
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
	_ = bar.Set(0)
	go func(runner *Runner.Runner) {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				progress := runner.GetProgress() * 10000
				//fmt.Printf("%f, %f Duration Seconds\n", time.Now().Sub(runner.Start).Seconds(), runner.Config.Duration.Seconds())
				bar.Describe(fmt.Sprintf("%d/%d Jobs Completed",
					atomic.LoadInt64(&runner.JobsProcessed),
					atomic.LoadInt64(&runner.JobsCreated)))
				_ = bar.Set(int(progress))
			}
		}
	}(runner)
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

func OnRunnerJobResponse(runner *Runner.Runner, _ *http.Response) {
	// Calc progress
	progress := time.Now().Sub(runner.Start).Seconds() / runner.Config.Duration.Seconds() * 10000
	_ = bar.Set(int(progress))
}
