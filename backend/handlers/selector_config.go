package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ileego/go_browser_capture/backend/application"
)

type SelectorConfigHandler struct {
	appService *application.SelectorConfigAppService
}

func NewSelectorConfigHandler(appService *application.SelectorConfigAppService) *SelectorConfigHandler {
	return &SelectorConfigHandler{appService: appService}
}

// CreateSelectorConfig godoc
// @Summary Create a new selector config
// @Description Create a new selector configuration
// @Tags selector-config
// @Accept json
// @Produce json
// @Param request body application.CreateSelectorConfigRequest true "Create request"
// @Success 201 {object} application.SelectorConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs [post]
func (h *SelectorConfigHandler) Create(c *gin.Context) {
	var req application.CreateSelectorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.appService.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetSelectorConfigByID godoc
// @Summary Get selector config by ID
// @Description Get a selector configuration by its ID
// @Tags selector-config
// @Accept json
// @Produce json
// @Param id path string true "Selector config ID"
// @Success 200 {object} application.SelectorConfigResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs/{id} [get]
func (h *SelectorConfigHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.appService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selector config not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSelectorConfigByName godoc
// @Summary Get selector config by name
// @Description Get a selector configuration by its name
// @Tags selector-config
// @Accept json
// @Produce json
// @Param name query string true "Selector config name"
// @Success 200 {object} application.SelectorConfigResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs/name/{name} [get]
func (h *SelectorConfigHandler) GetByName(c *gin.Context) {
	name := c.Param("name")

	resp, err := h.appService.GetByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selector config not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetSelectorConfigsByDomain godoc
// @Summary Get selector configs by domain
// @Description Get all selector configurations for a specific domain
// @Tags selector-config
// @Accept json
// @Produce json
// @Param domain query string true "Domain"
// @Success 200 {array} application.SelectorConfigResponse
// @Failure 500 {object} map[string]string
// @Router /selector-configs/domain/{domain} [get]
func (h *SelectorConfigHandler) GetByDomain(c *gin.Context) {
	domain := c.Param("domain")

	resp, err := h.appService.GetByDomain(c.Request.Context(), domain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAllSelectorConfigs godoc
// @Summary Get all selector configs
// @Description Get all selector configurations
// @Tags selector-config
// @Accept json
// @Produce json
// @Success 200 {array} application.SelectorConfigResponse
// @Failure 500 {object} map[string]string
// @Router /selector-configs [get]
func (h *SelectorConfigHandler) GetAll(c *gin.Context) {
	resp, err := h.appService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetActiveSelectorConfigs godoc
// @Summary Get active selector configs
// @Description Get all active selector configurations
// @Tags selector-config
// @Accept json
// @Produce json
// @Success 200 {array} application.SelectorConfigResponse
// @Failure 500 {object} map[string]string
// @Router /selector-configs/active [get]
func (h *SelectorConfigHandler) GetActive(c *gin.Context) {
	resp, err := h.appService.GetActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateSelectorConfig godoc
// @Summary Update selector config
// @Description Update a selector configuration
// @Tags selector-config
// @Accept json
// @Produce json
// @Param id path string true "Selector config ID"
// @Param request body application.UpdateSelectorConfigRequest true "Update request"
// @Success 200 {object} application.SelectorConfigResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs/{id} [put]
func (h *SelectorConfigHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req application.UpdateSelectorConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.appService.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selector config not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteSelectorConfig godoc
// @Summary Delete selector config
// @Description Delete a selector configuration
// @Tags selector-config
// @Accept json
// @Produce json
// @Param id path string true "Selector config ID"
// @Success 204 {string} string "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs/{id} [delete]
func (h *SelectorConfigHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.appService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ActivateSelectorConfig godoc
// @Summary Activate selector config
// @Description Activate a selector configuration
// @Tags selector-config
// @Accept json
// @Produce json
// @Param id path string true "Selector config ID"
// @Success 200 {object} application.SelectorConfigResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs/{id}/activate [post]
func (h *SelectorConfigHandler) Activate(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.appService.Activate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selector config not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeactivateSelectorConfig godoc
// @Summary Deactivate selector config
// @Description Deactivate a selector configuration
// @Tags selector-config
// @Accept json
// @Produce json
// @Param id path string true "Selector config ID"
// @Success 200 {object} application.SelectorConfigResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /selector-configs/{id}/deactivate [post]
func (h *SelectorConfigHandler) Deactivate(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.appService.Deactivate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selector config not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
