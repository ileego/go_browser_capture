package handlers

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ileego/go_browser_capture/backend/websocket"
)

type Handler struct {
	wsServer *websocket.Server
	mu       *sync.Mutex
}

func NewHandler(wsServer *websocket.Server) *Handler {
	return &Handler{
		wsServer: wsServer,
		mu:       &sync.Mutex{},
	}
}

func (h *Handler) HandleCaptureRequest(req *websocket.CaptureRequest, client *websocket.Client) {
	// 处理来自插件的捕获请求（通过 WebSocket 接收）
	// 主要用于日志记录
	log.Printf("Received capture request from plugin: ID=%s, URL=%s", req.ID, req.URL)
	_ = client // 未使用
}

// healthCheck godoc
// @Summary Health check endpoint
// @Description Check if the server is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}

// captureHandler godoc
// @Summary Capture DOM from URL
// @Description Send a capture request to Chrome extension via WebSocket
// @Tags capture
// @Accept json
// @Produce json
// @Param request body websocket.CaptureRequest true "Capture request"
// @Success 200 {object} websocket.CaptureResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /capture [post]
func (h *Handler) CaptureHandler(c *gin.Context) {
	var req websocket.CaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.URL == "" || req.Selector == "" {
		c.JSON(400, gin.H{"error": "URL and Selector are required"})
		return
	}

	if h.wsServer.GetClientCount() == 0 {
		c.JSON(500, gin.H{"error": "No Chrome extension connected"})
		return
	}

	ch := make(chan *websocket.CaptureResponse, 1)
	h.wsServer.RegisterPendingRequest(req.ID, ch)

	data, err := json.Marshal(req)
	if err != nil {
		h.wsServer.RemovePendingRequest(req.ID)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	h.wsServer.BroadcastMessage(data)

	timeout := time.Second * 30
	if req.Timeout > 0 {
		timeout = time.Millisecond * time.Duration(req.Timeout)
	}

	select {
	case resp := <-ch:
		c.JSON(200, resp)
	case <-time.After(timeout):
		h.wsServer.RemovePendingRequest(req.ID)
		c.JSON(500, gin.H{"error": "Request timeout"})
	}
}

// statusHandler godoc
// @Summary Get server status
// @Description Get the current status of the server and connected clients
// @Tags status
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /status [get]
func (h *Handler) StatusHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":            "running",
		"websocket_clients": h.wsServer.GetClientCount(),
		"timestamp":         time.Now().Format(time.RFC3339),
	})
}

