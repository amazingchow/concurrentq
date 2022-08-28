package concurrentq

import (
	"fmt"
	"sync"
	"testing"
)

func TestConLimitQueue(t *testing.T) {
	q := NewConLimitQueue(1024)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 512; i++ {
			q.Push(fmt.Sprintf("data-%06d", i))
		}
	}()

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Pop()
		}()
	}

	wg.Wait()
}
