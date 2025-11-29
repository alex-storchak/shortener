package observer

import (
	"context"
	"sync"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
)

// HTTP is an observer that sends audit events to a remote HTTP server.
// It uses a worker pool pattern to handle concurrent event sending with configurable limits.
type HTTP struct {
	cfg    config.Audit
	client *resty.Client
	logger *zap.Logger
	queue  chan model.AuditEvent
	wg     sync.WaitGroup
	once   sync.Once
}

// NewHTTP creates a new HTTP observer with the specified configuration.
// It initializes the HTTP client with timeouts and content-type headers.
//
// Parameters:
//   - cfg: Audit configuration for HTTP-specific settings
//   - l: Structured logger for logging operations
//
// Returns: Initialized HTTP observer (not yet started)
func NewHTTP(cfg config.Audit, l *zap.Logger) *HTTP {
	client := resty.New().
		SetTimeout(cfg.HTTPTimeout).
		SetHeader("Content-Type", "application/json")

	h := &HTTP{
		cfg:    cfg,
		client: client,
		logger: l,
		queue:  make(chan model.AuditEvent, cfg.EventChanSize),
	}
	return h
}

// Name returns human-readable identifier for the observer.
func (o *HTTP) Name() string {
	return "audit_http"
}

// Start begins processing events by starting the worker pool.
// This method should be called after creating the observer to begin event processing.
func (o *HTTP) Start() {
	for i := 0; i < o.cfg.HTTPWorkersCount; i++ {
		o.wg.Add(1)
		go o.worker()
	}
}

// Notify queues an audit event for sending to the HTTP server.
// If the internal queue is full, the event is dropped with a warning.
//
// Parameters:
//   - ctx: Context (currently unused but part of interface)
//   - e: AuditEvent to send to the remote server
func (o *HTTP) Notify(_ context.Context, e model.AuditEvent) {
	o.logger.Debug("event received", zap.Any("event", e))
	select {
	case o.queue <- e:
	default:
		o.logger.Warn("drop event (queue full)", zap.Any("event", e))
	}
}

// worker processes events from the internal queue and sends them to the HTTP server.
// It runs in a separate goroutine and exits when the queue is closed.
func (o *HTTP) worker() {
	defer o.wg.Done()
	for e := range o.queue {
		o.send(e)
	}
}

// send sends a single audit event to the HTTP server.
// It uses the initialized HTTP client to make a POST request.
// If the request fails, it logs the error but does not return it.
func (o *HTTP) send(e model.AuditEvent) {
	_, err := o.client.R().
		SetBody(e).
		Post(o.cfg.URL)
	if err != nil {
		o.logger.Error("send event to http audit error", zap.Error(err))
		return
	}
}

// Close gracefully shuts down the HTTP observer.
// It stops accepting new events and waits for in-flight events to be processed.
//
// Parameters:
//   - ctx: Context for controlling shutdown timeout
//
// Returns: Always returns nil (error is logged but not returned)
func (o *HTTP) Close(ctx context.Context) error {
	o.once.Do(func() {
		close(o.queue)
	})

	done := make(chan struct{})
	go func() {
		o.wg.Wait()
		o.logger.Info("audit http observer workers closed")
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}
	o.logger.Info("audit http observer closed")
	return nil
}
