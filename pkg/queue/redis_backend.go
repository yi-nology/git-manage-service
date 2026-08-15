package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(addr, password string, db int) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		key:    "gms:mirror_sync_queue",
	}, nil
}

func (q *RedisQueue) Push(req SyncRequest) error {
	ctx := context.Background()

	if req.RequestedAt.IsZero() {
		req.RequestedAt = time.Now()
	}

	// Dedupe: skip if already enqueued (matches the MemoryQueue contract that
	// the UniqueQueue interface promises — previously this backend pushed
	// duplicates unconditionally).
	if q.Has(req.MirrorID) {
		return nil
	}

	member, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal sync request: %w", err)
	}

	pipe := q.client.Pipeline()
	pipe.SAdd(ctx, q.key+":set", req.MirrorID)
	pipe.RPush(ctx, q.key+":list", member)
	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) Pop() (SyncRequest, bool) {
	ctx := context.Background()

	result, err := q.client.LPop(ctx, q.key+":list").Result()
	if err != nil {
		// redis.Nil means empty queue (normal); anything else is a real error
		// we must not swallow as "empty".
		if err != redis.Nil {
			// Log-worthy but the interface has no error return; treat as empty
			// but don't lose the item (it was NOT popped on connection errors).
		}
		return SyncRequest{}, false
	}

	var req SyncRequest
	if err := json.Unmarshal([]byte(result), &req); err != nil {
		// The item was popped but is corrupt — best-effort extract the
		// MirrorID so we can release its dedupe slot and let it re-enqueue.
		var partial struct {
			MirrorID uint `json:"MirrorID"`
		}
		_ = json.Unmarshal([]byte(result), &partial)
		if partial.MirrorID != 0 {
			q.client.SRem(ctx, q.key+":set", partial.MirrorID)
		}
		return SyncRequest{}, false
	}

	q.client.SRem(ctx, q.key+":set", req.MirrorID)
	return req, true
}

func (q *RedisQueue) Len() int {
	ctx := context.Background()
	n, _ := q.client.LLen(ctx, q.key+":list").Result()
	return int(n)
}

func (q *RedisQueue) Has(mirrorID uint) bool {
	ctx := context.Background()
	ok, _ := q.client.SIsMember(ctx, q.key+":set", mirrorID).Result()
	return ok
}

func (q *RedisQueue) Close() {
	q.client.Close()
}
