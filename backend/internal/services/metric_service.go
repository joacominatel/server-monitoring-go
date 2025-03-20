package services

import (
	"time"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
)

// MetricService maneja la lógica de negocio relacionada con métricas
type MetricService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewMetricService crea una nueva instancia del servicio de métricas
func NewMetricService(db *gorm.DB, logger logger.Logger) *MetricService {
	return &MetricService{
		db:     db,
		logger: logger,
	}
}

// CreateMetric guarda una nueva métrica
func (s *MetricService) CreateMetric(metric *models.Metric) error {
	if err := s.db.Create(metric).Error; err != nil {
		s.logger.Errorf("Error al crear métrica para servidor ID %d: %v", metric.ServerID, err)
		return err
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