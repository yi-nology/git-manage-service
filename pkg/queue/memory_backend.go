package queue

import (
	"sync"
	"time"
)

type MemoryQueue struct {
	mu      sync.Mutex
	items   []SyncRequest
	present map[uint]bool
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		items:   make([]SyncRequest, 0),
		present: make(map[uint]bool),
	}
}

func (q *MemoryQueue) Push(req SyncRequest) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.present[req.MirrorID] {
		return nil
	}

	if req.RequestedAt.IsZero() {
		req.RequestedAt = time.Now()
	}

	q.items = append(q.items, req)
	q.present[req.MirrorID] = true
	return nil
}

func (q *MemoryQueue) Pop() (SyncRequest, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return SyncRequest{}, false
	}

	item := q.items[0]
	q.items = q.items[1:]
	delete(q.present, item.MirrorID)
	return item, true
}

func (q *MemoryQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

func (q *MemoryQueue) Has(mirrorID uint) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.present[mirrorID]
}

func (q *MemoryQueue) Close() {}
