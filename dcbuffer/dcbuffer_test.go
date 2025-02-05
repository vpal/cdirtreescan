package dcbuffer

import (
	"testing"
)

func TestDynamicCircularBuffer(t *testing.T) {
	buffer := NewDynamicCircularBuffer[int](2)

	// Test Enqueue and Peek
	buffer.Enqueue(1)
	if value, _ := buffer.Peek(); value != 1 {
		t.Errorf("Expected 1, got %v", value)
	}

	// Test Enqueue to full capacity and create new buffer
	buffer.Enqueue(2)
	buffer.Enqueue(3)
	if len(buffer.buffers) != 2 {
		t.Errorf("Expected 2 buffers, got %d", len(buffer.buffers))
	}

	// Test Dequeue
	if value, _ := buffer.Dequeue(); value != 1 {
		t.Errorf("Expected 1, got %v", value)
	}
	if value, _ := buffer.Dequeue(); value != 2 {
		t.Errorf("Expected 2, got %v", value)
	}

	// Test Dequeue from new buffer
	if value, _ := buffer.Dequeue(); value != 3 {
		t.Errorf("Expected 3, got %v", value)
	}

	// Test Dequeue from empty buffer
	if _, ok := buffer.Dequeue(); ok {
		t.Error("Expected false, got true")
	}

	// Test Peek from empty buffer
	if _, ok := buffer.Peek(); ok {
		t.Error("Expected false, got true")
	}
}

func TestDynamicCircularBuffer_Empty(t *testing.T) {
	buffer := NewDynamicCircularBuffer[int](2)

	// Test Empty on new buffer
	if !buffer.Empty() {
		t.Error("Expected true, got false")
	}

	// Test Empty after Enqueue
	buffer.Enqueue(1)
	if buffer.Empty() {
		t.Error("Expected false, got true")
	}

	// Test Empty after Dequeue
	buffer.Dequeue()
	if !buffer.Empty() {
		t.Error("Expected true, got false")
	}

	// Test buffer retention after being emptied
	buffer.Enqueue(2)
	buffer.Dequeue()
	if !buffer.Empty() {
		t.Error("Expected true, got false")
	}
	if len(buffer.buffers) != 1 {
		t.Errorf("Expected 1 buffer, got %d", len(buffer.buffers))
	}
}
