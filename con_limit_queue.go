package concurrentq

import (
	"sync"

	"github.com/gammazero/deque"
)

// ConLimitQueue is a kind of thread-safe cap-limit queue
type ConLimitQueue struct {
	cond *sync.Cond
	q    deque.Deque
	cap  int
}

// NewConLimitQueue returns a new ConLimitQueue instance.
func NewConLimitQueue(cap int) *ConLimitQueue {
	if cap == 0 {
		cap = 128
	}
	q := &ConLimitQueue{
		cap: cap,
	}
	q.cond = sync.NewCond(&sync.Mutex{})
	return q
}

// Push adds elements into the queue.
func (q *ConLimitQueue) Push(x interface{}) {
	q.cond.L.Lock()
	for q.q.Len() >= q.cap {
		// P1: queue is full now, wait for consumers to pop elements.
		q.cond.Wait()
	}
	q.q.PushBack(x)
	q.cond.L.Unlock()
	// P2 -> P3: tell consumers that there is elements enqueued.
	q.cond.Broadcast()
}

// Pop gets N elements from the queue.
func (q *ConLimitQueue) NPop(want int) []interface{} {
	q.cond.L.Lock()
	for q.q.Len() == 0 {
		// P3: queue is empty now, wait for producers to push elements.
		q.cond.Wait()
	}
	if want > q.q.Len() {
		want = q.q.Len()
	}
	elems := make([]interface{}, want)
	for i := 0; i < want; i++ {
		elems[i] = q.q.PopFront()
	}
	q.cond.L.Unlock()

	// P4 -> P1: tell producers that there is elements dequeued.
	q.cond.Broadcast()
	return elems
}

// Pop gets one element from the queue.
func (q *ConLimitQueue) Pop() interface{} {
	q.cond.L.Lock()
	for q.q.Len() == 0 {
		// P3: queue is empty now, wait for producers to push elements.
		q.cond.Wait()
	}
	output := q.q.PopFront()
	q.cond.L.Unlock()

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
