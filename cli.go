package main

import (
	"apigo/runner/models"
	"apigo/runner/services"
	"time"
)

func main() {
	config := models.NewRunner(10, 10)
	config.OutputCSVFilename = "output.csv"
	config.Duration, _ = time.ParseDuration("3s")
	services.RunnerRun(*config)
}
