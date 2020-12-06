package main

import (
	"apigo/utils"
	Runner "apigo/runner"
	"github.com/schollz/progressbar/v3"
	"net/http"
	"time"
)

func ConsoleRunnerOnJobComplete(runner *Runner.Runner) {
	bar.Finish()
	if len(runner.Config.OutputCSVFilename) > 0 {
		utils.MetricsToCsv(runner.Metrics, runner.Config.OutputCSVFilename)
	}
	utils.ConsoleOutput(runner)
}

func ConsoleRunnerOnJobStart(runner *Runner.Runner) {
	bar.Set(0)
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

func main() {
	bar.Reset()
	runner := Runner.NewRunner(10, 10)
	runner.Config.OutputCSVFilename = "output.csv"
	runner.Config.Duration, _ = time.ParseDuration("3s")
	runner.Config.NeedResponse = true
	runner.OnJobStart = ConsoleRunnerOnJobStart
	runner.OnJobComplete = ConsoleRunnerOnJobComplete
	runner.OnJobResponse = OnRunnerJobResponse
	Runner.WorkerRun(*runner)
}
