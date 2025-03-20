package database

import (
	"context"
	"errors"
	"time"

	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger es un adaptador para utilizar nuestro logger con GORM
type GormLogger struct {
	Logger        logger.Logger
	SlowThreshold time.Duration
}

// NewGormLogger crea una nueva instancia de GormLogger
func NewGormLogger(log logger.Logger) gormlogger.Interface {
	return &GormLogger{
		Logger:        log,
		SlowThreshold: 200 * time.Millisecond,
	}
}

// LogMode implementa la interfaz gormlogger.Interface
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info implementa la interfaz gormlogger.Interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Infof(msg, data...)
}

// Warn implementa la interfaz gormlogger.Interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warnf(msg, data...)
}

// Error implementa la interfaz gormlogger.Interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Errorf(msg, data...)
}

// Trace implementa la interfaz gormlogger.Interface
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	
	// Incluir información de filas afectadas
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.Logger.WithFields(map[string]interface{}{
			"error":    err,
			"sql":      sql,
			"duration": elapsed,
		}).Error("Error en consulta SQL")
	case elapsed > l.SlowThreshold:
		l.Logger.WithFields(map[string]interface{}{
			"sql":      sql,
			"rows":     rows,
			"duration": elapsed,
		}).Warn("Consulta SQL lenta")
	default:
		l.Logger.WithFields(map[string]interface{}{
			"sql":      sql,
			"rows":     rows,
			"duration": elapsed,
		}).Debug("Consulta SQL")
	}
} 