package models

import (
	"time"

	"gorm.io/gorm"
)

// MetricType define los tipos de métricas para alertas
type MetricType string

const (
	MetricTypeCPU        MetricType = "cpu"
	MetricTypeMemory     MetricType = "memory"
	MetricTypeDisk       MetricType = "disk"
	MetricTypeNetworkIn  MetricType = "network_in"
	MetricTypeNetworkOut MetricType = "network_out"
)

// AlertSeverity define los niveles de severidad para las alertas
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertThreshold representa un umbral para generar alertas de métricas
type AlertThreshold struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Description string `json:"description" gorm:"type:text"`

	// Configuración del umbral
	MetricType MetricType    `json:"metric_type" gorm:"size:20;not null"`
	Operator   string        `json:"operator" gorm:"size:5;not null"` // >, <, >=, <=, ==
	Value      float64       `json:"value" gorm:"not null"`           // Valor para comparar
	Duration   int           `json:"duration"`                        // Duración en segundos que debe mantenerse la condición
	Severity   AlertSeverity `json:"severity" gorm:"size:10;not null"`

	// Notificaciones
	EnableEmail   bool   `json:"enable_email" gorm:"default:false"`
	EnableDiscord bool   `json:"enable_discord" gorm:"default:false"`
	EnableWebhook bool   `json:"enable_webhook" gorm:"default:false"`
	WebhookURL    string `json:"webhook_url" gorm:"size:255"`

	// Configuración de cooldown
	CooldownMinutes int `json:"cooldown_minutes" gorm:"default:15"` // Evitar múltiples alertas en este periodo

	// Relaciones
	ServerID *uint        `json:"server_id" gorm:"index"` // Puede ser nulo para aplicar a todos los servidores
	Server   *Server      `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	GroupID  *uint        `json:"group_id" gorm:"index"` // Aplicar a todos los servidores del grupo
	Group    *ServerGroup `json:"group,omitempty" gorm:"foreignKey:GroupID"`

	// Campos comunes
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	CreatedBy uint           `json:"created_by"` // Usuario que creó el umbral
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Estado de la última alerta
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
}

// ValidateThreshold valida que el umbral tenga valores adecuados
func (at *AlertThreshold) ValidateThreshold() bool {
	// Verificar que solo se aplique a un servidor o grupo, no ambos
	if at.ServerID != nil && at.GroupID != nil {
		return false
	}

	// Verificar que el operador sea válido
	validOperators := map[string]bool{">": true, "<": true, ">=": true, "<=": true, "==": true}
	if !validOperators[at.Operator] {
		return false
	}

	return true
}
