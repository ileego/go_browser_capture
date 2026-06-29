package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ileego/go_browser_capture/backend/application"
)

type BatchCaptureHandler struct {
	appService *application.BatchCaptureAppService
}

func NewBatchCaptureHandler(appService *application.BatchCaptureAppService) *BatchCaptureHandler {
	return &BatchCaptureHandler{appService: appService}
}

// BatchCapture godoc
// @Summary Batch capture DOM from URLs
// @Description Receive URLs, match selector configs by domain, and send capture requests to Chrome extension
// @Tags capture
// @Accept json
// @Produce json
// @Param request body application.BatchCaptureRequest true "Batch capture request"
// @Success 200 {object} application.BatchCaptureResponse
// @Failure 400 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /batch-capture [post]
func (h *BatchCaptureHandler) BatchCapture(c *gin.Context) {
	var req application.BatchCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.URLs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URLs list cannot be empty"})
		return
	}

	resp, err := h.appService.CaptureBatch(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "no Chrome extension connected" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

