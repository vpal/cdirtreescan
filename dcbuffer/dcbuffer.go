package dcbuffer

import (
	"github.com/emirpasic/gods/v2/queues/circularbuffer"
)

type DynamicCircularBuffer[T comparable] struct {
	buffers  []*circularbuffer.Queue[T]
	capacity int
}

func NewDynamicCircularBuffer[T comparable](capacity int) *DynamicCircularBuffer[T] {
	return &DynamicCircularBuffer[T]{
		buffers:  []*circularbuffer.Queue[T]{circularbuffer.New[T](capacity)},
		capacity: capacity,
	}
}

func (dcb *DynamicCircularBuffer[T]) Enqueue(value T) {
	lastBuffer := dcb.buffers[len(dcb.buffers)-1]
	if lastBuffer.Full() {
		newBuffer := circularbuffer.New[T](dcb.capacity)
		newBuffer.Enqueue(value)
		dcb.buffers = append(dcb.buffers, newBuffer)
	} else {
		lastBuffer.Enqueue(value)
	}
}

func (dcb *DynamicCircularBuffer[T]) Dequeue() (T, bool) {
	firstBuffer := dcb.buffers[0]
	value, ok := firstBuffer.Dequeue()
	if !ok {
		return value, ok
	}

	if firstBuffer.Size() == 0 && len(dcb.buffers) > 1 {
		dcb.buffers = dcb.buffers[1:]
	}

	return value, ok
}

func (dcb *DynamicCircularBuffer[T]) Peek() (T, bool) {
	firstBuffer := dcb.buffers[0]
	return firstBuffer.Peek()
}

func (dcb *DynamicCircularBuffer[T]) Empty() bool {
	return len(dcb.buffers) == 1 && dcb.buffers[0].Size() == 0
}
