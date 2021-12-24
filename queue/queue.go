package queue


import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/eapache/queue.v1"
)

//Queue queue
type Queue struct {
	sync.Mutex
	popable *sync.Cond
	buffer  *queue.Queue
	closed  bool
	count   int32
	cc      chan interface{}
	once    sync.Once
}

//New 创建
func New() *Queue {
	ch := &Queue{
		buffer: queue.New(),
	}
	ch.popable = sync.NewCond(&ch.Mutex)
	return ch
}

//Pop 取出队列,（阻塞模式）
func (q *Queue) Pop() (v interface{}) {
	c := q.popable

	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	for q.Len() == 0 && !q.closed {
		c.Wait()
	}

	if q.closed { //已关闭
		return
	}

	if q.Len() > 0 {
		buffer := q.buffer
		v = buffer.Peek()
		buffer.Remove()
		atomic.AddInt32(&q.count, -1)
	}
	return
}

// TryPop 试着取出队列（非阻塞模式）返回ok == false 表示空
func (q *Queue) TryPop() (v interface{}, ok bool) {
	buffer := q.buffer

	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	if q.Len() > 0 {
		v = buffer.Peek()
		buffer.Remove()
		atomic.AddInt32(&q.count, -1)
		ok = true
	} else if q.closed {
		ok = true
	}

	return
}

// TryPopTimeout 试着取出队列（塞模式+timeout）返回ok == false 表示超时
func (q *Queue) TryPopTimeout(tm time.Duration) (v interface{}, ok bool) {
	q.once.Do(func() {
		q.cc = make(chan interface{}, 1)
	})
	go func() {
		q.popChan(&q.cc)
	}()

	ok = true
	timeout := time.After(tm)
	select {
	case v = <-q.cc:
	case <-timeout:
		if !q.closed {
			q.popable.Signal()
		}
		ok = false
	}

	return
}

//Pop 取出队列,（阻塞模式）
func (q *Queue) popChan(v *chan interface{}) {
	c := q.popable

	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	for q.Len() == 0 && !q.closed {
		c.Wait()
	}

	if q.closed { //已关闭
		*v <- nil
		return
	}

	if q.Len() > 0 {
		buffer := q.buffer
		tmp := buffer.Peek()
		buffer.Remove()
		atomic.AddInt32(&q.count, -1)
		*v <- tmp
	} else {
		*v <- nil
	}
	return
}

// Push 插入队列，非阻塞
func (q *Queue) Push(v interface{}) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if !q.closed {
		q.buffer.Add(v)
		atomic.AddInt32(&q.count, 1)
		q.popable.Signal()
	}
}

// Len 获取队列长度
func (q *Queue) Len() int {
	return (int)(atomic.LoadInt32(&q.count))
}

// Close Queue
// After close, Pop will return nil without block, and TryPop will return v=nil, ok=True
func (q *Queue) Close() {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	if !q.closed {
		q.closed = true
		atomic.StoreInt32(&q.count, 0)
		q.popable.Broadcast() //广播
	}
}

// IsClose check is closed
func (q *Queue) IsClose() bool {
	return q.closed
}

//Wait 等待队列消费完成
func (q *Queue) Wait() {
	for {
		if q.closed || q.Len() == 0 {
			break
		}

		runtime.Gosched() //出让时间片
	}
}
