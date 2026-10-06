package queue

import (
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisQueue(t *testing.T) (*RedisQueue, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { client.Close() })
	return &RedisQueue{client: client, key: redisQueueKey}, mr
}

func TestRedisQueue_PushDedupes(t *testing.T) {
	q, _ := newTestRedisQueue(t)

	if err := q.Push(SyncRequest{MirrorID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(SyncRequest{MirrorID: 1}); err != nil {
		t.Fatal(err)
	}
	if got := q.Len(); got != 1 {
		t.Fatalf("expected 1 item after duplicate push, got %d", got)
	}
	if !q.Has(1) {
		t.Fatal("expected mirror 1 to be marked as queued")
	}
}

func TestRedisQueue_PushConcurrentSingleCopy(t *testing.T) {
	q, _ := newTestRedisQueue(t)

	const n = 32
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = q.Push(SyncRequest{MirrorID: 7})
		}()
	}
	wg.Wait()

	if got := q.Len(); got != 1 {
		t.Fatalf("expected exactly 1 item from %d concurrent pushes, got %d", n, got)
	}
}

func TestRedisQueue_PopReleasesSlot(t *testing.T) {
	q, _ := newTestRedisQueue(t)

	if err := q.Push(SyncRequest{MirrorID: 5, TriggerType: "manual"}); err != nil {
		t.Fatal(err)
	}

	req, ok := q.Pop()
	if !ok {
		t.Fatal("expected one item")
	}
	if req.MirrorID != 5 {
		t.Fatalf("unexpected mirror id %d", req.MirrorID)
	}
	if q.Has(5) {
		t.Fatal("slot should be released after pop")
	}

	// The released slot must allow re-enqueueing immediately (this is the
	// exact scenario the old non-atomic pop permanently deadlocked).
	if err := q.Push(SyncRequest{MirrorID: 5}); err != nil {
		t.Fatal(err)
	}
	if got := q.Len(); got != 1 {
		t.Fatalf("expected re-enqueue after pop, len=%d", got)
	}
}

func TestRedisQueue_PopCorruptPayloadReleasesSlot(t *testing.T) {
	q, mr := newTestRedisQueue(t)

	// Inject a payload whose MirrorID parses but whose body is unusable.
	if _, err := mr.Lpush(redisQueueKey+":list", `{"MirrorID":9,"broken`); err != nil {
		t.Fatal(err)
	}
	if _, err := mr.SAdd(redisQueueKey+":set", "9"); err != nil {
		t.Fatal(err)
	}
	if _, ok := q.Pop(); ok {
		t.Fatal("corrupt payload should be dropped, not returned")
	}
	if q.Has(9) {
		t.Fatal("slot should be released for corrupt payload")
	}
	if err := q.Push(SyncRequest{MirrorID: 9}); err != nil {
		t.Fatal(err)
	}
	if got := q.Len(); got != 1 {
		t.Fatalf("mirror 9 must be enqueueable after corrupt pop, len=%d", got)
	}
}

func TestRedisQueue_PopEmpty(t *testing.T) {
	q, _ := newTestRedisQueue(t)
	if _, ok := q.Pop(); ok {
		t.Fatal("expected empty pop")
	}
}

func TestRedisQueue_PopRoundTripsFields(t *testing.T) {
	q, _ := newTestRedisQueue(t)
	if err := q.Push(SyncRequest{MirrorID: 3, TriggerType: "cron"}); err != nil {
		t.Fatal(err)
	}
	req, ok := q.Pop()
	if !ok {
		t.Fatal("expected one item")
	}
	if req.TriggerType != "cron" {
		t.Fatalf("trigger type lost: %+v", req)
	}
	if req.RequestedAt.IsZero() {
		t.Fatal("RequestedAt should be defaulted on push")
	}
}
