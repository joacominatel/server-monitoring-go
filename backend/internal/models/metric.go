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
	CPUUsage  float64 `gorm:"not null" json:"cpu_usage"` // Porcentaje de uso de CPU (0-100)
	CPUTemp   float64 `json:"cpu_temp,omitempty"`        // Temperatura en grados Celsius
	CPUFreq   float64 `json:"cpu_freq,omitempty"`        // Frecuencia actual en MHz
	LoadAvg1  float64 `json:"load_avg_1,omitempty"`      // Carga promedio 1 minuto
	LoadAvg5  float64 `json:"load_avg_5,omitempty"`      // Carga promedio 5 minutos
	LoadAvg15 float64 `json:"load_avg_15,omitempty"`     // Carga promedio 15 minutos

	// Métricas de Memoria
	MemoryTotal   int64 `gorm:"not null" json:"memory_total"` // Total de memoria en bytes
	MemoryUsed    int64 `gorm:"not null" json:"memory_used"`  // Memoria usada en bytes
	MemoryFree    int64 `gorm:"not null" json:"memory_free"`  // Memoria libre en bytes
	MemoryCache   int64 `json:"memory_cache,omitempty"`       // Memoria caché en bytes
	MemoryBuffers int64 `json:"memory_buffers,omitempty"`     // Memoria en buffers en bytes
	SwapTotal     int64 `json:"swap_total,omitempty"`         // Memoria swap total en bytes
	SwapUsed      int64 `json:"swap_used,omitempty"`          // Memoria swap usada en bytes
	SwapFree      int64 `json:"swap_free,omitempty"`          // Memoria swap libre en bytes

	// Métricas de Disco
	DiskTotal      int64 `gorm:"not null" json:"disk_total"` // Espacio total en disco en bytes
	DiskUsed       int64 `gorm:"not null" json:"disk_used"`  // Espacio usado en disco en bytes
	DiskFree       int64 `gorm:"not null" json:"disk_free"`  // Espacio libre en disco en bytes
	DiskReads      int64 `json:"disk_reads,omitempty"`       // Operaciones de lectura desde último muestreo
	DiskWrites     int64 `json:"disk_writes,omitempty"`      // Operaciones de escritura desde último muestreo
	DiskReadBytes  int64 `json:"disk_read_bytes,omitempty"`  // Bytes leídos desde último muestreo
	DiskWriteBytes int64 `json:"disk_write_bytes,omitempty"` // Bytes escritos desde último muestreo
	DiskIOTime     int64 `json:"disk_io_time,omitempty"`     // Tiempo de IO en milisegundos

	// Métricas de Red
	NetUpload     int64 `json:"net_upload"`                // Bytes subidos desde la última medición
	NetDownload   int64 `json:"net_download"`              // Bytes descargados desde la última medición
	NetPacketsIn  int64 `json:"net_packets_in,omitempty"`  // Paquetes recibidos desde último muestreo
	NetPacketsOut int64 `json:"net_packets_out,omitempty"` // Paquetes enviados desde último muestreo
	NetErrorsIn   int64 `json:"net_errors_in,omitempty"`   // Errores de recepción
	NetErrorsOut  int64 `json:"net_errors_out,omitempty"`  // Errores de envío
	NetDropsIn    int64 `json:"net_drops_in,omitempty"`    // Paquetes desechados de entrada
	NetDropsOut   int64 `json:"net_drops_out,omitempty"`   // Paquetes desechados de salida

	// Procesos y servicios
	ProcessCount int `json:"process_count,omitempty"` // Número total de procesos
	ThreadCount  int `json:"thread_count,omitempty"`  // Número total de hilos
	HandleCount  int `json:"handle_count,omitempty"`  // Número de handles/descriptores abiertos

	// Tiempo de actividad
	Uptime int64 `json:"uptime,omitempty"` // Tiempo de actividad en segundos

	// Relaciones
	Server Server `gorm:"foreignKey:ServerID" json:"-"`
}

// BeforeCreate es un hook GORM que se ejecuta antes de crear un registro
func (m *Metric) BeforeCreate(tx *gorm.DB) error {
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}
	return nil
}
