package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// lockInfo 锁信息
type lockInfo struct {
	expireAt time.Time
	owner    string // 获取锁时生成的属主 token
}

// MemoryLock 本地内存锁实现
//
// 属主语义：同一实例内先后多次获取同一把锁时按 FIFO 跟踪——Down 只释放
// 最早尚未释放的那次获取（匹配 defer Down 的调用模式），TTL 过期后被他人
// 抢走的锁不会被先持有者的 Down 误删。
type MemoryLock struct {
	mu     sync.Mutex
	locks  map[string]*lockInfo
	mine   map[string][]string // key → 本实例未释放的属主 token，按获取顺序
	stopCh chan struct{}
}

// NewMemoryLock 创建内存锁实例
func NewMemoryLock() *MemoryLock {
	m := &MemoryLock{
		locks:  make(map[string]*lockInfo),
		mine:   make(map[string][]string),
		stopCh: make(chan struct{}),
	}
	// 启动后台清理过期锁的 goroutine
	go m.cleanupExpiredLocks()
	return m
}

func newToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("t%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// cleanupExpiredLocks 定期清理过期的锁
func (m *MemoryLock) cleanupExpiredLocks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, info := range m.locks {
				if now.After(info.expireAt) {
					delete(m.locks, key)
				}
			}
			m.mu.Unlock()
		case <-m.stopCh:
			return
		}
	}
}

// Up 尝试获取锁（非阻塞）
func (m *MemoryLock) Up(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// 检查锁是否存在且未过期
	if info, exists := m.locks[key]; exists {
		if now.Before(info.expireAt) {
			return false, nil // 锁被占用
		}
		// 锁已过期，可以获取
	}

	token := newToken()
	m.locks[key] = &lockInfo{
		expireAt: now.Add(ttl),
		owner:    token,
	}
	m.mine[key] = append(m.mine[key], token)
	return true, nil
}

// Down 释放锁（仅释放本实例最早未释放的那次获取）
func (m *MemoryLock) Down(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue := m.mine[key]
	if len(queue) == 0 {
		return nil
	}
	token := queue[0]
	m.mine[key] = queue[1:]
	if len(m.mine[key]) == 0 {
		delete(m.mine, key)
	}
	if info, exists := m.locks[key]; exists && info.owner == token {
		delete(m.locks, key)
	}
	return nil
}

// Refresh 续期（仅当锁仍由本实例持有时生效；以最新一次获取为准）
func (m *MemoryLock) Refresh(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue := m.mine[key]
	if len(queue) == 0 {
		return false, nil
	}
	info, exists := m.locks[key]
	if !exists || info.owner != queue[len(queue)-1] {
		return false, nil
	}
	info.expireAt = time.Now().Add(ttl)
	return true, nil
}

// UpWait 等待获取锁（阻塞）
func (m *MemoryLock) UpWait(ctx context.Context, key string, ttl time.Duration, waitTimeout time.Duration) error {
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
		success, err := m.Up(ctx, key, ttl)
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

// Close 关闭锁服务
func (m *MemoryLock) Close() error {
	close(m.stopCh)
	return nil
}
