package mirror

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/queue"
)

var GlobalScheduler *Scheduler

// GlobalWorkerPool / GlobalQueue 由装配方（cmd/server/main.go 的
// initMirrorSystem）赋值，供 Shutdown 统一停机。
var (
	GlobalWorkerPool *queue.WorkerPool
	GlobalQueue      queue.UniqueQueue
)

// Shutdown 依序停止镜像子系统：先停调度器（不再入队），再排空 worker
// （等待在跑的同步完成），最后关闭队列连接。
func Shutdown() {
	StopScheduler()
	if GlobalWorkerPool != nil {
		GlobalWorkerPool.Stop()
	}
	if GlobalQueue != nil {
		GlobalQueue.Close()
	}
	log.Println("[Mirror] system shut down")
}

type Scheduler struct {
	mirrorDAO *db.MirrorDAO
	queue     queue.UniqueQueue
	cron      *cron.Cron
	scanTick  *time.Ticker
	cfg       configs.MirrorConfig
	cronMap   map[uint]cron.EntryID
	cronMu    sync.Mutex
	stopCh    chan struct{}
}

func NewScheduler(mirrorDAO *db.MirrorDAO, q queue.UniqueQueue, cfg configs.MirrorConfig) *Scheduler {
	return &Scheduler{
		mirrorDAO: mirrorDAO,
		queue:     q,
		cron:      cron.New(cron.WithSeconds()),
		cfg:       cfg,
		cronMap:   make(map[uint]cron.EntryID),
		stopCh:    make(chan struct{}),
	}
}

func InitScheduler(mirrorDAO *db.MirrorDAO, q queue.UniqueQueue, cfg configs.MirrorConfig) {
	GlobalScheduler = NewScheduler(mirrorDAO, q, cfg)
	GlobalScheduler.Start()
}

func StopScheduler() {
	if GlobalScheduler != nil {
		GlobalScheduler.Stop()
	}
}

func (s *Scheduler) Start() {
	s.loadCronMirrors()

	scanInterval := s.cfg.ScanInterval
	if scanInterval <= 0 {
		scanInterval = 30
	}
	s.scanTick = time.NewTicker(time.Duration(scanInterval) * time.Second)

	go s.scanLoop()

	s.cron.Start()
	log.Println("[MirrorScheduler] started")
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.cron.Stop()
	if s.scanTick != nil {
		s.scanTick.Stop()
	}
	log.Println("[MirrorScheduler] stopped")
}

// AddCronMirror / RemoveCronMirror 对 nil 接收者安全：desktop 等未装配
// scheduler 的入口调用时不 panic。
func (s *Scheduler) AddCronMirror(mirror *po.Mirror) {
	if s == nil {
		return
	}
	if mirror.CronExpr == "" || !mirror.Enabled {
		return
	}

	s.cronMu.Lock()
	defer s.cronMu.Unlock()

	// Remove old entry if exists (inline to avoid recursive lock)
	if entryID, exists := s.cronMap[mirror.ID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronMap, mirror.ID)
	}

	entryID, err := s.cron.AddFunc(mirror.CronExpr, func() {
		s.queue.Push(queue.SyncRequest{
			MirrorID:    mirror.ID,
			TriggerType: po.TriggerTypeCron,
			RequestedAt: time.Now(),
		})
	})
	if err != nil {
		log.Printf("[MirrorScheduler] failed to add cron for mirror %d: %v", mirror.ID, err)
		return
	}

	s.cronMap[mirror.ID] = entryID
}

func (s *Scheduler) RemoveCronMirror(mirrorID uint) {
	if s == nil {
		return
	}
	s.cronMu.Lock()
	defer s.cronMu.Unlock()
	if entryID, exists := s.cronMap[mirrorID]; exists {
		s.cron.Remove(entryID)
		delete(s.cronMap, mirrorID)
	}
}

func (s *Scheduler) scanLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.scanTick.C:
			s.scanDueMirrors()
		}
	}
}

func (s *Scheduler) scanDueMirrors() {
	mirrors, err := s.mirrorDAO.FindDueForSync()
	if err != nil {
		log.Printf("[MirrorScheduler] scan error: %v", err)
		return
	}

	for i := range mirrors {
		m := &mirrors[i]
		if m.CronExpr != "" {
			continue
		}

		err := s.queue.Push(queue.SyncRequest{
			MirrorID:    m.ID,
			TriggerType: po.TriggerTypeCron,
			RequestedAt: time.Now(),
		})
		if err != nil {
			log.Printf("[MirrorScheduler] failed to push mirror %d: %v", m.ID, err)
		}
	}
}

func (s *Scheduler) loadCronMirrors() {
	mirrors, err := s.mirrorDAO.FindEnabled()
	if err != nil {
		log.Printf("[MirrorScheduler] failed to load mirrors: %v", err)
		return
	}

	for i := range mirrors {
		m := &mirrors[i]
		if m.CronExpr != "" {
			s.AddCronMirror(m)
		}
	}
}
