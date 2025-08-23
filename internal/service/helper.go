package service

import (
	"fmt"
	"time"
)

func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.2f μs", float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf("%.2f ms", float64(d.Milliseconds()))
	default:
		return fmt.Sprintf("%.2f s", d.Seconds())
	}
}
