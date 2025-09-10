package typepool

import "sync"

type Pool[T any] struct {
	sync.Pool
}

func New[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		Pool: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	v := p.Pool.Get()
	if v == nil {
		var zero T
		return zero
	}
	return v.(T)
}

func (p *Pool[T]) Put(x T) {
	p.Pool.Put(x)
}
