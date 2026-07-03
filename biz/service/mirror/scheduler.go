package mirror

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/queue"
)

var GlobalScheduler *Scheduler

type Scheduler struct {
	mirrorDAO *db.MirrorDAO
	queue     queue.UniqueQueue
	cron      *cron.Cron
	scanTick  *time.Ticker
	cfg       configs.MirrorConfig
	cronMap   map[uint]cron.EntryID
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

func (s *Scheduler) Reload() {
	for id, entryID := range s.cronMap {
		s.cron.Remove(entryID)
		delete(s.cronMap, id)
	}
	s.loadCronMirrors()
}

func (s *Scheduler) AddCronMirror(mirror *po.Mirror) {
	if mirror.CronExpr == "" || !mirror.Enabled {
		return
	}

	if _, exists := s.cronMap[mirror.ID]; exists {
		s.RemoveCronMirror(mirror.ID)
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
