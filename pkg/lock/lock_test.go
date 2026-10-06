package lock

import (
	"context"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	memorylock "github.com/yi-nology/git-manage-service/pkg/lock/memory"
	redislock "github.com/yi-nology/git-manage-service/pkg/lock/redis"
)

// 跨实现的行为契约：属主语义、TTL 过期抢占、Refresh。
//
// 时钟说明：memory 锁按真实时间过期；miniredis 不随真实时间过期 key，
// 通过 fastForward 推进其内部时钟。两个工厂的 expire 都保证「调用返回后
// 此前以短 TTL 获取的锁已过期」。

type lockFactory struct {
	name  string
	make  func(t *testing.T) DistLock
	close func(DistLock)
}

var factories = []lockFactory{
	{
		name:  "memory",
		make:  func(t *testing.T) DistLock { return memorylock.NewMemoryLock() },
		close: func(l DistLock) { _ = l.Close() },
	},
	{
		name: "redis",
		make: func(t *testing.T) DistLock {
			mr := miniredis.RunT(t)
			l, err := redislock.NewRedisLock(mr.Addr(), "", 0)
			if err != nil {
				t.Fatalf("new redis lock: %v", err)
			}
			t.Cleanup(func() { _ = l.Close() })
			return l
		},
	},
}

// expireHelper 返回锁实例和一个「推进时钟使短 TTL 锁过期」的函数：
// memory 按真实时间睡眠；miniredis 推进其内部时钟。
func expireHelper(f lockFactory, t *testing.T) (DistLock, func(time.Duration)) {
	if f.name == "redis" {
		mr := miniredis.RunT(t)
		l, err := redislock.NewRedisLock(mr.Addr(), "", 0)
		if err != nil {
			t.Fatalf("new redis lock: %v", err)
		}
		t.Cleanup(func() { _ = l.Close() })
		return l, func(d time.Duration) { mr.FastForward(d) }
	}
	l := f.make(t)
	if f.close != nil {
		t.Cleanup(func() { f.close(l) })
	}
	return l, func(d time.Duration) { time.Sleep(d + 20*time.Millisecond) }
}

func TestDistLock_UpDown(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			l := f.make(t)
			if f.close != nil {
				defer f.close(l)
			}
			ctx := context.Background()

			ok, err := l.Up(ctx, "k1", time.Minute)
			if err != nil || !ok {
				t.Fatalf("first Up should succeed, got ok=%v err=%v", ok, err)
			}
			ok, _ = l.Up(ctx, "k1", time.Minute)
			if ok {
				t.Fatal("second Up while held must fail")
			}
			if err := l.Down(ctx, "k1"); err != nil {
				t.Fatalf("Down: %v", err)
			}
			ok, _ = l.Up(ctx, "k1", time.Minute)
			if !ok {
				t.Fatal("Up after Down should succeed")
			}
		})
	}
}

// 同一实例内先后两次获取同一把锁：Down 按 FIFO 只释放自己那次获取，
// 不得动到后一次获取的锁（defer Down 的典型模式）。
func TestDistLock_DownFollowsFIFOOwnership(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			l, expire := expireHelper(f, t)
			ctx := context.Background()

			if ok, _ := l.Up(ctx, "k", 20*time.Millisecond); !ok {
				t.Fatal("first acquire failed")
			}
			expire(40 * time.Millisecond) // 第一次获取过期
			if ok, _ := l.Up(ctx, "k", time.Minute); !ok {
				t.Fatal("second acquire after expiry failed")
			}
			// 最早那次获取的 Down：TTL 已过期且锁已被第二次获取占走，不得误删
			if err := l.Down(ctx, "k"); err != nil {
				t.Fatalf("stale Down should not error: %v", err)
			}
			if ok, _ := l.Up(ctx, "k", time.Minute); ok {
				t.Fatal("stale Down released the second acquisition")
			}
			// 第二次获取的 Down 正常释放
			if err := l.Down(ctx, "k"); err != nil {
				t.Fatalf("current Down: %v", err)
			}
			if ok, _ := l.Up(ctx, "k", time.Minute); !ok {
				t.Fatal("Up after current Down should succeed")
			}
		})
	}
}

