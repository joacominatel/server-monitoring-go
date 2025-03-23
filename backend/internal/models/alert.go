package models

import (
	"time"

	"gorm.io/gorm"
)

// AlertStatus define los estados posibles de una alerta
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"       // Alerta activa
	AlertStatusResolved     AlertStatus = "resolved"     // Problema resuelto
	AlertStatusAcknowledged AlertStatus = "acknowledged" // Reconocida pero no resuelta
	AlertStatusSuppressed   AlertStatus = "suppressed"   // Alerta suprimida temporalmente
)

// Alert representa una alerta generada a partir de un umbral
type Alert struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Información de la alerta
	Title       string        `json:"title" gorm:"size:200;not null"`
	Message     string        `json:"message" gorm:"type:text;not null"`
	MetricType  MetricType    `json:"metric_type" gorm:"size:20;not null"`
	MetricValue float64       `json:"metric_value"`
	Threshold   float64       `json:"threshold"`
	Operator    string        `json:"operator" gorm:"size:5"`
	Severity    AlertSeverity `json:"severity" gorm:"size:10;not null"`
	Status      AlertStatus   `json:"status" gorm:"size:15;not null;default:'active'"`

	// Relaciones
	ServerID       uint           `json:"server_id" gorm:"index;not null"`
	Server         Server         `json:"server" gorm:"foreignKey:ServerID"`
	ThresholdID    uint           `json:"threshold_id" gorm:"index"`
	AlertThreshold AlertThreshold `json:"alert_threshold,omitempty" gorm:"foreignKey:ThresholdID"`

	// Campos temporales
	TriggeredAt    time.Time  `json:"triggered_at"`    // Momento en que se detectó la condición de alerta
	ResolvedAt     *time.Time `json:"resolved_at"`     // Momento en que la condición se resolvió
	AcknowledgedAt *time.Time `json:"acknowledged_at"` // Momento en que se reconoció la alerta
	AcknowledgedBy *uint      `json:"acknowledged_by"` // Usuario que reconoció la alerta

	// Campos para notificaciones
	NotifiedAt     *time.Time `json:"notified_at"`                            // Momento en que se envió la notificación
	NotifyChannels []string   `json:"notify_channels" gorm:"serializer:json"` // Canales por los que se notificó

	// Notas y comentarios
	Notes string `json:"notes" gorm:"type:text"`

	// Campos comunes
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName especifica el nombre de la tabla en la base de datos
func (Alert) TableName() string {
	return "alerts"
}

// IsActive verifica si la alerta está activa
func (a *Alert) IsActive() bool {
	return a.Status == AlertStatusActive
}

// CanAcknowledge verifica si la alerta puede ser reconocida
func (a *Alert) CanAcknowledge() bool {
	return a.Status == AlertStatusActive
}

// CanResolve verifica si la alerta puede ser resuelta
func (a *Alert) CanResolve() bool {
	return a.Status == AlertStatusActive || a.Status == AlertStatusAcknowledged
}
