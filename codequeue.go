package codequeue

import (
	"sync"
)

const (
	// 必须得是2^x
	_MinCap = 32
)

// CoDequeue 一种线程安全的双端队列
type CoDequeue struct {
	mu   sync.RWMutex
	cond *sync.Cond

	buffer []interface{}
	cnt    int
	head   int
	tail   int
}

// NewCoDequeue 返回一个CoDequeue实例.
func NewCoDequeue(cap int) *CoDequeue {
	if cap < 32 {
		cap = 32
	}

	q := &CoDequeue{}
	lock := sync.Mutex{}
	q.cond = sync.NewCond(&lock)

	q.buffer = make([]interface{}, cap)
	q.cnt = 0
	q.head = 0
	q.tail = 0
	return q
}

// Len 返回队列长度.
func (q *CoDequeue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.cnt
}

// FPush 数据从前端入队.
func (q *CoDequeue) FPush(x interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.growIfNeeded()

	q.head = q.prev(q.head)
	q.buffer[q.head] = x
	q.cnt++
}

// FPop 数据从前端出队.
func (q *CoDequeue) FPop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.cnt <= 0 {
		panic("CoDequeue: FPop() is called on empty queue")
	}

	x := q.buffer[q.head]
	q.buffer[q.head] = nil
	q.head = q.next(q.head)
	q.cnt--

	q.shrinkIfNeeded()

	return x
}

// BPush 数据从后端入队.
func (q *CoDequeue) BPush(elem interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.growIfNeeded()

	q.buffer[q.tail] = elem
	q.tail = q.next(q.tail)
	q.cnt++
}

// BPop 数据从后端出队.
func (q *CoDequeue) BPop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.cnt <= 0 {
		panic("CoDequeue: BPop() is called on empty queue")
	}

	q.tail = q.prev(q.tail)
	x := q.buffer[q.tail]
	q.buffer[q.tail] = nil
	q.cnt--

	q.shrinkIfNeeded()

	return x
}

// Front 数据从前端读取.
func (q *CoDequeue) Front() interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.cnt <= 0 {
		panic("CoDequeue: Front() is called on empty queue")
	}

	return q.buffer[q.head]
}

// Back 数据从后端读取.
func (q *CoDequeue) Back() interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.cnt <= 0 {
		panic("CoDequeue: Back() is called on empty queue")
	}

	return q.buffer[q.prev(q.tail)]
}

// At 数据从索引idx处读取.
func (q *CoDequeue) At(idx int) interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if idx < 0 || idx >= q.cnt {
		panic("CoDequeue: At() is called with index out of range")
	}

	return q.buffer[(q.head+idx)&(len(q.buffer)-1)]
}

// Set 在索引idx处修改数据.
func (q *CoDequeue) Set(idx int, x interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if idx < 0 || idx >= q.cnt {
		panic("CoDequeue: Set() is called with index out of range")
	}

	q.buffer[(q.head+idx)&(len(q.buffer)-1)] = x
}

// Clear 清空队列.
func (q *CoDequeue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for cur := q.head; cur != q.tail; cur = (cur + 1) & (len(q.buffer) - 1) {
		q.buffer[cur] = nil
	}
	q.head = 0
	q.tail = 0
	q.cnt = 0
}

// Rotate 如果n > 0, 就将前端的n个数据依次放到后端; 如果n < 0, 就将后端的n个数据依次放到前端.
func (q *CoDequeue) Rotate(n int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.cnt <= 1 {
		return
	}

	n %= q.cnt
	if n == 0 {
		return
	}

	if q.head == q.tail {
		q.head = (q.head + n) & (len(q.buffer) - 1)
		q.tail = (q.tail + n) & (len(q.buffer) - 1)
		return
	}

	if n < 0 {
		for ; n < 0; n++ {
			q.head = q.prev(q.head)
			q.tail = q.prev(q.tail)
			q.buffer[q.head] = q.buffer[q.tail]
			q.buffer[q.tail] = nil
		}
	} else {
		for ; n > 0; n-- {
			q.buffer[q.tail] = q.buffer[q.head]
			q.buffer[q.head] = nil
			q.head = q.next(q.head)
			q.tail = q.next(q.tail)
		}
	}
}

func (q *CoDequeue) prev(idx int) int {
	return (idx - 1) & (len(q.buffer) - 1)
}

func (q *CoDequeue) next(idx int) int {
	return (idx + 1) & (len(q.buffer) - 1)
}

func (q *CoDequeue) growIfNeeded() {
	// buffer size == queue len
	if len(q.buffer) == q.cnt {
		// buffer szie grows to 2 * queue len
		q.resize()
	}
}

func (q *CoDequeue) shrinkIfNeeded() {
	// buffer size == 4 * queue len
	if len(q.buffer) > _MinCap && len(q.buffer) == (q.cnt<<2) {
		// buffer szie shrinks to 2 * queue len
		q.resize()
	}
}

func (q *CoDequeue) resize() {
	newBuffer := make([]interface{}, q.cnt<<1)

	if q.tail > q.head {
		copy(newBuffer, q.buffer[q.head:q.tail])
	} else {
		n := copy(newBuffer, q.buffer[q.head:])
		copy(newBuffer[n:], q.buffer[:q.tail])
	}

	q.buffer = newBuffer
	q.head = 0
	q.tail = q.cnt
}
