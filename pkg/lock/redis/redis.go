package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLock Redis 分布式锁实现
//
// 属主语义：每把锁写入随机属主 token，释放时用 Lua 比对后删除。同一实例内
// 先后多次获取按 FIFO 跟踪——Down 只释放本实例最早未释放的那次获取；跨实例
// 场景下 TTL 过期后被其他实例抢走的锁不会被原持有者的 Down 误删。
type RedisLock struct {
	client *redis.Client
	mu     sync.Mutex
	tokens map[string][]string // key → 本实例未释放的属主 token，按获取顺序
}

// NewRedisLock 创建 Redis 锁实例
func NewRedisLock(addr, password string, db int) (*RedisLock, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisLock{client: client, tokens: make(map[string][]string)}, nil
}

func newToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand 失败基本意味着系统熵源不可用；退化到时间戳仍优于固定值。
		return fmt.Sprintf("t%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// releaseScript 仅当值等于本实例的 token 时才删除锁。
var releaseScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('DEL', KEYS[1])
end
return 0
`)

// refreshScript 仅当值等于本实例的 token 时才续期。
var refreshScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`)

// Up 尝试获取锁（非阻塞）
func (r *RedisLock) Up(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	token := newToken()
	success, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}
	if success {
		r.mu.Lock()
		r.tokens[key] = append(r.tokens[key], token)
		r.mu.Unlock()
	}
	return success, nil
}

// Down 释放锁（仅释放本实例最早未释放的那次获取）
func (r *RedisLock) Down(ctx context.Context, key string) error {
	r.mu.Lock()
	queue := r.tokens[key]
	if len(queue) == 0 {
		r.mu.Unlock()
		return nil
	}
	token := queue[0]
	r.tokens[key] = queue[1:]
	if len(r.tokens[key]) == 0 {
		delete(r.tokens, key)
	}
	r.mu.Unlock()

	if err := releaseScript.Run(ctx, r.client, []string{key}, token).Err(); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}

// Refresh 续期（仅当锁仍由本实例持有时生效；以最新一次获取为准）
func (r *RedisLock) Refresh(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	r.mu.Lock()
	queue := r.tokens[key]
	if len(queue) == 0 {
		r.mu.Unlock()
		return false, nil
	}
	token := queue[len(queue)-1]
	r.mu.Unlock()

	n, err := refreshScript.Run(ctx, r.client, []string{key}, token, ttl.Milliseconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("failed to refresh lock: %w", err)
	}
	return n == 1, nil
}

// UpWait 等待获取锁（阻塞）
func (r *RedisLock) UpWait(ctx context.Context, key string, ttl time.Duration, waitTimeout time.Duration) error {
	deadline := time.Time{}
	if waitTimeout > 0 {
		deadline = time.Now().Add(waitTimeout)
	}

	for {
		// 检查 context 是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 尝试获取锁
		success, err := r.Up(ctx, key, ttl)
		if err != nil {
			return err
		}
		if success {
			return nil
		}

		// 检查是否超时
		if !deadline.IsZero() && time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for lock: %s", key)
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// Close 关闭 Redis 连接
func (r *RedisLock) Close() error {
	return r.client.Close()
}
