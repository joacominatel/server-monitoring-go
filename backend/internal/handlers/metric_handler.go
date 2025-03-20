package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// MetricHandler maneja las rutas relacionadas con métricas
type MetricHandler struct {
	metricService *services.MetricService
	serverService *services.ServerService
	logger        logger.Logger
}

// NewMetricHandler crea una nueva instancia del manejador de métricas
func NewMetricHandler(metricService *services.MetricService, serverService *services.ServerService, logger logger.Logger) *MetricHandler {
	return &MetricHandler{
		metricService: metricService,
		serverService: serverService,
		logger:        logger,
	}
}

// RegisterRoutes registra las rutas del manejador en el router
func (h *MetricHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/metrics")
	{
		api.POST("", h.CreateMetric)
		api.GET("/server/:server_id", h.GetMetricsByServerID)
		api.GET("/server/:server_id/latest", h.GetLatestMetricByServerID)
		api.GET("/server/:server_id/timerange", h.GetMetricsByTimeRange)
	}
}

// CreateMetric recibe y almacena una nueva métrica
func (h *MetricHandler) CreateMetric(c *gin.Context) {
	var metric models.Metric
	
	if err := c.ShouldBindJSON(&metric); err != nil {
		h.logger.Warnf("Datos de métrica inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de métrica inválidos"})
		return
	}
	
	// Verificar que el servidor existe
	_, err := h.serverService.GetServerByID(metric.ServerID)
	if err != nil {
		h.logger.Warnf("Servidor con ID %d no encontrado", metric.ServerID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Servidor no encontrado"})
		return
	}
	
	if err := h.metricService.CreateMetric(&metric); err != nil {
		h.logger.Errorf("Error al crear métrica: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear métrica"})
		return
	}
	
	c.JSON(http.StatusCreated, metric)
}

// GetMetricsByServerID obtiene métricas por ID de servidor con paginación
func (h *MetricHandler) GetMetricsByServerID(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		h.logger.Warnf("ID de servidor inválido: %s", serverIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}
	
	// Parámetros de paginación con valores por defecto
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 100
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	metrics, err := h.metricService.GetMetricsByServerID(uint(serverID), limit, offset)
	if err != nil {
		h.logger.Errorf("Error al obtener métricas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener métricas"})
		return
	}
	
	c.JSON(http.StatusOK, metrics)
}

// GetLatestMetricByServerID obtiene la métrica más reciente de un servidor
func (h *MetricHandler) GetLatestMetricByServerID(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		h.logger.Warnf("ID de servidor inválido: %s", serverIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}
	
	metric, err := h.metricService.GetLatestMetricByServerID(uint(serverID))
	if err != nil {
		h.logger.Errorf("Error al obtener última métrica: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Métrica no encontrada"})
		return
	}
	
	c.JSON(http.StatusOK, metric)
}

// GetMetricsByTimeRange obtiene métricas por ID de servidor en un rango de tiempo
func (h *MetricHandler) GetMetricsByTimeRange(c *gin.Context) {
	serverIDStr := c.Param("server_id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		h.logger.Warnf("ID de servidor inválido: %s", serverIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}
	
	// Parámetros de rango de tiempo
	startStr := c.Query("start")
	endStr := c.Query("end")
	
	// Formato de tiempo ISO8601
	var startTime, endTime time.Time
	
	if startStr != "" {
		startTime, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			h.logger.Warnf("Formato de tiempo de inicio inválido: %s", startStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de tiempo de inicio inválido"})
			return
		}
	} else {
		// Si no se proporciona tiempo de inicio, usar hace 24 horas
		startTime = time.Now().Add(-24 * time.Hour)
	}
	
	if endStr != "" {
		endTime, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			h.logger.Warnf("Formato de tiempo de fin inválido: %s", endStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de tiempo de fin inválido"})
			return
		}
	} else {
		// Si no se proporciona tiempo de fin, usar ahora
		endTime = time.Now()
	}
	
	metrics, err := h.metricService.GetMetricsByTimeRange(uint(serverID), startTime, endTime)
	if err != nil {
		h.logger.Errorf("Error al obtener métricas por rango de tiempo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener métricas"})
		return
	}
	
	c.JSON(http.StatusOK, metrics)
} 