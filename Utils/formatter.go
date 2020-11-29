package Utils

import (
	"fmt"
	"github.com/icza/gox/timex"
	"time"
)

func ByteCountIEC(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func DurationToString(duration time.Duration) string {
	return timex.Round(duration, 2).String()
}
