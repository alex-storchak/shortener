package pool

import "sync"

type Resetter interface {
	Reset()
}

type ResetterPool[T Resetter] struct {
	pool sync.Pool
}

func NewResetterPool[T Resetter](newFunc func() T) *ResetterPool[T] {
	return &ResetterPool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *ResetterPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *ResetterPool[T]) Put(r T) {
	p.pool.Put(r)
}
