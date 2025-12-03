package audit

import (
	"context"
	"sync"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
	"go.uber.org/zap"
)

type Observer interface {
	Notify(ctx context.Context, e model.AuditEvent)
	Close(ctx context.Context) error
	Name() string
}

type EventManager struct {
	observers []Observer
	ch        chan model.AuditEvent
	wg        sync.WaitGroup
	closed    chan struct{}
	once      sync.Once
	logger    *zap.Logger
}

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

func (m *EventManager) Publish(e model.AuditEvent) {
	select {
	case m.ch <- e:
	default:
		m.logger.Warn("drop event (queue full)", zap.Any("event", e))
	}
}

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
