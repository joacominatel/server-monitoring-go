package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/internal/middleware"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// AlertHandler manejador para las rutas de alertas
type AlertHandler struct {
	service *services.AlertService
	logger  logger.Logger
}

// NewAlertHandler crea un nuevo manejador para alertas
func NewAlertHandler(service *services.AlertService, log logger.Logger) *AlertHandler {
	return &AlertHandler{
		service: service,
		logger:  log,
	}
}

// RegisterRoutes registra las rutas relacionadas con alertas
func (h *AlertHandler) RegisterRoutes(router gin.IRouter, authMiddleware *middleware.AuthMiddleware) {
	alerts := router.Group("/alerts")
	{
		// Rutas accesibles a todos los usuarios autenticados
		alerts.GET("", h.GetAllAlerts)
		alerts.GET("/active", h.GetActiveAlerts)
		alerts.GET("/:id", h.GetAlertByID)

		// Rutas para gestionar alertas (requieren rol de admin o user)
		adminOrUser := alerts.Group("")
		adminOrUser.Use(authMiddleware.RequireRole(models.RoleAdmin, models.RoleUser))
		{
			adminOrUser.POST("/:id/acknowledge", h.AcknowledgeAlert)
			adminOrUser.POST("/:id/resolve", h.ResolveAlert)
		}

		// Rutas de umbrales (todas requieren admin)
		thresholds := router.Group("/alert-thresholds")
		thresholds.Use(authMiddleware.RequireRole(models.RoleAdmin))
		{
			thresholds.GET("", h.GetAllThresholds)
			thresholds.GET("/:id", h.GetThresholdByID)
			thresholds.POST("", h.CreateThreshold)
			thresholds.PUT("/:id", h.UpdateThreshold)
			thresholds.DELETE("/:id", h.DeleteThreshold)
			thresholds.GET("/server/:server_id", h.GetThresholdsByServer)
		}
	}
}

// GetAllAlerts obtiene todas las alertas con filtros opcionales
func (h *AlertHandler) GetAllAlerts(c *gin.Context) {
	// Preparar filtros
	filters := make(map[string]interface{})

	// Filtro por servidor
	if serverIDStr := c.Query("server_id"); serverIDStr != "" {
		serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
		if err == nil {
			filters["server_id"] = uint(serverID)
		} else {
			h.logger.Warnf("ID de servidor inválido: %s", serverIDStr)
		}
	}

	// Filtro por estado
	if status := c.Query("status"); status != "" {
		filters["status"] = models.AlertStatus(status)
	}

	// Filtro por severidad
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = models.AlertSeverity(severity)
	}

	// Filtros por fecha
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			filters["start_time"] = startTime
		} else {
			h.logger.Warnf("Formato de fecha inicial inválido: %s", startTimeStr)
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			filters["end_time"] = endTime
		} else {
			h.logger.Warnf("Formato de fecha final inválido: %s", endTimeStr)
		}
	}

	// Obtener alertas
	alerts, err := h.service.GetAllAlerts(filters)
	if err != nil {
		h.logger.Errorf("Error al obtener alertas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener alertas"})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetActiveAlerts obtiene solo las alertas activas
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	alerts, err := h.service.GetActiveAlerts()
	if err != nil {
		h.logger.Errorf("Error al obtener alertas activas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener alertas activas"})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetAlertByID obtiene una alerta por su ID
func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	alert, err := h.service.GetAlert(uint(id))
	if err != nil {
		h.logger.Errorf("Error al obtener alerta %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Alerta no encontrada"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// AcknowledgeAlert marca una alerta como reconocida
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var input struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Obtener ID de usuario del contexto
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	if err := h.service.AcknowledgeAlert(uint(id), userID, input.Notes); err != nil {
		h.logger.Errorf("Error al reconocer alerta %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alerta reconocida correctamente"})
}

// ResolveAlert marca una alerta como resuelta
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var input struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Obtener ID de usuario del contexto
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}

	if err := h.service.ResolveAlert(uint(id), userID, input.Notes); err != nil {
		h.logger.Errorf("Error al resolver alerta %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alerta resuelta correctamente"})
}

// GetAllThresholds obtiene todos los umbrales de alerta
func (h *AlertHandler) GetAllThresholds(c *gin.Context) {
	thresholds, err := h.service.GetAllThresholds()
	if err != nil {
		h.logger.Errorf("Error al obtener umbrales: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener umbrales"})
		return
	}

	c.JSON(http.StatusOK, thresholds)
}

// GetThresholdByID obtiene un umbral por su ID
func (h *AlertHandler) GetThresholdByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	threshold, err := h.service.GetThreshold(uint(id))
	if err != nil {
		h.logger.Errorf("Error al obtener umbral %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Umbral no encontrado"})
		return
	}

	c.JSON(http.StatusOK, threshold)
}

// CreateThreshold crea un nuevo umbral de alerta
func (h *AlertHandler) CreateThreshold(c *gin.Context) {
	var threshold models.AlertThreshold
	if err := c.ShouldBindJSON(&threshold); err != nil {
		h.logger.Warnf("Error en binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Obtener ID del usuario del contexto
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	threshold.CreatedBy = userID

	if err := h.service.CreateThreshold(&threshold); err != nil {
		h.logger.Errorf("Error al crear umbral: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, threshold)
}

// UpdateThreshold actualiza un umbral existente
func (h *AlertHandler) UpdateThreshold(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var threshold models.AlertThreshold
	if err := c.ShouldBindJSON(&threshold); err != nil {
		h.logger.Warnf("Error en binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	threshold.ID = uint(id)

	if err := h.service.UpdateThreshold(&threshold); err != nil {
		h.logger.Errorf("Error al actualizar umbral %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, threshold)
}

// DeleteThreshold elimina un umbral
func (h *AlertHandler) DeleteThreshold(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.DeleteThreshold(uint(id)); err != nil {
		h.logger.Errorf("Error al eliminar umbral %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Umbral eliminado correctamente"})
}

// GetThresholdsByServer obtiene los umbrales aplicables a un servidor
func (h *AlertHandler) GetThresholdsByServer(c *gin.Context) {
	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	thresholds, err := h.service.GetThresholdsByServer(uint(serverID))
	if err != nil {
		h.logger.Errorf("Error al obtener umbrales para servidor %d: %v", serverID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener umbrales"})
		return
	}

	c.JSON(http.StatusOK, thresholds)
}
