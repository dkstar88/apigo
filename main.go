package main

import (
	Runner "apigo/runner"
	"apigo/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"net/http"
	"time"
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
	progressbar.OptionSetWidth(15),
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
		Args: cobra.MinimumNArgs(1),
	}
	headers := ""
	rootCmd.PersistentFlags().IntVarP(&runnerConfig.Workers, "Concurrent Connections", "c", 10, "Number of concurrent connections")
	rootCmd.PersistentFlags().StringVar(&runnerConfig.OutputCSVFilename, "csv", "", "Output metrics to CSV file")
	rootCmd.PersistentFlags().StringVarP(&duration, "duration", "t", "1m", "Test duration")
	rootCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Method, "method", "x", "GET", "Request method")
	rootCmd.PersistentFlags().StringVarP(&runnerConfig.Request.Body, "body", "d", "", "Request body")
	rootCmd.PersistentFlags().StringVarP(&headers, "headers", "H", "", "Request Headers")
	runnerConfig.Request.Headers = Runner.StrToHeaders(headers)
	rootCmd.Run = func(cmd *cobra.Command, args []string) {

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
	rootCmd.Execute()
}
