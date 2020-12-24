package conqueue

import (
	"sync"

	"github.com/gammazero/deque"
)

// ConLimitQueue is a kind of thread-safe cap-limit queue
type ConLimitQueue struct {
	cond *sync.Cond

	q   deque.Deque
	cap int
}

// NewConLimitQueue returns a new ConLimitQueue instance.
func NewConLimitQueue(cap int) *ConLimitQueue {
	if cap == 0 {
		cap = 64
	}

	q := &ConLimitQueue{
		cap: cap,
	}

	q.cond = sync.NewCond(&sync.Mutex{})

	return q
}

// Push adds data into the queue.
func (q *ConLimitQueue) Push(x interface{}) {
	q.cond.L.Lock()
	for q.q.Len() >= q.cap {
		// P1: queue is full now, wait for consumers to pop data.
		q.cond.Wait()
	}
	defer q.cond.L.Unlock()

	q.q.PushBack(x)
	// P2 -> P3: tell consumers that there is data enqueued.
	q.cond.Broadcast()
}

// Pop gets data from the queue.
func (q *ConLimitQueue) Pop(want int) []interface{} {
	q.cond.L.Lock()
	for q.q.Len() == 0 {
		// P3: queue is empty now, wait for producers to push data.
		q.cond.Wait()
	}
	defer q.cond.L.Unlock()

	if want > q.q.Len() {
		want = q.q.Len()
	}

	output := make([]interface{}, want)
	for i := 0; i < want; i++ {
		output[i] = q.q.PopFront()
	}
	// P4 -> P1: tell producers that there is data dequeued.
	q.cond.Broadcast()

	return output
}

// Len gets the length of the queue.
func (q *ConLimitQueue) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.q.Len()
}
