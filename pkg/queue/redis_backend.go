package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/redis/go-redis/v9"
)

// Key names carry a v2 suffix: the pre-v2 implementation released the dedupe
// slot in a separate SRem after LPop, so a crash in between left orphaned set
// members that permanently blocked that mirror from ever re-enqueueing. The
// new namespace starts every deployment from a clean slate.
const redisQueueKey = "gms:mirror_sync_queue:v2"

// pushScript enqueues atomically: the SISMEMBER dedupe check and the SADD+RPUSH
// happen in one script, so two concurrent Push calls can never both pass the
// check and enqueue the same mirror twice.
var pushScript = redis.NewScript(`
if redis.call('SISMEMBER', KEYS[1], ARGV[1]) == 1 then
	return 0
end
redis.call('SADD', KEYS[1], ARGV[1])
redis.call('RPUSH', KEYS[2], ARGV[2])
return 1
`)

// popScript pops atomically: LPOP and the slot release (SREM) are one unit, so
// a crash can no longer leave the slot behind with the item already consumed.
// If the payload is corrupt, the best-effort MirrorID still releases the slot
// so the mirror can re-enqueue on its next trigger.
var popScript = redis.NewScript(`
local v = redis.call('LPOP', KEYS[1])
if not v then
	return false
end
local ok, obj = pcall(cjson.decode, v)
if ok and type(obj) == 'table' and obj["MirrorID"] then
	redis.call('SREM', KEYS[2], obj["MirrorID"])
end
return v
`)

var corruptMirrorIDRe = regexp.MustCompile(`"MirrorID"\s*:\s*(\d+)`)

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
		key:    redisQueueKey,
	}, nil
}

func (q *RedisQueue) Push(req SyncRequest) error {
	ctx := context.Background()

	if req.RequestedAt.IsZero() {
		req.RequestedAt = time.Now()
	}

	member, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal sync request: %w", err)
	}

	_, err = pushScript.Run(ctx, q.client,
		[]string{q.key + ":set", q.key + ":list"},
		req.MirrorID, member,
	).Int64()
	if err != nil {
		return fmt.Errorf("enqueue sync request: %w", err)
	}
	return nil
}

func (q *RedisQueue) Pop() (SyncRequest, bool) {
	ctx := context.Background()

	result, err := popScript.Run(ctx, q.client,
		[]string{q.key + ":list", q.key + ":set"},
	).Text()
	if err != nil {
		// redis.Nil means empty queue (normal); anything else is a real error
		// we must not swallow as "empty" — the item was NOT popped, so the
		// next tick retries it.
		if err != redis.Nil {
			log.Printf("[queue] pop sync request failed (will retry): %v", err)
		}
		return SyncRequest{}, false
	}

	var req SyncRequest
	if err := json.Unmarshal([]byte(result), &req); err != nil {
		// The script releases the slot only when cjson can parse the payload;
		// syntactically broken JSON lands here, so extract the MirrorID
		// best-effort and release the slot anyway. A dropped item must never
		// leave its mirror permanently stuck.
		if m := corruptMirrorIDRe.FindStringSubmatch(result); m != nil {
			q.client.SRem(ctx, q.key+":set", m[1])
		}
		log.Printf("[queue] dropping corrupt sync request payload: %v", err)
		return SyncRequest{}, false
	}
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
