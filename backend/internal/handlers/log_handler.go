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

// LogHandler maneja las rutas relacionadas con logs
type LogHandler struct {
	logService *services.LogService
	logger     logger.Logger
}

// NewLogHandler crea una nueva instancia del manejador de logs
func NewLogHandler(logService *services.LogService, logger logger.Logger) *LogHandler {
	return &LogHandler{
		logService: logService,
		logger:     logger,
	}
}

// RegisterRoutes registra las rutas del manejador en el router
func (h *LogHandler) RegisterRoutes(router gin.IRouter) {
	logs := router.Group("/logs")
	{
		// Rutas para consulta y mantenimiento de logs
		// Nota: estas rutas ya están protegidas en main.go con RequireRole(models.RoleAdmin)
		logs.GET("", h.GetLogs)
		logs.DELETE("/cleanup", h.CleanupOldLogs)
	}
}

// LogQueryParams representa los parámetros de consulta para filtrar logs
type LogQueryParams struct {
	Level     string `form:"level"`
	Source    string `form:"source"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Limit     int    `form:"limit,default=100"`
	Offset    int    `form:"offset,default=0"`
}

// GetLogs obtiene logs con filtros y paginación
func (h *LogHandler) GetLogs(c *gin.Context) {
	var params LogQueryParams
	
	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Warnf("Error en parámetros de consulta de logs: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros de consulta inválidos"})
		return
	}
	
	// Convertir nivel a LogLevel
	var level models.LogLevel
	if params.Level != "" {
		level = models.LogLevel(params.Level)
	}
	
	// Convertir fechas
	var startDate, endDate time.Time
	var err error
	
	if params.StartDate != "" {
		startDate, err = time.Parse(time.RFC3339, params.StartDate)
		if err != nil {
			h.logger.Warnf("Fecha de inicio inválida: %s", params.StartDate)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha de inicio inválido"})
			return
		}
	}
	
	if params.EndDate != "" {
		endDate, err = time.Parse(time.RFC3339, params.EndDate)
		if err != nil {
			h.logger.Warnf("Fecha de fin inválida: %s", params.EndDate)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de fecha de fin inválido"})
			return
		}
	}
	
	// Limitar los resultados por defecto a 100
	if params.Limit <= 0 {
		params.Limit = 100
	} else if params.Limit > 1000 {
		params.Limit = 1000 // Establecer un límite máximo
	}
	
	if params.Offset < 0 {
		params.Offset = 0
	}
	
	logs, err := h.logService.GetLogs(level, params.Source, startDate, endDate, params.Limit, params.Offset)
	if err != nil {
		h.logger.Errorf("Error al obtener logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener logs"})
		return
	}
	
	c.JSON(http.StatusOK, logs)
}

// CleanupOldLogs elimina logs más antiguos que una fecha específica
func (h *LogHandler) CleanupOldLogs(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		h.logger.Warnf("Valor de días inválido: %s", daysStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valor de días inválido"})
		return
	}
	
	cutoffDate := time.Now().AddDate(0, 0, -days)
	
	count, err := h.logService.DeleteOldLogs(cutoffDate)
	if err != nil {
		h.logger.Errorf("Error al eliminar logs antiguos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar logs antiguos"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Logs antiguos eliminados exitosamente",
		"deleted":    count,
		"older_than": cutoffDate,
	})
} 