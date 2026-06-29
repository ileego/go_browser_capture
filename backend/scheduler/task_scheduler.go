package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/ileego/go_browser_capture/backend/pool"
	"github.com/ileego/go_browser_capture/backend/websocket"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "PENDING"
	TaskStatusAssigned  TaskStatus = "ASSIGNED"
	TaskStatusRunning   TaskStatus = "RUNNING"
	TaskStatusCompleted TaskStatus = "COMPLETED"
	TaskStatusFailed    TaskStatus = "FAILED"
	TaskStatusTimeout   TaskStatus = "TIMEOUT"
)

type Task struct {
	ID          string
	URL         string
	Selector    string
	Timeout     time.Duration
	Status      TaskStatus
	PluginID    string
	RetryCount  int
	MaxRetries  int
	CreatedAt   time.Time
	AssignedAt  time.Time
	CompletedAt time.Time
	Result      *websocket.CaptureResponse
	Error       string
}

type TaskScheduler struct {
	pluginPool         *pool.PluginPool
	tasks              map[string]*Task
	pendingRequests    map[string]chan *websocket.CaptureResponse
	maxQueueSize       int
	maxConcurrentTasks int
	timeout            time.Duration
	maxRetries         int
	mu                 sync.RWMutex
}

func NewTaskScheduler(pluginPool *pool.PluginPool, maxQueueSize, maxConcurrentTasks int, timeout time.Duration, maxRetries int) *TaskScheduler {
	return &TaskScheduler{
		pluginPool:         pluginPool,
		tasks:              make(map[string]*Task),
		pendingRequests:    make(map[string]chan *websocket.CaptureResponse),
		maxQueueSize:       maxQueueSize,
		maxConcurrentTasks: maxConcurrentTasks,
		timeout:            timeout,
		maxRetries:         maxRetries,
	}
}

func (s *TaskScheduler) SubmitTask(ctx context.Context, id, url, selector string, timeout time.Duration) (<-chan *websocket.CaptureResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.getPendingTaskCountUnsafe() >= s.maxQueueSize {
		return nil, errors.New("task queue is full")
	}

	if s.pluginPool.GetActivePlugins() == nil || len(s.pluginPool.GetActivePlugins()) == 0 {
		return nil, errors.New("no available plugin")
	}

	if timeout <= 0 {
		timeout = s.timeout
	}

	task := &Task{
		ID:         id,
		URL:        url,
		Selector:   selector,
		Timeout:    timeout,
		Status:     TaskStatusPending,
		RetryCount: 0,
		MaxRetries: s.maxRetries,
		CreatedAt:  time.Now(),
	}

	ch := make(chan *websocket.CaptureResponse, 1)
	s.pendingRequests[id] = ch
	s.tasks[id] = task

	go s.processTask(task)

	return ch, nil
}