// Redis 跨实例场景：实例 A 的锁过期后被实例 B 抢走，A 的 Down 不得删除 B 的锁。
func TestRedisLock_StaleInstanceDownDoesNotReleaseNewHolder(t *testing.T) {
	mr := miniredis.RunT(t)
	newLock := func() DistLock {
		l, err := redislock.NewRedisLock(mr.Addr(), "", 0)
		if err != nil {
			t.Fatalf("new redis lock: %v", err)
		}
		t.Cleanup(func() { _ = l.Close() })
		return l
	}
	a, b := newLock(), newLock()
	ctx := context.Background()

	if ok, _ := a.Up(ctx, "k", 20*time.Millisecond); !ok {
		t.Fatal("A acquire failed")
	}
	mr.FastForward(40 * time.Millisecond)
	if ok, _ := b.Up(ctx, "k", time.Minute); !ok {
		t.Fatal("B acquire after expiry failed")
	}
	if err := a.Down(ctx, "k"); err != nil {
		t.Fatalf("stale A.Down should not error: %v", err)
	}
	if ok, _ := b.Up(ctx, "k", time.Minute); ok {
		t.Fatal("A.Down released B's lock")
	}
	if err := b.Down(ctx, "k"); err != nil {
		t.Fatalf("B.Down: %v", err)
	}
	if ok, _ := a.Up(ctx, "k", time.Minute); !ok {
		t.Fatal("Up after B.Down should succeed")
	}
}

func TestDistLock_Refresh(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			l, expire := expireHelper(f, t)
			ctx := context.Background()

			if ok, _ := l.Up(ctx, "k", 30*time.Millisecond); !ok {
				t.Fatal("acquire failed")
			}
			if ok, err := l.Refresh(ctx, "k", time.Minute); err != nil || !ok {
				t.Fatalf("Refresh while held should succeed, got ok=%v err=%v", ok, err)
			}
			// 刷新前的 TTL（30ms）已过：若 Refresh 未生效，锁已丢失
			expire(50 * time.Millisecond)
			if ok, _ := l.Up(ctx, "k", time.Minute); ok {
				t.Fatal("lock expired despite Refresh")
			}
			if err := l.Down(ctx, "k"); err != nil {
				t.Fatalf("Down after renew: %v", err)
			}
			// 未持有的锁 Refresh 返回 false 而非报错
			if ok, err := l.Refresh(ctx, "k", time.Minute); ok || err != nil {
				t.Fatalf("Refresh on unheld lock should be (false, nil), got (%v, %v)", ok, err)
			}
		})
	}
}

func TestDistLock_UpWait(t *testing.T) {
	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			l, expire := expireHelper(f, t)
			ctx := context.Background()

			if ok, _ := l.Up(ctx, "k", 150*time.Millisecond); !ok {
				t.Fatal("acquire failed")
			}
			if err := l.UpWait(ctx, "k", time.Minute, 50*time.Millisecond); err == nil {
				t.Fatal("UpWait should time out while lock is held")
			}
			expire(150 * time.Millisecond)
			if err := l.UpWait(ctx, "k", time.Minute, 2*time.Second); err != nil {
				t.Fatalf("UpWait should succeed after expiry: %v", err)
			}
		})
	}
}

func TestMemoryLock_ConcurrentSingleWinner(t *testing.T) {
	l := memorylock.NewMemoryLock()
	defer func() { _ = l.Close() }()
	ctx := context.Background()

	const n = 32
	var mu sync.Mutex
	winners := 0
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if ok, _ := l.Up(ctx, "contended", time.Minute); ok {
				mu.Lock()
				winners++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if winners != 1 {
		t.Fatalf("expected exactly 1 winner among %d, got %d", n, winners)
	}
}
