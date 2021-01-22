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

	ColorPrintSummary("Iterations", color.FgGreen, fmt.Sprintf("%d", runner.JobsProcessed))
	ColorPrintSummary("Success", color.FgGreen, fmt.Sprintf("%d", runner.JobsSuccessful))
	ColorPrintSummary("Failed", color.FgRed, fmt.Sprintf("%d", runner.JobsFailed))
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