func (s *TaskScheduler) processTask(task *Task) {
	for task.RetryCount <= task.MaxRetries {
		s.mu.Lock()
		if task.Status == TaskStatusCompleted {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()

		plugin, err := s.pluginPool.AssignTask(task.ID)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		s.mu.Lock()
		if task.Status == TaskStatusCompleted {
			s.mu.Unlock()
			s.pluginPool.CompleteTask(plugin.ClientID, task.ID, time.Since(time.Now()))
			return
		}
		task.Status = TaskStatusAssigned
		task.PluginID = plugin.ClientID
		task.AssignedAt = time.Now()
		reqCh, ok := s.pendingRequests[task.ID]
		s.mu.Unlock()

		if !ok {
			log.Printf("Task %s: no pending request channel found", task.ID)
			s.pluginPool.CompleteTask(plugin.ClientID, task.ID, time.Since(task.AssignedAt))
			s.mu.Lock()
			task.RetryCount++
			task.Status = TaskStatusPending
			task.PluginID = ""
			s.mu.Unlock()
			continue
		}

		s.sendTaskToPlugin(task, plugin)

		select {
		case resp := <-reqCh:
			s.mu.Lock()
			delete(s.pendingRequests, task.ID)
			if resp != nil && resp.Success {
				task.Status = TaskStatusCompleted
				task.Result = resp
				task.CompletedAt = time.Now()
				s.mu.Unlock()
				s.pluginPool.CompleteTask(plugin.ClientID, task.ID, time.Since(task.AssignedAt))
				return
			} else if resp != nil && !resp.Success {
				task.Error = resp.Error
				task.Status = TaskStatusFailed
			} else {
				task.Error = "timeout"
				task.Status = TaskStatusTimeout
			}
			s.mu.Unlock()

		case <-time.After(task.Timeout):
			s.mu.Lock()
			delete(s.pendingRequests, task.ID)
			task.Error = "timeout"
			task.Status = TaskStatusTimeout
			s.mu.Unlock()
		}

		s.pluginPool.CompleteTask(plugin.ClientID, task.ID, time.Since(task.AssignedAt))

		s.mu.Lock()
		if task.Status == TaskStatusCompleted {
			s.mu.Unlock()
			return
		}
		task.RetryCount++
		task.Status = TaskStatusPending
		task.PluginID = ""
		newCh := make(chan *websocket.CaptureResponse, 1)
		s.pendingRequests[task.ID] = newCh
		s.mu.Unlock()
	}

	s.mu.Lock()
	task.Status = TaskStatusFailed
	if task.Error == "" {
		task.Error = "max retries exceeded"
	}
	task.Result = &websocket.CaptureResponse{
		ID:      task.ID,
		Success: false,
		Error:   task.Error,
	}

	if ch, ok := s.pendingRequests[task.ID]; ok {
		ch <- task.Result
		delete(s.pendingRequests, task.ID)
	}
	s.mu.Unlock()
}

func (s *TaskScheduler) sendTaskToPlugin(task *Task, plugin *pool.PluginInfo) {
	req := &websocket.CaptureRequest{
		ID:       task.ID,
		Action:   "capture",
		URL:      task.URL,
		Selector: task.Selector,
	}

	data, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error marshaling task: %v", err)
		return
	}

	plugin.Sender.SendMessage(data)
}

func (s *TaskScheduler) waitForResponse(taskID string, timeout time.Duration) <-chan *websocket.CaptureResponse {
	ch := make(chan *websocket.CaptureResponse, 1)

	go func() {
		s.mu.Lock()
		reqCh, ok := s.pendingRequests[taskID]
		s.mu.Unlock()

		if !ok {
			ch <- nil
			return
		}

		select {
		case resp := <-reqCh:
			ch <- resp
		case <-time.After(timeout):
			ch <- nil
		}
	}()

	return ch
}

func (s *TaskScheduler) HandleResponse(resp *websocket.CaptureResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.pendingRequests[resp.ID]; ok {
		select {
		case ch <- resp:
		default:
		}
		delete(s.pendingRequests, resp.ID)
	}

	if task, ok := s.tasks[resp.ID]; ok {
		if resp.Success {
			task.Status = TaskStatusCompleted
			task.Result = resp
			task.CompletedAt = time.Now()
		} else {
			task.Error = resp.Error
			task.Status = TaskStatusFailed
		}
	}
}

func (s *TaskScheduler) GetTask(taskID string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[taskID]
	return task, ok
}

func (s *TaskScheduler) GetAllTasks(status TaskStatus, hasStatusFilter bool) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if !hasStatusFilter || task.Status == status {
			tasks = append(tasks, task)
		}
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	return tasks
}

func (s *TaskScheduler) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pending := 0
	assigned := 0
	running := 0
	completed := 0
	failed := 0

	for _, task := range s.tasks {
		switch task.Status {
		case TaskStatusPending:
			pending++
		case TaskStatusAssigned:
			assigned++
		case TaskStatusRunning:
			running++
		case TaskStatusCompleted:
			completed++
		case TaskStatusFailed, TaskStatusTimeout:
			failed++
		}
	}

	return map[string]interface{}{
		"pending":    pending,
		"assigned":   assigned,
		"running":    running,
		"completed":  completed,
		"failed":     failed,
		"total":      len(s.tasks),
		"queue_size": s.maxQueueSize,
	}
}

func (s *TaskScheduler) getPendingTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getPendingTaskCountUnsafe()
}

func (s *TaskScheduler) getPendingTaskCountUnsafe() int {
	count := 0
	for _, task := range s.tasks {
		if task.Status == TaskStatusPending {
			count++
		}
	}
	return count
}
