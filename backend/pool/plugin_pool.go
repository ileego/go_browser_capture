package pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MessageSender interface {
	SendMessage(data []byte)
}

type PluginStatus string

const (
	StatusIdle         PluginStatus = "IDLE"
	StatusBusy         PluginStatus = "BUSY"
	StatusDisconnected PluginStatus = "DISCONNECTED"
)

type PluginInfo struct {
	ClientID            string
	Status              PluginStatus
	CurrentTasks        map[string]bool
	TotalTasksProcessed int
	AverageResponseTime time.Duration
	LastHeartbeat       time.Time
	LastResponseTime    time.Time
	Sender              MessageSender
	connectedAt         time.Time
}

func (p *PluginInfo) ConnectedAt() time.Time {
	return p.connectedAt
}

func NewPluginInfo(sender MessageSender) *PluginInfo {
	return &PluginInfo{
		ClientID:            uuid.New().String(),
		Status:              StatusIdle,
		CurrentTasks:        make(map[string]bool),
		TotalTasksProcessed: 0,
		LastHeartbeat:       time.Now(),
		Sender:              sender,
		connectedAt:         time.Now(),
	}
}

func (p *PluginInfo) GetTaskCount() int {
	return len(p.CurrentTasks)
}

func (p *PluginInfo) AddTask(taskID string) {
	p.CurrentTasks[taskID] = true
	if len(p.CurrentTasks) > 0 {
		p.Status = StatusBusy
	}
}

func (p *PluginInfo) RemoveTask(taskID string) {
	delete(p.CurrentTasks, taskID)
	if len(p.CurrentTasks) == 0 {
		p.Status = StatusIdle
	}
}

func (p *PluginInfo) UpdateHeartbeat() {
	p.LastHeartbeat = time.Now()
}

func (p *PluginInfo) IsAlive(heartbeatTimeout time.Duration) bool {
	return time.Since(p.LastHeartbeat) < heartbeatTimeout
}

type PluginPool struct {
	plugins           map[string]*PluginInfo
	mu                sync.Mutex
	heartbeatTimeout  time.Duration
	maxTasksPerPlugin int
}

func NewPluginPool(heartbeatTimeout time.Duration, maxTasksPerPlugin int) *PluginPool {
	pool := &PluginPool{
		plugins:           make(map[string]*PluginInfo),
		heartbeatTimeout:  heartbeatTimeout,
		maxTasksPerPlugin: maxTasksPerPlugin,
	}
	go pool.cleanupLoop()
	return pool
}

func (p *PluginPool) Register(sender MessageSender) *PluginInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	pluginInfo := NewPluginInfo(sender)
	p.plugins[pluginInfo.ClientID] = pluginInfo
	return pluginInfo
}

func (p *PluginPool) RegisterWithID(sender MessageSender, clientID string) *PluginInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	pluginInfo := &PluginInfo{
		ClientID:            clientID,
		Status:              StatusIdle,
		CurrentTasks:        make(map[string]bool),
		TotalTasksProcessed: 0,
		LastHeartbeat:       time.Now(),
		Sender:              sender,
		connectedAt:         time.Now(),
	}
	p.plugins[clientID] = pluginInfo
	return pluginInfo
}

func (p *PluginPool) Unregister(clientID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.plugins, clientID)
}

func (p *PluginPool) GetByID(clientID string) (*PluginInfo, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	plugin, ok := p.plugins[clientID]
	return plugin, ok
}

func (p *PluginPool) UpdateHeartbeat(clientID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if plugin, ok := p.plugins[clientID]; ok {
		plugin.UpdateHeartbeat()
	}
}

func (p *PluginPool) AssignTask(taskID string) (*PluginInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var bestPlugin *PluginInfo
	minTasks := p.maxTasksPerPlugin + 1

	for _, plugin := range p.plugins {
		if !plugin.IsAlive(p.heartbeatTimeout) {
			continue
		}

		taskCount := plugin.GetTaskCount()
		if taskCount < minTasks && taskCount < p.maxTasksPerPlugin {
			minTasks = taskCount
			bestPlugin = plugin
		}
	}

	if bestPlugin == nil {
		return nil, fmt.Errorf("no available plugin")
	}

	bestPlugin.AddTask(taskID)
	return bestPlugin, nil
}

func (p *PluginPool) CompleteTask(clientID string, taskID string, responseTime time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if plugin, ok := p.plugins[clientID]; ok {
		plugin.RemoveTask(taskID)
		plugin.TotalTasksProcessed++

		if plugin.AverageResponseTime == 0 {
			plugin.AverageResponseTime = responseTime
		} else {
			plugin.AverageResponseTime = (plugin.AverageResponseTime*time.Duration(plugin.TotalTasksProcessed-1) + responseTime) / time.Duration(plugin.TotalTasksProcessed)
		}
	}
}

func (p *PluginPool) GetIdlePlugins() []*PluginInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	var result []*PluginInfo
	for _, plugin := range p.plugins {
		if plugin.IsAlive(p.heartbeatTimeout) && plugin.Status == StatusIdle {
			result = append(result, plugin)
		}
	}
	return result
}

func (p *PluginPool) GetActivePlugins() []*PluginInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	var result []*PluginInfo
	for _, plugin := range p.plugins {
		if plugin.IsAlive(p.heartbeatTimeout) {
			result = append(result, plugin)
		}
	}
	return result
}

func (p *PluginPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	total := len(p.plugins)
	active := 0
	idle := 0
	busy := 0
	totalTasks := 0
	currentTasks := 0

	for _, plugin := range p.plugins {
		if plugin.IsAlive(p.heartbeatTimeout) {
			active++
			switch plugin.Status {
			case StatusIdle:
				idle++
			case StatusBusy:
				busy++
			}
			totalTasks += plugin.TotalTasksProcessed
			currentTasks += plugin.GetTaskCount()
		}
	}

	return map[string]interface{}{
		"total":         total,
		"active":        active,
		"idle":          idle,
		"busy":          busy,
		"total_tasks":   totalTasks,
		"current_tasks": currentTasks,
	}
}

func (p *PluginPool) cleanupLoop() {
	ticker := time.NewTicker(p.heartbeatTimeout * 2)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		for clientID, plugin := range p.plugins {
			if !plugin.IsAlive(p.heartbeatTimeout) {
				delete(p.plugins, clientID)
			}
		}
		p.mu.Unlock()
	}
}

