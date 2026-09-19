package git

import (
	"log"
	"sync"
	"time"
)

type TaskManager struct {
	tasks         sync.Map
	maxConcurrent int
	runningTasks  int
	mutex         sync.Mutex
	taskQueue     chan *Task
	slots         chan struct{} // 容量 = maxConcurrent 的信号量，容量闸门以此为准
	cleanupTicker *time.Ticker
}

type Task struct {
	mu        sync.Mutex `json:"-"`
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	Progress  []string   `json:"progress"`
	Error     string     `json:"error"`
	StartTime time.Time  `json:"startTime"`
	EndTime   time.Time  `json:"endTime"`
}

func (t *Task) Snapshot() *Task {
	t.mu.Lock()
	defer t.mu.Unlock()
	progress := make([]string, len(t.Progress))
	copy(progress, t.Progress)
	return &Task{
		ID:        t.ID,
		Status:    t.Status,
		Progress:  progress,
		Error:     t.Error,
		StartTime: t.StartTime,
		EndTime:   t.EndTime,
	}
}

var GlobalTaskManager = &TaskManager{
	maxConcurrent: 100,
	taskQueue:     make(chan *Task, 1000),
}

func (tm *TaskManager) Init() {
	tm.slots = make(chan struct{}, tm.maxConcurrent)
	go tm.processTaskQueue()
	tm.cleanupTicker = time.NewTicker(time.Hour)
	go tm.cleanupTasks()
}

func (tm *TaskManager) processTaskQueue() {
	for task := range tm.taskQueue {
		// 阻塞获取槽位：队列满时上游 AddTask 依赖本队列缓冲排队，
		// 这里绝不回投（回投会让唯一消费者等自己，造成死锁）。
		tm.slots <- struct{}{}
		log.Printf("[INFO] Starting task: %s", task.ID)
	}
}

func (tm *TaskManager) AddTask(id string) *Task {
	t := &Task{
		ID:        id,
		Status:    "running",
		Progress:  []string{},
		StartTime: time.Now(),
	}
	tm.mutex.Lock()
	tm.runningTasks++
	tm.mutex.Unlock()
	tm.tasks.Store(id, t)
	tm.taskQueue <- t
	return t
}

func (tm *TaskManager) GetTask(id string) (*Task, bool) {
	v, ok := tm.tasks.Load(id)
	if !ok {
		return nil, false
	}
	return v.(*Task).Snapshot(), true
}

func (tm *TaskManager) AppendLog(id string, msg string) {
	if v, ok := tm.tasks.Load(id); ok {
		t := v.(*Task)
		t.mu.Lock()
		t.Progress = append(t.Progress, msg)
		t.mu.Unlock()
	}
}

func (tm *TaskManager) UpdateStatus(id string, status string, errStr string) {
	if v, ok := tm.tasks.Load(id); ok {
		t := v.(*Task)
		t.mu.Lock()
		t.Status = status
		t.Error = errStr
		t.EndTime = time.Now()
		t.mu.Unlock()

		if status == "success" || status == "failed" {
			tm.mutex.Lock()
			tm.runningTasks--
			tm.mutex.Unlock()
			<-tm.slots
			log.Printf("[INFO] Task %s completed with status: %s", id, status)
		}
	}
}

func (tm *TaskManager) cleanupTasks() {
	for range tm.cleanupTicker.C {
		log.Printf("[INFO] Starting task cleanup")
		count := 0
		tm.tasks.Range(func(key, value interface{}) bool {
			task := value.(*Task)
			task.mu.Lock()
			done := task.Status == "success" || task.Status == "failed"
			endTime := task.EndTime
			task.mu.Unlock()
			if done && time.Since(endTime) > 24*time.Hour {
				tm.tasks.Delete(key)
				count++
			}
			return true
		})
		log.Printf("[INFO] Cleaned up %d completed tasks", count)
	}
}

func (tm *TaskManager) GetRunningTasksCount() int {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	return tm.runningTasks
}

func (tm *TaskManager) GetQueueLength() int {
	return len(tm.taskQueue)
}
