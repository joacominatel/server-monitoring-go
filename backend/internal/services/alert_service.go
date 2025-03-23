package services

import (
	"fmt"
	"time"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/notifications"
	"gorm.io/gorm"
)

// AlertService servicio para gestionar alertas y umbrales
type AlertService struct {
	db            *gorm.DB
	logger        logger.Logger
	notifyManager *notifications.NotificationManager
	metricService *MetricService // Añadir para evitar dependencias circulares
}

// NewAlertService crea un nuevo servicio de alertas
func NewAlertService(db *gorm.DB, log logger.Logger, notifyManager *notifications.NotificationManager) *AlertService {
	return &AlertService{
		db:            db,
		logger:        log,
		notifyManager: notifyManager,
		// metricService se establecerá después para evitar dependencias circulares
	}
}

// SetMetricService establece el servicio de métricas (se llama después de la creación para evitar dependencias circulares)
func (as *AlertService) SetMetricService(metricService *MetricService) {
	as.metricService = metricService
	as.logger.Info("Servicio de métricas configurado en el servicio de alertas")
}

// CreateThreshold crea un nuevo umbral de alerta
func (as *AlertService) CreateThreshold(threshold *models.AlertThreshold) error {
	if !threshold.ValidateThreshold() {
		return fmt.Errorf("umbral de alerta inválido")
	}

	if err := as.db.Create(threshold).Error; err != nil {
		as.logger.Errorf("Error al crear umbral de alerta: %v", err)
		return err
	}

	as.logger.Infof("Umbral de alerta creado: %s", threshold.Name)
	return nil
}

// UpdateThreshold actualiza un umbral existente
func (as *AlertService) UpdateThreshold(threshold *models.AlertThreshold) error {
	if !threshold.ValidateThreshold() {
		return fmt.Errorf("umbral de alerta inválido")
	}

	if err := as.db.Save(threshold).Error; err != nil {
		as.logger.Errorf("Error al actualizar umbral de alerta: %v", err)
		return err
	}

	as.logger.Infof("Umbral de alerta actualizado: %s", threshold.Name)
	return nil
}

// DeleteThreshold elimina un umbral
func (as *AlertService) DeleteThreshold(id uint) error {
	if err := as.db.Delete(&models.AlertThreshold{}, id).Error; err != nil {
		as.logger.Errorf("Error al eliminar umbral de alerta: %v", err)
		return err
	}

	as.logger.Infof("Umbral de alerta eliminado: %d", id)
	return nil
}

// GetThreshold obtiene un umbral por ID
func (as *AlertService) GetThreshold(id uint) (*models.AlertThreshold, error) {
	var threshold models.AlertThreshold
	if err := as.db.First(&threshold, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("umbral de alerta no encontrado")
		}
		as.logger.Errorf("Error al obtener umbral de alerta: %v", err)
		return nil, err
	}

	return &threshold, nil
}

// GetAllThresholds obtiene todos los umbrales
func (as *AlertService) GetAllThresholds() ([]models.AlertThreshold, error) {
	var thresholds []models.AlertThreshold
	if err := as.db.Find(&thresholds).Error; err != nil {
		as.logger.Errorf("Error al obtener umbrales de alerta: %v", err)
		return nil, err
	}

	return thresholds, nil
}

