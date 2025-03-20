package services

import (
	"time"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
)

// LogService maneja la lógica de negocio relacionada con logs
type LogService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewLogService crea una nueva instancia del servicio de logs
func NewLogService(db *gorm.DB, logger logger.Logger) *LogService {
	return &LogService{
		db:     db,
		logger: logger,
	}
}

// CreateLog guarda un nuevo log en la base de datos
func (s *LogService) CreateLog(level models.LogLevel, message, source string, metadata models.Metadata) error {
	log := &models.Log{
		Level:     level,
		Message:   message,
		Source:    source,
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(log).Error; err != nil {
		s.logger.Errorf("Error al guardar log en base de datos: %v", err)
		return err
	}

	return nil
}

// GetLogs obtiene logs con filtros y paginación
func (s *LogService) GetLogs(level models.LogLevel, source string, startDate, endDate time.Time, limit, offset int) ([]models.Log, error) {
	var logs []models.Log
	
	query := s.db.Model(&models.Log{})
	
	// Aplicar filtros si se especifican
	if level != "" {
		query = query.Where("level = ?", level)
	}
	
	if source != "" {
		query = query.Where("source = ?", source)
	}
	
	// Filtrar por rango de fechas si se especifican
	if !startDate.IsZero() && !endDate.IsZero() {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	} else if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	} else if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}
	
	// Ordenar por fecha descendente (más reciente primero)
	query = query.Order("created_at DESC")
	
	// Aplicar paginación
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&logs).Error; err != nil {
		s.logger.Errorf("Error al consultar logs: %v", err)
		return nil, err
	}
	
	return logs, nil
}

// DeleteOldLogs elimina logs más antiguos que la fecha especificada
func (s *LogService) DeleteOldLogs(olderThan time.Time) (int64, error) {
	result := s.db.Where("created_at < ?", olderThan).Delete(&models.Log{})
	
	if result.Error != nil {
		s.logger.Errorf("Error al eliminar logs antiguos: %v", result.Error)
		return 0, result.Error
	}
	
	s.logger.Infof("Se eliminaron %d logs antiguos (anteriores a %v)", result.RowsAffected, olderThan)
	return result.RowsAffected, nil
} 