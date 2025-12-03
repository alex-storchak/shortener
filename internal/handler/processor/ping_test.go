package processor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/service"
)

type stubPinger struct {
	PingErr error
}

func (m *stubPinger) IsReady(_ context.Context) error {
	return m.PingErr
}

type stubSleepPinger struct {
	sleep time.Duration
}

func (s *stubSleepPinger) IsReady(ctx context.Context) error {
	select {
	case <-time.After(s.sleep):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestPingService_Ping(t *testing.T) {
	tests := []struct {
		name          string
		db            service.Pinger
		wantErr       bool
		checkDuration bool
	}{
		{
			name:    "success ping",
			db:      &stubPinger{PingErr: nil},
			wantErr: false,
		},
		{
			name:    "db returns ErrFailedToPingDB error",
			db:      &stubPinger{PingErr: errors.New("db down")},
			wantErr: true,
		},
		{
			name:          "timeout maps to ErrFailedToPingDB and returns quickly",
			db:            &stubSleepPinger{sleep: 5 * time.Second},
			wantErr:       true,
			checkDuration: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Ping{pinger: tt.db, logger: zap.NewNop()}

			var start time.Time
			if tt.checkDuration {
				start = time.Now()
			}

			err := srv.Process()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.checkDuration {
				assert.LessOrEqual(t, time.Since(start), 2*time.Second)
			}
		})
	}
}
