package concurrentq

import (
	"sync"
)

const (
	// must be equal to 2^x
	_MinCap = 32
	// must be equal to 2^x
	_MinShrinkThreshold = 65536
)

// ConDequeue is a kind of thread-safe double-end-queue
type ConDequeue struct {
	mu   sync.RWMutex
	buf  []interface{}
	cap  int
	size int
	head int
	tail int
}

// NewConDequeue returns a new ConDequeue instance.
func NewConDequeue(cap int) *ConDequeue {
	if cap < _MinCap {
		cap = _MinCap
	} else {
		cap = nextTwoPower(cap)
	}

	q := &ConDequeue{}
	q.buf = make([]interface{}, cap)
	q.cap = cap
	q.size = 0
	q.head = 0
	q.tail = 0
	return q
}

func nextTwoPower(n int) int {
	m := 32
	for m < n {
		m <<= 1
	}
	return m
}

// Len gives the length of the double-end-queue.
func (q *ConDequeue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size
}

// FPush enqueues one element from front end.
func (q *ConDequeue) FPush(x interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.growIfNeeded()

	q.head = q.prev(q.head)
	q.buf[q.head] = x
	q.size++
}

// FPop dequeues one element from front end.
func (q *ConDequeue) FPop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size <= 0 {
		panic("ConDequeue: FPop() called on empty queue")
	}

	x := q.buf[q.head]
	q.buf[q.head] = nil
	q.head = q.next(q.head)
	q.size--

	q.shrinkIfNeeded()

	return x
}

// BPush enqueues one element from back end.
func (q *ConDequeue) BPush(x interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.growIfNeeded()

	q.buf[q.tail] = x
	q.tail = q.next(q.tail)
	q.size++
}

// BPop dequeues one element from back end.
func (q *ConDequeue) BPop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size <= 0 {
		panic("ConDequeue: BPop() called on empty queue")
	}

	q.tail = q.prev(q.tail)
	x := q.buf[q.tail]
	q.buf[q.tail] = nil
	q.size--

	q.shrinkIfNeeded()

	return x
}

// Front reads one element from front end.
func (q *ConDequeue) Front() interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size <= 0 {
		panic("ConDequeue: Front() called on empty queue")
	}

	return q.buf[q.head]
}

// Back reads one element from back end.
func (q *ConDequeue) Back() interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size <= 0 {
		panic("ConDequeue: Back() called on empty queue")
	}

	return q.buf[q.prev(q.tail)]
}

// At reads one element at index idx.
func (q *ConDequeue) At(idx int) interface{} {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if idx < 0 || idx >= q.size {
		panic("ConDequeue: At() called with index out of range")
	}

	return q.buf[(q.head+idx)&(q.cap-1)]
}

// Set sets one element at index idx.
func (q *ConDequeue) Set(idx int, x interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if idx < 0 || idx >= q.size {
		panic("ConDequeue: Set() called with index out of range")
	}

	q.buf[(q.head+idx)&(q.cap-1)] = x
}

// Clear clears all elements.
func (q *ConDequeue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for cur := q.head; cur != q.tail; cur = (cur + 1) & (q.cap - 1) {
		q.buf[cur] = nil
	}
	q.head = 0
	q.tail = 0
	q.size = 0
}

// Rotate 如果n > 0, 就将前端的n个数据依次放到后端; 如果n < 0, 就将后端的n个数据依次放到前端.
func (q *ConDequeue) Rotate(n int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size <= 1 {
		return
	}

	n %= q.size
	if n == 0 {
		return
	}

	if q.head == q.tail {
		q.head = (q.head + n) & (q.cap - 1)
		q.tail = (q.tail + n) & (q.cap - 1)
		return
	}

	if n < 0 {
		for ; n < 0; n++ {
			q.head = q.prev(q.head)
			q.tail = q.prev(q.tail)
			q.buf[q.head] = q.buf[q.tail]
			q.buf[q.tail] = nil
		}
	} else {
		for ; n > 0; n-- {
			q.buf[q.tail] = q.buf[q.head]
			q.buf[q.head] = nil
			q.head = q.next(q.head)
			q.tail = q.next(q.tail)
		}
	}
}

func (q *ConDequeue) prev(idx int) int {
	// since (-1 & (2^x - 1)) == (2^x - 1)
	return (idx - 1) & (q.cap - 1)
}

func (q *ConDequeue) next(idx int) int {
	return (idx + 1) & (q.cap - 1)
}

func (q *ConDequeue) growIfNeeded() {
	if q.size == q.cap {
		// buffer size gets double growth
		q.resize(false)
	}
}

func (q *ConDequeue) shrinkIfNeeded() {
	// buffer size equals to 4 * queue size
	// TODO: consider the case that int-type element overflows
	if q.cap >= _MinShrinkThreshold && q.cap >= (q.size<<2) {
		// buffer size shrinks to 1/2
		q.resize(true)
	}
}

func (q *ConDequeue) resize(shrink bool) {
	if shrink {
		if q.cap == _MinCap {
			return
		}
		q.cap >>= 1
	} else {
		q.cap <<= 1
	}

	newBuffer := make([]interface{}, q.cap)
	if q.tail > q.head {
		copy(newBuffer, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuffer, q.buf[q.head:])
		copy(newBuffer[n:], q.buf[:q.tail])
	}

	q.buf = newBuffer
	q.head = 0
	q.tail = q.size
}
