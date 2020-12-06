package utils

import (
	"apigo/runner"
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"os"
)

func ConsoleOutput(runner *runner.Runner) {
	// Summary
	ColorPrintSummary("URL", color.FgGreen, runner.Config.Request.URL)
	ColorPrintSummary("Workers", color.FgGreen, fmt.Sprintf("%d", runner.Config.Workers))
	ColorPrintSummary("Time Started", color.FgGreen, runner.Start.String())
	ColorPrintSummary("Iterations", color.FgGreen, fmt.Sprintf("%d", runner.JobsProcessed))
	ColorPrintSummary("RPS", color.FgGreen, fmt.Sprintf("%.2f", float64(runner.JobsProcessed)/runner.Config.Duration.Seconds()))

	fields := []string{"HTTPBlocked", "HTTPDNS", "HTTPTls",
		"HTTPConnecting", "HTTPSending", "HTTPWaiting",
		"HTTPReceiving", "HTTPTotal"}
	fieldNames := []string{"Blocked", "DNS", "Tls",
		"Connecting", "Sending", "Waiting",
		"Receiving", "Total"}
	columns := []string{"", "Min", "Avg", "P50", "P90", "P95", "Max"}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(columns)
	for i, f := range fields {
		durations := MetricsExtract(runner.Metrics, f)
		stats := GetStats(durations)
		table.Append([]string{
			fieldNames[i],
			DurationToString(stats.Min),
			DurationToString(stats.Avg),
			DurationToString(stats.P50),
			DurationToString(stats.P90),
			DurationToString(stats.P95),
			DurationToString(stats.Max),
		})
	}
	table.Render() // Send output
}
