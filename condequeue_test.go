package condequeue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConDequeue(t *testing.T) {
	q := NewConDequeue(500)

	q.FPush("zhangshan@gmail.com")
	q.BPush("lisi@gmail.com")
	assert.Equal(t, 2, q.Len())
	assert.Equal(t, "zhangshan@gmail.com", q.Front())
	assert.Equal(t, "lisi@gmail.com", q.Back())

	q.FPop() // remove "zhangshan@gmail.com"
	q.BPop() // remove "lisi@gmail.com"

	q.FPush("hello")
	q.BPush("world")
	assert.Equal(t, 2, q.Len())
	assert.Equal(t, "hello", q.At(0))
	assert.Equal(t, "world", q.At(1))
	q.Set(1, "golang")
	assert.Equal(t, "golang", q.At(1))

	q.Clear()
	assert.Equal(t, 0, q.Len())
}
