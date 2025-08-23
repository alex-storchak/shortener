package service

import (
	"fmt"
	"time"
)

const (
	nanoSecFmt  = "%d ns"
	microSecFmt = "%.2f μs"
	milliSecFmt = "%.2f ms"
	secFmt      = "%.2f s"
)

func FormatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf(nanoSecFmt, d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf(microSecFmt, float64(d.Microseconds()))
	case d < time.Second:
		return fmt.Sprintf(milliSecFmt, float64(d.Milliseconds()))
	default:
		return fmt.Sprintf(secFmt, d.Seconds())
	}
}
