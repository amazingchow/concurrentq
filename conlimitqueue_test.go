package conqueue

import (
	"fmt"
	"sync"
	"testing"
)

func TestConLimitQueue(t *testing.T) {
	q := NewConLimitQueue(128)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			q.Push(fmt.Sprintf("data-%06d", i))
		}
	}()

	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q.Pop(1)
		}()
	}

	wg.Wait()
}
