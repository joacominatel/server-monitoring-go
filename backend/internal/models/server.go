package models

import (
	"time"

	"gorm.io/gorm"
)

// Server representa un servidor registrado en el sistema
type Server struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Hostname    string    `gorm:"size:255;not null;uniqueIndex" json:"hostname"`
	IP          string    `gorm:"size:45;not null" json:"ip"`
	Description string    `gorm:"size:500" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relaciones
	Metrics []Metric `gorm:"foreignKey:ServerID" json:"metrics,omitempty"`
} 