// GetThresholdsByServer obtiene umbrales para un servidor específico
func (as *AlertService) GetThresholdsByServer(serverID uint) ([]models.AlertThreshold, error) {
	var thresholds []models.AlertThreshold

	// Obtener umbrales específicos del servidor y los globales (sin ServerID)
	if err := as.db.Where("server_id = ? OR server_id IS NULL", serverID).Find(&thresholds).Error; err != nil {
		as.logger.Errorf("Error al obtener umbrales para servidor %d: %v", serverID, err)
		return nil, err
	}

	// Obtener umbrales por grupo (si el servidor pertenece a algún grupo)
	var server models.Server
	if err := as.db.Preload("ServerGroups").First(&server, serverID).Error; err != nil {
		as.logger.Warnf("Error al obtener grupos del servidor %d: %v", serverID, err)
	} else {
		// Para cada grupo del servidor, obtener sus umbrales
		for _, group := range server.ServerGroups {
			var groupThresholds []models.AlertThreshold
			if err := as.db.Where("group_id = ?", group.ID).Find(&groupThresholds).Error; err != nil {
				as.logger.Warnf("Error al obtener umbrales para grupo %d: %v", group.ID, err)
				continue
			}
			thresholds = append(thresholds, groupThresholds...)
		}
	}

	return thresholds, nil
}

