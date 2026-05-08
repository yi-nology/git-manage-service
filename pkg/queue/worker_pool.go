package queue

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/yi-nology/git-manage-service/pkg/configs"
)

type WorkerPool struct {
	maxWorkers  int32
	queue       UniqueQueue
	handler     func(SyncRequest)
	activeCount int32
	wg          sync.WaitGroup
	stop        chan struct{}
}

func NewWorkerPool(queue UniqueQueue, handler func(SyncRequest), maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	return &WorkerPool{
		maxWorkers: int32(maxWorkers),
		queue:      queue,
		handler:    handler,
		stop:       make(chan struct{}),
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < int(p.maxWorkers); i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *WorkerPool) Stop() {
	close(p.stop)
	p.wg.Wait()
}

func (p *WorkerPool) ActiveWorkers() int32 {
	return atomic.LoadInt32(&p.activeCount)
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stop:
			return
		default:
		}

		req, ok := p.queue.Pop()
		if !ok {
			return
		}

		atomic.AddInt32(&p.activeCount, 1)
		p.handler(req)
		atomic.AddInt32(&p.activeCount, -1)
	}
}

func NewQueueFromConfig(cfg configs.MirrorConfig, redisCfg configs.RedisConfig) (UniqueQueue, error) {
	switch cfg.QueueBackend {
	case "redis":
		if redisCfg.Addr == "" {
			return nil, fmt.Errorf("redis addr is required for redis queue backend")
		}
		return NewRedisQueue(redisCfg.Addr, redisCfg.Password, redisCfg.DB)
	case "memory", "":
		return NewMemoryQueue(), nil
	default:
		return nil, fmt.Errorf("unsupported queue backend: %s", cfg.QueueBackend)
	}
}
