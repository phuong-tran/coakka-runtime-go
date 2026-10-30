package coakka_v2_connector

import "sync"

const (
	handlerDispatchCapacity  = 64
	handlerDispatchByteLimit = 8 * 1024 * 1024
)

// handlerDispatchQueue owns Go-packed jobs, not Core ingress/terminal policy.
// The sole reader backpressures here; a large valid record runs exclusively.
type handlerDispatchQueue struct {
	jobs     chan deliveredRequestJob
	mu       sync.Mutex
	changed  *sync.Cond
	retained int
	closing  bool
}

func newHandlerDispatchQueue(capacity int) *handlerDispatchQueue {
	if capacity > handlerDispatchCapacity {
		capacity = handlerDispatchCapacity
	}
	queue := &handlerDispatchQueue{jobs: make(chan deliveredRequestJob, capacity)}
	queue.changed = sync.NewCond(&queue.mu)
	return queue
}

func dispatchJobBytes(job deliveredRequestJob) int {
	r := job.request
	return len(r.Frame) + len(r.Target) + 1 + len(r.Payload)
}

func (q *handlerDispatchQueue) enqueue(job deliveredRequestJob, closed <-chan struct{}) bool {
	cost := dispatchJobBytes(job)
	q.mu.Lock()
	for !q.closing && q.retained != 0 && (cost > handlerDispatchByteLimit || q.retained > handlerDispatchByteLimit-cost) {
		q.changed.Wait()
	}
	if q.closing {
		q.mu.Unlock()
		return false
	}
	q.retained += cost
	q.mu.Unlock()
	select {
	case q.jobs <- job:
		return true
	case <-closed:
		q.release(job)
		return false
	}
}

func (q *handlerDispatchQueue) release(job deliveredRequestJob) {
	q.mu.Lock()
	q.retained -= dispatchJobBytes(job)
	q.changed.Signal() // One reader owns admission.
	q.mu.Unlock()
}

func (q *handlerDispatchQueue) stop() {
	q.mu.Lock()
	q.closing = true
	q.changed.Broadcast()
	q.mu.Unlock()
}
