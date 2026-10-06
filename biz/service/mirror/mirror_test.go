package mirror

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/queue"
	"github.com/yi-nology/go-git-platform/gitbackend"
)

func TestMirrorStatus_ValidTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{po.MirrorStatusActive, po.MirrorStatusSyncing, true},
		{po.MirrorStatusActive, po.MirrorStatusPaused, true},
		{po.MirrorStatusActive, po.MirrorStatusFailed, false},
		{po.MirrorStatusSyncing, po.MirrorStatusActive, true},
		{po.MirrorStatusSyncing, po.MirrorStatusFailed, true},
		{po.MirrorStatusSyncing, po.MirrorStatusPaused, false},
		{po.MirrorStatusFailed, po.MirrorStatusSyncing, true},
		{po.MirrorStatusFailed, po.MirrorStatusActive, true},
		{po.MirrorStatusFailed, po.MirrorStatusPaused, true},
		{po.MirrorStatusPaused, po.MirrorStatusActive, true},
		{po.MirrorStatusPaused, po.MirrorStatusSyncing, false},
		{"bogus", po.MirrorStatusActive, false},
	}
	for _, c := range cases {
		s := NewMirrorStatus(c.from)
		if got := s.CanTransitionTo(c.to); got != c.want {
			t.Errorf("%s -> %s: CanTransitionTo=%v, want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestMirrorStatus_TransitionToMutates(t *testing.T) {
	s := NewMirrorStatus(po.MirrorStatusActive)
	if err := s.TransitionTo(po.MirrorStatusSyncing); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Current != po.MirrorStatusSyncing {
		t.Fatalf("current = %s, want syncing", s.Current)
	}
	if err := s.TransitionTo(po.MirrorStatusPaused); err == nil {
		t.Fatal("syncing -> paused must be rejected")
	}
	if s.Current != po.MirrorStatusSyncing {
		t.Fatal("rejected transition must not mutate state")
	}
}

func TestCanStartSync(t *testing.T) {
	if !CanStartSync(po.MirrorStatusActive) || !CanStartSync(po.MirrorStatusFailed) {
		t.Error("active and failed should be syncable")
	}
	if CanStartSync(po.MirrorStatusSyncing) || CanStartSync(po.MirrorStatusPaused) {
		t.Error("syncing/paused must not start sync")
	}
}

func TestRetryStrategy_DefaultMaxRetry(t *testing.T) {
	if r := NewRetryStrategy(0); r.MaxRetry != 5 {
		t.Fatalf("non-positive maxRetry should default to 5, got %d", r.MaxRetry)
	}
	if r := NewRetryStrategy(3); r.MaxRetry != 3 {
		t.Fatalf("MaxRetry = %d, want 3", r.MaxRetry)
	}
}

func TestRetryStrategy_BackoffLadder(t *testing.T) {
	r := NewRetryStrategy(10)
	cases := map[int]time.Duration{
		1: 1 * time.Minute,
		2: 5 * time.Minute,
		3: 30 * time.Minute,
		4: 2 * time.Hour,
		5: 32 * time.Hour,
		6: 64 * time.Hour,
	}
	for count, want := range cases {
		if got := r.GetNextRetryDelay(count); got != want {
			t.Errorf("delay(%d) = %s, want %s", count, got, want)
		}
	}
	// 大 retryCount 封顶 7 天且不得溢出为负
	if got := r.GetNextRetryDelay(100); got != 168*time.Hour || got < 0 {
		t.Fatalf("delay(100) = %s, want capped 168h and positive", got)
	}
}

func TestRetryStrategy_ShouldPause(t *testing.T) {
	r := NewRetryStrategy(3)
	if r.ShouldPause(2) {
		t.Error("below max must not pause")
	}
	if !r.ShouldPause(3) {
		t.Error("at max must pause")
	}
	if !r.ShouldPause(4) {
		t.Error("beyond max must pause")
	}
}

func TestResolveAuth(t *testing.T) {
	if got := resolveAuth(&po.Mirror{}); got.Type != gitbackend.AuthNone {
		t.Errorf("no credential should yield AuthNone, got %v", got.Type)
	}
	cases := []struct {
		credType string
		want     gitbackend.AuthType
	}{
		{"ssh_key", gitbackend.AuthSSH},
		{"http_basic", gitbackend.AuthHTTPBasic},
		{"http_token", gitbackend.AuthHTTPToken},
		{"unknown", gitbackend.AuthNone},
	}
	for _, c := range cases {
		m := &po.Mirror{Credential: &po.Credential{Type: c.credType, Secret: "s", Username: "u", SSHKeyPath: "/k"}}
		if got := resolveAuth(m); got.Type != c.want {
			t.Errorf("cred type %s: got %v, want %v", c.credType, got.Type, c.want)
		}
	}
}

func TestWorkerPool_StopDrains(t *testing.T) {
	db.DB = nil // 不触库：handler 只计数
	q := queue.NewMemoryQueue()
	var calls atomic.Int32
	done := make(chan struct{})
	go func() {
		wp := queue.NewWorkerPool(q, func(req queue.SyncRequest) { calls.Add(1) }, 2)
		wp.Start()
		_ = q.Push(queue.SyncRequest{MirrorID: 1})
		// 等 handler 被调后立刻 Stop：Stop 必须等在跑的 handler 返回
		for calls.Load() == 0 {
			time.Sleep(5 * time.Millisecond)
		}
		wp.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop did not return")
	}
	if calls.Load() == 0 {
		t.Fatal("handler never ran")
	}
}
