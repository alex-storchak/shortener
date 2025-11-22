package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "nano second",
			duration: 123 * time.Nanosecond,
			want:     "123 ns",
		},
		{
			name:     "micro second",
			duration: 123 * time.Microsecond,
			want:     "123.00 μs",
		},
		{
			name:     "milli second",
			duration: 123 * time.Millisecond,
			want:     "123.00 ms",
		},
		{
			name:     "second",
			duration: 123 * time.Second,
			want:     "123.00 s",
		},
		{
			name:     "composite",
			duration: 123*time.Second + 456*time.Millisecond,
			want:     "123.46 s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatDuration(tt.duration))
		})
	}
}
