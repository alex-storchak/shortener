package observer

import (
	"context"
	"sync"

	"github.com/alex-storchak/shortener/internal/config"
	"github.com/alex-storchak/shortener/internal/model"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type HTTP struct {
	cfg    config.Audit
	client *resty.Client
	logger *zap.Logger
	queue  chan model.AuditEvent
	wg     sync.WaitGroup
	once   sync.Once
}

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

func (o *HTTP) Name() string {
	return "audit_http"
}

func (o *HTTP) Start() {
	for i := 0; i < o.cfg.HTTPWorkersCount; i++ {
		o.wg.Add(1)
		go o.worker()
	}
}

func (o *HTTP) Notify(_ context.Context, e model.AuditEvent) {
	o.logger.Debug("event received", zap.Any("event", e))
	select {
	case o.queue <- e:
	default:
		o.logger.Warn("drop event (queue full)", zap.Any("event", e))
	}
}

func (o *HTTP) worker() {
	defer o.wg.Done()
	for e := range o.queue {
		o.send(e)
	}
}

func (o *HTTP) send(e model.AuditEvent) {
	_, err := o.client.R().
		SetBody(e).
		Post(o.cfg.URL)
	if err != nil {
		o.logger.Error("send event to http audit error", zap.Error(err))
		return
	}
}

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
