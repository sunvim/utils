/*
 * Linear Allocator
 *
 * Improve the memory allocation and garbage collection performance.
 *
 * Copyright (C) 2020-2023 crazybie@github.com.
 * https://github.com/crazybie/linear_ac
 */

// Pool pros over sync.Pool:
// 1. generic API
// 2. no boxing
// 3. support reserving.
// 4. debug mode: leak detecting, duplicated put.
// 5. max size: memory leak protection.

package lac

type Pool[T any] struct {
	Logger
	m      spinLock
	New    func() T
	pool   []T
	Cap    int
	newCnt int
	// the max count of call to New function.
	MaxNew int
	Name   string

	CheckDuplication bool
	// require CheckDuplication=true
	// check duplicated put.
	Equal func(a, b T) bool
}

func (p *Pool[T]) Get() T {
	p.m.Lock()
	defer p.m.Unlock()

	if len(p.pool) == 0 {
		return p.doNew()
	}

	last := len(p.pool) - 1
	r := p.pool[last]
	var zero T
	p.pool[last] = zero
	p.pool = p.pool[:last]
	return r
}

func (p *Pool[T]) doNew() T {
	p.newCnt++
	if p.MaxNew > 0 && p.newCnt > p.MaxNew {
		errorf(p, "%s: pool exhausted (%v), potential leak", p.Name, p.MaxNew)
	}
	return p.New()
}

func (p *Pool[T]) Put(v T) bool {
	p.m.Lock()
	defer p.m.Unlock()

	if p.CheckDuplication && p.Equal != nil {
		for _, i := range p.pool {
			if p.Equal(i, v) {
				errorf(p, "%s: duplicated: %v, %v", p.Name, i, v)
			}
		}
	}

	if (p.Cap == 0 || len(p.pool) < p.Cap) || p.CheckDuplication {
		p.pool = append(p.pool, v)
		return true
	} else {
		return false
	}
}

func (p *Pool[T]) Clear() {
	p.m.Lock()
	defer p.m.Unlock()

	p.pool = nil
}

func (p *Pool[T]) Reserve(cnt int) {
	p.m.Lock()
	defer p.m.Unlock()

	p.pool = make([]T, cnt)
	for i := 0; i < cnt; i++ {
		p.pool[i] = p.doNew()
	}
}

func (p *Pool[T]) DebugCheck() {
	l := len(p.pool)
	if l != p.newCnt {
		errorf(p, "%s: %d leaked. cur:%v,max: %v", p.Name, p.newCnt-l, p.newCnt, l)
	}
}
