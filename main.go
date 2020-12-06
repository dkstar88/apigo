package main

import (
	"apigo/utils"
	Runner "apigo/runner"
	"time"
)

func ConsoleRunnerOnJobComplete(runner *Runner.Runner) {
	if len(runner.Config.OutputCSVFilename) > 0 {
		utils.MetricsToCsv(runner.Metrics, runner.Config.OutputCSVFilename)
	}
	utils.ConsoleOutput(runner)
}

func main() {
	runner := Runner.NewRunner(10, 10)
	runner.Config.OutputCSVFilename = "output.csv"
	runner.Config.Duration, _ = time.ParseDuration("3s")
	runner.OnJobComplete = ConsoleRunnerOnJobComplete
	Runner.WorkerRun(*runner)
}
