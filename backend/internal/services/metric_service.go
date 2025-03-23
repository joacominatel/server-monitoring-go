package services

import (
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/websocket"
	"gorm.io/gorm"
)

// MetricService maneja la lógica de negocio relacionada con métricas
type MetricService struct {
	db           *gorm.DB
	logger       logger.Logger
	hub          *websocket.Hub
	redisClient  *redis.Client
	alertService *AlertService // Servicio de alertas para verificar umbrales
}

// NewMetricService crea una nueva instancia del servicio de métricas
func NewMetricService(db *gorm.DB, logger logger.Logger, hub *websocket.Hub, redisClient *redis.Client) *MetricService {
	return &MetricService{
		db:          db,
		logger:      logger,
		hub:         hub,
		redisClient: redisClient,
		// alertService se establecerá después para evitar dependencias circulares
	}
}

// SetAlertService establece el servicio de alertas (se llama después de la creación para evitar dependencias circulares)
func (s *MetricService) SetAlertService(alertService *AlertService) {
	s.alertService = alertService
	s.logger.Info("Servicio de alertas configurado en el servicio de métricas")
}

// CreateMetric guarda una nueva métrica
func (s *MetricService) CreateMetric(metric *models.Metric) error {
	if err := s.db.Create(metric).Error; err != nil {
		s.logger.Errorf("Error al crear métrica para servidor ID %d: %v", metric.ServerID, err)
		return err
	}

	// Transmitir la métrica a través de WebSockets
	if s.hub != nil {
		s.broadcastMetric(metric)
	}

	// Verificar umbrales de alerta si el servicio está configurado
	if s.alertService != nil {
		if err := s.alertService.CheckMetricAgainstThresholds(metric); err != nil {
			s.logger.Warnf("Error al verificar umbrales para métrica del servidor %d: %v", metric.ServerID, err)
			// No devolvemos este error para no interrumpir el flujo principal
		}
	}

	s.logger.Infof("Métrica creada exitosamente para servidor ID %d", metric.ServerID)
	return nil
}

// GetMetricsByServerID obtiene métricas por ID de servidor con paginación
func (s *MetricService) GetMetricsByServerID(serverID uint, limit, offset int) ([]models.Metric, error) {
	var metrics []models.Metric

	query := s.db.Where("server_id = ?", serverID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&metrics).Error; err != nil {
		s.logger.Errorf("Error al obtener métricas para servidor ID %d: %v", serverID, err)
		return nil, err
	}

	return metrics, nil
}

// GetMetricsByTimeRange obtiene métricas por ID de servidor en un rango de tiempo
func (s *MetricService) GetMetricsByTimeRange(serverID uint, startTime, endTime time.Time) ([]models.Metric, error) {
	var metrics []models.Metric

	query := s.db.Where("server_id = ? AND timestamp BETWEEN ? AND ?",
		serverID, startTime, endTime).
		Order("timestamp ASC")

	if err := query.Find(&metrics).Error; err != nil {
		s.logger.Errorf("Error al obtener métricas por rango de tiempo para servidor ID %d: %v", serverID, err)
		return nil, err
	}

	return metrics, nil
}

// GetLatestMetricByServerID obtiene la métrica más reciente de un servidor
func (s *MetricService) GetLatestMetricByServerID(serverID uint) (*models.Metric, error) {
	var metric models.Metric

	if err := s.db.Where("server_id = ?", serverID).
		Order("timestamp DESC").
		First(&metric).Error; err != nil {

		s.logger.Errorf("Error al obtener última métrica para servidor ID %d: %v", serverID, err)
		return nil, err
	}

	return &metric, nil
}

// DeleteOldMetrics elimina métricas más antiguas que la fecha especificada
func (s *MetricService) DeleteOldMetrics(olderThan time.Time) (int64, error) {
	result := s.db.Where("timestamp < ?", olderThan).Delete(&models.Metric{})

	if result.Error != nil {
		s.logger.Errorf("Error al eliminar métricas antiguas: %v", result.Error)
		return 0, result.Error
	}

	s.logger.Infof("Se eliminaron %d métricas antiguas (anteriores a %v)", result.RowsAffected, olderThan)
	return result.RowsAffected, nil
}

// broadcastMetric transmite una métrica a través de WebSockets
func (s *MetricService) broadcastMetric(metric *models.Metric) {
	if s.hub == nil {
		return
	}

	// Enviar a través del hub WebSocket
	s.hub.BroadcastToServer(metric.ServerID, metric)
}

// HasWebSocketHub retorna true si el servicio tiene un hub WebSocket configurado
func (s *MetricService) HasWebSocketHub() bool {
	return s.hub != nil
}

// HasRedisClient retorna true si el servicio tiene un cliente Redis configurado
func (s *MetricService) HasRedisClient() bool {
	return s.redisClient != nil
}
