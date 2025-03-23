package models

import (
	"time"

	"gorm.io/gorm"
)

// ServerGroup representa un grupo de servidores para organización jerárquica
type ServerGroup struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null"`
	Description string         `json:"description" gorm:"type:text"`
	ParentID    *uint          `json:"parent_id" gorm:"index"` // ID del grupo padre (puede ser nulo para grupos raíz)
	Parent      *ServerGroup   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children    []*ServerGroup `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Servers     []Server       `json:"servers,omitempty" gorm:"many2many:server_group_servers;"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	CreatedBy   uint           `json:"created_by"` // ID del usuario que creó el grupo
}

// TableName especifica el nombre de la tabla en la base de datos
func (ServerGroup) TableName() string {
	return "server_groups"
}

// BeforeCreate es un hook que se ejecuta antes de crear un nuevo grupo
func (sg *ServerGroup) BeforeCreate(tx *gorm.DB) error {
	// Evitar ciclos en la jerarquía
	if sg.ParentID != nil && *sg.ParentID == sg.ID {
		return gorm.ErrInvalidField
	}
	return nil
}
