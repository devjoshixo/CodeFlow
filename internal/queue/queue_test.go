package queue

import "testing"

func TestQueueConcurrent(t *testing.T) {
	q := NewQueue()
	results := make(chan *Job, 2)

	// Spawn 2 workers
	go func() {
		job := q.Pop()
		results <- job
	}()
	go func() {
		job := q.Pop()
		results <- job
	}()

	// Push 2 jobs (high priority first, then low)
	q.Push(&Job{ID: "job_low", Priority: 5})
	q.Push(&Job{ID: "job_high", Priority: 20})

	// Collect results
	job1 := <-results
	job2 := <-results

	// Verify priority order (high priority should be popped first)
	if job1.Priority < job2.Priority {
		t.Errorf("expected high priority first, got %d then %d", job1.Priority, job2.Priority)
	}
}
