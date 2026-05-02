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
	go tm.processTaskQueue()
	tm.cleanupTicker = time.NewTicker(time.Hour)
	go tm.cleanupTasks()
}

func (tm *TaskManager) processTaskQueue() {
	for task := range tm.taskQueue {
		tm.mutex.Lock()
		if tm.runningTasks >= tm.maxConcurrent {
			tm.mutex.Unlock()
			time.Sleep(100 * time.Millisecond)
			tm.taskQueue <- task
			continue
		}
		tm.runningTasks++
		tm.mutex.Unlock()
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
