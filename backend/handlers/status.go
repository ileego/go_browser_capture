package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ileego/go_browser_capture/backend/pool"
	"github.com/ileego/go_browser_capture/backend/scheduler"
)

type StatusHandler struct {
	pluginPool    *pool.PluginPool
	taskScheduler *scheduler.TaskScheduler
}

func NewStatusHandler(pluginPool *pool.PluginPool, taskScheduler *scheduler.TaskScheduler) *StatusHandler {
	return &StatusHandler{
		pluginPool:    pluginPool,
		taskScheduler: taskScheduler,
	}
}

func (h *StatusHandler) GetPlugins(c *gin.Context) {
	plugins := h.pluginPool.GetActivePlugins()

	var result []map[string]interface{}
	for _, p := range plugins {
		result = append(result, map[string]interface{}{
			"id":                  p.ClientID,
			"status":              p.Status,
			"taskCount":           p.GetTaskCount(),
			"totalTasksProcessed": p.TotalTasksProcessed,
			"avgResponseTime":     p.AverageResponseTime.String(),
			"lastHeartbeat":       p.LastHeartbeat,
			"connectedAt":         p.ConnectedAt(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"total":   len(result),
	})
}

func (h *StatusHandler) GetPluginByID(c *gin.Context) {
	plugin, ok := h.pluginPool.GetByID(c.Param("id"))
	if !ok || plugin == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "插件不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":                  plugin.ClientID,
			"status":              plugin.Status,
			"taskCount":           plugin.GetTaskCount(),
			"totalTasksProcessed": plugin.TotalTasksProcessed,
			"avgResponseTime":     plugin.AverageResponseTime.String(),
			"lastHeartbeat":       plugin.LastHeartbeat,
			"connectedAt":         plugin.ConnectedAt(),
		},
	})
}

func (h *StatusHandler) GetTasks(c *gin.Context) {
	statusFilter := c.Query("status")

	var status scheduler.TaskStatus
	var hasStatusFilter bool
	if statusFilter != "" {
		status = scheduler.TaskStatus(statusFilter)
		hasStatusFilter = true
	}

	tasks := h.taskScheduler.GetAllTasks(status, hasStatusFilter)
	stats := h.taskScheduler.GetStats()

	var result []map[string]interface{}
	for _, task := range tasks {
		result = append(result, map[string]interface{}{
			"id":          task.ID,
			"url":         task.URL,
			"selector":    task.Selector,
			"status":      task.Status,
			"pluginID":    task.PluginID,
			"error":       task.Error,
			"retryCount":  task.RetryCount,
			"maxRetries":  task.MaxRetries,
			"createdAt":   task.CreatedAt,
			"assignedAt":  task.AssignedAt,
			"completedAt": task.CompletedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
		"total":   len(result),
		"stats":   stats,
	})
}

func (h *StatusHandler) GetTaskByID(c *gin.Context) {
	task, ok := h.taskScheduler.GetTask(c.Param("id"))
	if !ok || task == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"id":          task.ID,
			"url":         task.URL,
			"selector":    task.Selector,
			"status":      task.Status,
			"pluginID":    task.PluginID,
			"error":       task.Error,
			"retryCount":  task.RetryCount,
			"maxRetries":  task.MaxRetries,
			"createdAt":   task.CreatedAt,
			"assignedAt":  task.AssignedAt,
			"completedAt": task.CompletedAt,
		},
	})
}

func (h *StatusHandler) GetSystemStatus(c *gin.Context) {
	pluginStats := h.pluginPool.GetStats()
	taskStats := h.taskScheduler.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"plugins": pluginStats,
			"tasks":   taskStats,
		},
	})
}
