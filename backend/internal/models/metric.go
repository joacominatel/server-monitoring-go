package models

import (
	"time"

	"gorm.io/gorm"
)

// Metric representa una medición de métricas de un servidor
type Metric struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ServerID  uint      `gorm:"not null;index" json:"server_id"`
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	
	// Métricas de CPU
	CPUUsage     float64 `gorm:"not null" json:"cpu_usage"`      // Porcentaje de uso de CPU (0-100)
	CPUTemp      float64 `json:"cpu_temp,omitempty"`             // Temperatura en grados Celsius
	
	// Métricas de Memoria
	MemoryTotal  int64   `gorm:"not null" json:"memory_total"`   // Total de memoria en bytes
	MemoryUsed   int64   `gorm:"not null" json:"memory_used"`    // Memoria usada en bytes
	MemoryFree   int64   `gorm:"not null" json:"memory_free"`    // Memoria libre en bytes
	
	// Métricas de Disco
	DiskTotal    int64   `gorm:"not null" json:"disk_total"`     // Espacio total en disco en bytes
	DiskUsed     int64   `gorm:"not null" json:"disk_used"`      // Espacio usado en disco en bytes
	DiskFree     int64   `gorm:"not null" json:"disk_free"`      // Espacio libre en disco en bytes
	
	// Métricas de Red
	NetUpload    int64   `json:"net_upload"`                     // Bytes subidos desde la última medición
	NetDownload  int64   `json:"net_download"`                   // Bytes descargados desde la última medición
	
	// Relaciones
	Server       Server  `gorm:"foreignKey:ServerID" json:"-"`
}

// BeforeCreate es un hook GORM que se ejecuta antes de crear un registro
func (m *Metric) BeforeCreate(tx *gorm.DB) error {
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	return nil
} 