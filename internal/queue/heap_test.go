package queue

import (
	"container/heap"
	"testing"
)

func TestMaxHeap(t *testing.T) {
	h := &MaxHeap{}
	heap.Init(h)

	heap.Push(h, &Job{ID: "job1", Priority: 5})
	heap.Push(h, &Job{ID: "job2", Priority: 20})
	heap.Push(h, &Job{ID: "job3", Priority: 10})

	job1 := heap.Pop(h).(*Job)
	if job1.Priority != 20 {
		t.Errorf("expected priority 20, got %d", job1.Priority)
	}

	job2 := heap.Pop(h).(*Job)
	if job2.Priority != 10 {
		t.Errorf("expected priority 10, got %d", job2.Priority)
	}

	job3 := heap.Pop(h).(*Job)
	if job3.Priority != 5 {
		t.Errorf("expected priority 5, got %d", job3.Priority)
	}
}
