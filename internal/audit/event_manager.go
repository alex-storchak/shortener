package audit

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
)

// Observer defines the interface for audit event destinations.
// Implementations handle receiving, processing, and storing audit events.
type Observer interface {
	// Notify sends an audit event to the observer for processing.
	Notify(ctx context.Context, e model.AuditEvent)

	// Close gracefully shuts down the observer and releases resources.
	Close(ctx context.Context) error

	// Name returns the identifier of the observer for logging purposes.
	Name() string
}

// EventManager is the central dispatcher for audit events.
// It receives events from publishers and distributes them to all registered observers.
type EventManager struct {
	observers []Observer
	ch        chan model.AuditEvent
	wg        sync.WaitGroup
	closed    chan struct{}
	once      sync.Once
	logger    *zap.Logger
}

// NewEventManager creates a new EventManager with the specified observers and configuration.
// It starts the internal dispatch goroutine immediately.
//
// Parameters:
//   - observers: List of observers to receive events
//   - cfg: Audit configuration for channel sizing
//   - l: Structured logger for logging operations
//
// Returns: Initialized EventManager ready to receive events
func NewEventManager(observers []Observer, cfg config.Audit, l *zap.Logger) *EventManager {
	em := &EventManager{
		observers: observers,
		ch:        make(chan model.AuditEvent, cfg.EventChanSize),
		closed:    make(chan struct{}),
		logger:    l,
	}
	em.wg.Add(1)
	go em.dispatch()
	return em
}

// dispatch runs in a separate goroutine
// and handles the distribution of audit events to observers.
// It processes events from the internal channel and forwards them to each observer.
// If the channel is closed, it gracefully shuts down all observers and exits.
func (m *EventManager) dispatch() {
	defer m.wg.Done()
	for {
		select {
		case e, ok := <-m.ch:
			if !ok {
				m.logger.Info("audit events channel closed, finish dispatching")
				return
			}
			for _, obs := range m.observers {
				obs.Notify(context.Background(), e)
			}
		case <-m.closed:
			m.logger.Info("audit EM is closing, finish dispatching")
			return
		}
	}
}

// Publish sends an audit event to all registered observers.
// If no observers are registered, it returns immediately (no-op).
// If the event channel is full, the event is dropped with a warning.
//
// Parameters:
//   - e: AuditEvent to publish to all observers
func (m *EventManager) Publish(e model.AuditEvent) {
	if len(m.observers) == 0 {
		return // no-op
	}
	select {
	case m.ch <- e:
	default:
		m.logger.Warn("drop event (queue full)", zap.Any("event", e))
	}
}

// Close gracefully shuts down the EventManager and all observers.
// It ensures all in-flight events are processed before closing.
//
// Parameters:
//   - ctx: Context for controlling shutdown timeout
func (m *EventManager) Close(ctx context.Context) {
	m.logger.Info("closing audit EM")
	m.once.Do(func() {
		close(m.closed)
		close(m.ch)
		done := make(chan struct{})
		go func() {
			m.wg.Wait()
			for _, o := range m.observers {
				if cErr := o.Close(ctx); cErr != nil {
					m.logger.Error("failed to close audit observer",
						zap.String("name", o.Name()),
						zap.Error(cErr),
					)
				}
			}
			m.logger.Info("audit observers closed")
			close(done)
		}()
		select {
		case <-done:
			m.logger.Debug("audit EM closed after closing all observers")
		case <-ctx.Done():
			m.logger.Debug("audit EM closed after closing context")
		}
		m.logger.Info("audit EM closed")
	})
}
