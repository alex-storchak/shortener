package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type stubDBPinger struct {
	PingErr error
}

func (m *stubDBPinger) PingContext(_ context.Context) error {
	return m.PingErr
}

type stubSleepDBPinger struct {
	sleep time.Duration
}

func (s *stubSleepDBPinger) PingContext(ctx context.Context) error {
	select {
	case <-time.After(s.sleep):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestPingDBService_Ping(t *testing.T) {
	tests := []struct {
		name          string
		db            DBPinger
		wantErr       bool
		wantErrIs     error
		checkDuration bool
	}{
		{
			name:    "success ping",
			db:      &stubDBPinger{PingErr: nil},
			wantErr: false,
		},
		{
			name:      "db returns ErrFailedToPingDB error",
			db:        &stubDBPinger{PingErr: errors.New("db down")},
			wantErr:   true,
			wantErrIs: ErrFailedToPingDB,
		},
		{
			name:          "timeout maps to ErrFailedToPingDB and returns quickly",
			db:            &stubSleepDBPinger{sleep: 5 * time.Second},
			wantErr:       true,
			wantErrIs:     ErrFailedToPingDB,
			checkDuration: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &PingDBService{db: tt.db, logger: zap.NewNop()}

			var start time.Time
			if tt.checkDuration {
				start = time.Now()
			}

			err := srv.Ping()

			if tt.wantErr {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NoError(t, err)
			}

			if tt.checkDuration {
				assert.LessOrEqual(t, time.Since(start), 2*time.Second)
			}
		})
	}
}
