package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// LogLevel representa el nivel de severidad del log
type LogLevel string

// Constantes para los niveles de log
const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// Metadata representa información adicional del log en formato JSON
type Metadata map[string]interface{}

// Value implementa la interfaz driver.Valuer para guardar Metadata como JSON en la base de datos
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implementa la interfaz sql.Scanner para cargar JSON como Metadata
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(Metadata)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("tipo de dato no válido para Metadata")
	}
	
	return json.Unmarshal(bytes, m)
}

// Log representa un registro de log almacenado en la base de datos
type Log struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Level     LogLevel  `gorm:"size:10;not null;index" json:"level"`
	Message   string    `gorm:"size:1000;not null" json:"message"`
	Source    string    `gorm:"size:100;index" json:"source"`
	Metadata  Metadata  `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
} 