package queue

import (
	"container/heap"
	"sync"
)

type Queue struct {
	mu   sync.Mutex
	cond *sync.Cond
	jobs *MaxHeap
}

func NewQueue() *Queue {
	h := &MaxHeap{}
	heap.Init(h)
	q := &Queue{jobs: h}
	q.cond = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Push(job *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(q.jobs, job)
	q.cond.Signal()
}

func (q *Queue) Pop() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.jobs.Len() == 0 {
		q.cond.Wait()
	}
	return heap.Pop(q.jobs).(*Job)
}

func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.jobs.Len()
}
