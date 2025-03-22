package models

import (
	"time"

	"gorm.io/gorm"
)

// Server representa un servidor registrado en el sistema
type Server struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Hostname    string         `gorm:"size:255;not null;uniqueIndex" json:"hostname"`
	IP          string         `gorm:"size:45;not null" json:"ip"`
	Description string         `gorm:"size:500" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Información del sistema operativo
	OS        string `gorm:"size:100" json:"os"`        // Nombre del SO (ej. "Windows Server 2022", "Ubuntu 22.04")
	OSVersion string `gorm:"size:50" json:"os_version"` // Versión específica del SO
	OSArch    string `gorm:"size:20" json:"os_arch"`    // Arquitectura (ej. "x86_64", "arm64")
	Kernel    string `gorm:"size:100" json:"kernel"`    // Versión del kernel

	// Información de hardware
	CPUModel    string `gorm:"size:200" json:"cpu_model"` // Modelo de CPU
	CPUCores    int    `json:"cpu_cores"`                 // Número de núcleos físicos
	CPUThreads  int    `json:"cpu_threads"`               // Número de hilos
	TotalMemory int64  `json:"total_memory"`              // Memoria total en bytes
	TotalDisk   int64  `json:"total_disk"`                // Almacenamiento total en bytes

	// Relaciones
	Metrics []Metric `gorm:"foreignKey:ServerID" json:"metrics,omitempty"`
}