// CreateAlert crea una nueva alerta
func (as *AlertService) CreateAlert(alert *models.Alert) error {
	// Transacción para crear la alerta y actualizar el umbral
	err := as.db.Transaction(func(tx *gorm.DB) error {
		// Crear la alerta
		if err := tx.Create(alert).Error; err != nil {
			return err
		}

		// Actualizar el umbral si existe
		if alert.ThresholdID != 0 {
			if err := tx.Model(&models.AlertThreshold{}).Where("id = ?", alert.ThresholdID).
				Update("last_triggered_at", time.Now()).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		as.logger.Errorf("Error al crear alerta: %v", err)
		return err
	}

	// Enviar notificaciones si hay un umbral asociado
	if alert.ThresholdID != 0 {
		threshold, err := as.GetThreshold(alert.ThresholdID)
		if err == nil && threshold.Enabled {
			if err := as.notifyManager.NotifyAlert(alert, threshold); err != nil {
				as.logger.Errorf("Error al enviar notificaciones para alerta %d: %v", alert.ID, err)
			} else {
				// Actualizar la alerta con la información de notificación
				as.db.Model(alert).Updates(map[string]interface{}{
					"notified_at":     time.Now(),
					"notify_channels": alert.NotifyChannels,
				})
			}
		}
	}

	as.logger.Infof("Alerta creada: %s (ID: %d)", alert.Title, alert.ID)
	return nil
}

// GetAllAlerts obtiene todas las alertas con filtrado opcional
func (as *AlertService) GetAllAlerts(params map[string]interface{}) ([]models.Alert, error) {
	var alerts []models.Alert
	query := as.db.Order("triggered_at DESC")

	// Aplicar filtros si existen
	if serverID, ok := params["server_id"].(uint); ok {
		query = query.Where("server_id = ?", serverID)
	}

	if status, ok := params["status"].(models.AlertStatus); ok {
		query = query.Where("status = ?", status)
	}

	if severity, ok := params["severity"].(models.AlertSeverity); ok {
		query = query.Where("severity = ?", severity)
	}

	if startTime, ok := params["start_time"].(time.Time); ok {
		query = query.Where("triggered_at >= ?", startTime)
	}

	if endTime, ok := params["end_time"].(time.Time); ok {
		query = query.Where("triggered_at <= ?", endTime)
	}

	// Ejecutar consulta con preload de relaciones
	if err := query.Preload("Server").Find(&alerts).Error; err != nil {
		as.logger.Errorf("Error al obtener alertas: %v", err)
		return nil, err
	}

	return alerts, nil
}

// GetActiveAlerts obtiene todas las alertas activas
func (as *AlertService) GetActiveAlerts() ([]models.Alert, error) {
	var alerts []models.Alert
	if err := as.db.Where("status = ?", models.AlertStatusActive).
		Preload("Server").Find(&alerts).Error; err != nil {
		as.logger.Errorf("Error al obtener alertas activas: %v", err)
		return nil, err
	}

	return alerts, nil
}

// GetAlert obtiene una alerta por ID
func (as *AlertService) GetAlert(id uint) (*models.Alert, error) {
	var alert models.Alert
	if err := as.db.Preload("Server").First(&alert, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("alerta no encontrada")
		}
		as.logger.Errorf("Error al obtener alerta: %v", err)
		return nil, err
	}

	return &alert, nil
}

// AcknowledgeAlert marca una alerta como reconocida
func (as *AlertService) AcknowledgeAlert(id, userID uint, notes string) error {
	alert, err := as.GetAlert(id)
	if err != nil {
		return err
	}

	if !alert.CanAcknowledge() {
		return fmt.Errorf("la alerta no puede ser reconocida")
	}

	now := time.Now()
	if err := as.db.Model(alert).Updates(map[string]interface{}{
		"status":          models.AlertStatusAcknowledged,
		"acknowledged_at": now,
		"acknowledged_by": userID,
		"notes":           notes,
	}).Error; err != nil {
		as.logger.Errorf("Error al reconocer alerta: %v", err)
		return err
	}

	as.logger.Infof("Alerta %d reconocida por usuario %d", id, userID)
	return nil
}

// ResolveAlert marca una alerta como resuelta
func (as *AlertService) ResolveAlert(id, userID uint, notes string) error {
	alert, err := as.GetAlert(id)
	if err != nil {
		return err
	}

	if !alert.CanResolve() {
		return fmt.Errorf("la alerta no puede ser resuelta")
	}

	now := time.Now()
	if err := as.db.Model(alert).Updates(map[string]interface{}{
		"status":      models.AlertStatusResolved,
		"resolved_at": now,
		"notes":       notes,
	}).Error; err != nil {
		as.logger.Errorf("Error al resolver alerta: %v", err)
		return err
	}

	// Enviar notificación de resolución si la alerta fue notificada
	if len(alert.NotifyChannels) > 0 {
		alert.ResolvedAt = &now // Actualizar para que el tiempo esté disponible
		if err := as.notifyManager.NotifyResolvedAlert(alert); err != nil {
			as.logger.Errorf("Error al enviar notificación de resolución: %v", err)
		}
	}

	as.logger.Infof("Alerta %d resuelta por usuario %d", id, userID)
	return nil
}

// AutoResolveAlert marca una alerta como resuelta automáticamente
func (as *AlertService) AutoResolveAlert(id uint) error {
	alert, err := as.GetAlert(id)
	if err != nil {
		return err
	}

	if !alert.IsActive() {
		return nil // Ignorar si ya no está activa
	}

	now := time.Now()
	if err := as.db.Model(alert).Updates(map[string]interface{}{
		"status":      models.AlertStatusResolved,
		"resolved_at": now,
		"notes":       "Resuelta automáticamente al normalizarse los valores",
	}).Error; err != nil {
		as.logger.Errorf("Error al resolver alerta automáticamente: %v", err)
		return err
	}

	// Enviar notificación de resolución si la alerta fue notificada
	if len(alert.NotifyChannels) > 0 {
		alert.ResolvedAt = &now // Actualizar para que el tiempo esté disponible
		if err := as.notifyManager.NotifyResolvedAlert(alert); err != nil {
			as.logger.Errorf("Error al enviar notificación de resolución: %v", err)
		}
	}

	as.logger.Infof("Alerta %d resuelta automáticamente", id)
	return nil
}

// CheckMetricAgainstThresholds verifica una métrica contra los umbrales aplicables
func (as *AlertService) CheckMetricAgainstThresholds(metric *models.Metric) error {
	// Obtener umbrales aplicables a este servidor
	thresholds, err := as.GetThresholdsByServer(metric.ServerID)
	if err != nil {
		return err
	}

	for _, threshold := range thresholds {
		if !threshold.Enabled {
			continue
		}

		// Comprobar si ya se envió una alerta dentro del período de cooldown
		if threshold.LastTriggeredAt != nil {
			cooldownEnds := threshold.LastTriggeredAt.Add(time.Duration(threshold.CooldownMinutes) * time.Minute)
			if time.Now().Before(cooldownEnds) {
				continue
			}
		}

		// Verificar según el tipo de métrica
		var metricValue float64
		var metricName string

		switch threshold.MetricType {
		case models.MetricTypeCPU:
			metricValue = metric.CPUUsage
			metricName = "CPU"
		case models.MetricTypeMemory:
			// Calcular porcentaje de memoria usada
			memoryPercent := float64(metric.MemoryUsed) / float64(metric.MemoryTotal) * 100
			metricValue = memoryPercent
			metricName = "Memoria"
		case models.MetricTypeDisk:
			// Calcular porcentaje de disco usado
			diskPercent := float64(metric.DiskUsed) / float64(metric.DiskTotal) * 100
			metricValue = diskPercent
			metricName = "Disco"
		case models.MetricTypeNetworkIn:
			// Convertir a MB para mejor legibilidad
			metricValue = float64(metric.NetDownload) / 1024 / 1024
			metricName = "Red (entrada)"
		case models.MetricTypeNetworkOut:
			// Convertir a MB para mejor legibilidad
			metricValue = float64(metric.NetUpload) / 1024 / 1024
			metricName = "Red (salida)"
		default:
			continue
		}

		// Evaluar la condición
		triggered := false

		switch threshold.Operator {
		case ">":
			triggered = metricValue > threshold.Value
		case "<":
			triggered = metricValue < threshold.Value
		case ">=":
			triggered = metricValue >= threshold.Value
		case "<=":
			triggered = metricValue <= threshold.Value
		case "==":
			triggered = metricValue == threshold.Value
		}

		// Crear una alerta si se cumple la condición
		if triggered {
			var serverName string
			var server models.Server

			if err := as.db.First(&server, metric.ServerID).Error; err != nil {
				serverName = fmt.Sprintf("Servidor #%d", metric.ServerID)
			} else {
				serverName = server.Hostname
			}

			alert := &models.Alert{
				Title: fmt.Sprintf("Alerta: %s en %s", metricName, serverName),
				Message: fmt.Sprintf("La métrica %s ha alcanzado un valor de %.2f%%, superando el umbral establecido de %.2f%%",
					metricName, metricValue, threshold.Value),
				MetricType:  threshold.MetricType,
				MetricValue: metricValue,
				Threshold:   threshold.Value,
				Operator:    threshold.Operator,
				Severity:    threshold.Severity,
				Status:      models.AlertStatusActive,
				ServerID:    metric.ServerID,
				ThresholdID: threshold.ID,
				TriggeredAt: time.Now(),
			}

			if err := as.CreateAlert(alert); err != nil {
				as.logger.Errorf("Error al crear alerta para umbral %d: %v", threshold.ID, err)
				continue
			}
		} else {
			// Verificar si hay alertas activas que deban resolverse
			var activeAlerts []models.Alert
			if err := as.db.Where("server_id = ? AND metric_type = ? AND status = ? AND threshold_id = ?",
				metric.ServerID, threshold.MetricType, models.AlertStatusActive, threshold.ID).
				Find(&activeAlerts).Error; err != nil {
				continue
			}

			for _, activeAlert := range activeAlerts {
				// Verificar si la condición ya no se cumple para resolverla
				shouldResolve := false

				switch threshold.Operator {
				case ">":
					shouldResolve = metricValue <= threshold.Value
				case "<":
					shouldResolve = metricValue >= threshold.Value
				case ">=":
					shouldResolve = metricValue < threshold.Value
				case "<=":
					shouldResolve = metricValue > threshold.Value
				case "==":
					shouldResolve = metricValue != threshold.Value
				}

				if shouldResolve {
					if err := as.AutoResolveAlert(activeAlert.ID); err != nil {
						as.logger.Errorf("Error al resolver automáticamente alerta %d: %v", activeAlert.ID, err)
					}
				}
			}
		}
	}

	return nil
}
