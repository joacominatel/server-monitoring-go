package logger

import (
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
)

// LogPersister es una interfaz para persistir logs en base de datos
type LogPersister interface {
	CreateLog(level models.LogLevel, message, source string, metadata models.Metadata) error
}

// DBLogger es un adaptador que combina el logger normal con persistencia en base de datos
type DBLogger struct {
	Logger
	persister LogPersister
	source    string
}

// NewDBLogger crea una nueva instancia de DBLogger
func NewDBLogger(logger Logger, persister LogPersister, source string) Logger {
	return &DBLogger{
		Logger:    logger,
		persister: persister,
		source:    source,
	}
}

// mapLogLevel convierte nivel de logger a modelo de log
func mapLogLevel(level string) models.LogLevel {
	switch level {
	case "debug":
		return models.LogLevelDebug
	case "info":
		return models.LogLevelInfo
	case "warn":
		return models.LogLevelWarn
	case "error":
		return models.LogLevelError
	case "fatal":
		return models.LogLevelFatal
	default:
		return models.LogLevelInfo
	}
}

// Debug implementa la interfaz Logger
func (l *DBLogger) Debug(args ...interface{}) {
	l.Logger.Debug(args...)
	
	msg := formatArgs(args...)
	// Ignoramos el error para evitar ciclos infinitos de logging
	_ = l.persister.CreateLog(models.LogLevelDebug, msg, l.source, nil)
}

// Info implementa la interfaz Logger
func (l *DBLogger) Info(args ...interface{}) {
	l.Logger.Info(args...)
	
	msg := formatArgs(args...)
	_ = l.persister.CreateLog(models.LogLevelInfo, msg, l.source, nil)
}

// Warn implementa la interfaz Logger
func (l *DBLogger) Warn(args ...interface{}) {
	l.Logger.Warn(args...)
	
	msg := formatArgs(args...)
	_ = l.persister.CreateLog(models.LogLevelWarn, msg, l.source, nil)
}

// Error implementa la interfaz Logger
func (l *DBLogger) Error(args ...interface{}) {
	l.Logger.Error(args...)
	
	msg := formatArgs(args...)
	_ = l.persister.CreateLog(models.LogLevelError, msg, l.source, nil)
}

// Fatal implementa la interfaz Logger
func (l *DBLogger) Fatal(args ...interface{}) {
	l.Logger.Fatal(args...)
	
	msg := formatArgs(args...)
	_ = l.persister.CreateLog(models.LogLevelFatal, msg, l.source, nil)
}

// Debugf implementa la interfaz Logger
func (l *DBLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
	
	msg := formatf(format, args...)
	_ = l.persister.CreateLog(models.LogLevelDebug, msg, l.source, nil)
}

// Infof implementa la interfaz Logger
func (l *DBLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
	
	msg := formatf(format, args...)
	_ = l.persister.CreateLog(models.LogLevelInfo, msg, l.source, nil)
}

// Warnf implementa la interfaz Logger
func (l *DBLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
	
	msg := formatf(format, args...)
	_ = l.persister.CreateLog(models.LogLevelWarn, msg, l.source, nil)
}

// Errorf implementa la interfaz Logger
func (l *DBLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
	
	msg := formatf(format, args...)
	_ = l.persister.CreateLog(models.LogLevelError, msg, l.source, nil)
}

// Fatalf implementa la interfaz Logger
func (l *DBLogger) Fatalf(format string, args ...interface{}) {
	l.Logger.Fatalf(format, args...)
	
	msg := formatf(format, args...)
	_ = l.persister.CreateLog(models.LogLevelFatal, msg, l.source, nil)
}

// WithField implementa la interfaz Logger
func (l *DBLogger) WithField(key string, value interface{}) Logger {
	return &DBLogger{
		Logger:    l.Logger.WithField(key, value),
		persister: l.persister,
		source:    l.source,
	}
}

// WithFields implementa la interfaz Logger
func (l *DBLogger) WithFields(fields map[string]interface{}) Logger {
	metadata := models.Metadata(fields)
	
	return &DBLoggerWithMetadata{
		DBLogger: DBLogger{
			Logger:    l.Logger.WithFields(fields),
			persister: l.persister,
			source:    l.source,
		},
		metadata: metadata,
	}
}

// DBLoggerWithMetadata extiende DBLogger con metadatos
type DBLoggerWithMetadata struct {
	DBLogger
	metadata models.Metadata
}

// withMetadata persiste los logs con los metadatos
func (l *DBLoggerWithMetadata) withMetadata(level models.LogLevel, msg string) {
	_ = l.persister.CreateLog(level, msg, l.source, l.metadata)
}

// Debug implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Debug(args ...interface{}) {
	l.DBLogger.Logger.Debug(args...)
	msg := formatArgs(args...)
	l.withMetadata(models.LogLevelDebug, msg)
}

// Info implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Info(args ...interface{}) {
	l.DBLogger.Logger.Info(args...)
	msg := formatArgs(args...)
	l.withMetadata(models.LogLevelInfo, msg)
}

// Warn implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Warn(args ...interface{}) {
	l.DBLogger.Logger.Warn(args...)
	msg := formatArgs(args...)
	l.withMetadata(models.LogLevelWarn, msg)
}

// Error implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Error(args ...interface{}) {
	l.DBLogger.Logger.Error(args...)
	msg := formatArgs(args...)
	l.withMetadata(models.LogLevelError, msg)
}

// Fatal implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Fatal(args ...interface{}) {
	l.DBLogger.Logger.Fatal(args...)
	msg := formatArgs(args...)
	l.withMetadata(models.LogLevelFatal, msg)
}

// Debugf implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Debugf(format string, args ...interface{}) {
	l.DBLogger.Logger.Debugf(format, args...)
	msg := formatf(format, args...)
	l.withMetadata(models.LogLevelDebug, msg)
}

// Infof implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Infof(format string, args ...interface{}) {
	l.DBLogger.Logger.Infof(format, args...)
	msg := formatf(format, args...)
	l.withMetadata(models.LogLevelInfo, msg)
}

// Warnf implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Warnf(format string, args ...interface{}) {
	l.DBLogger.Logger.Warnf(format, args...)
	msg := formatf(format, args...)
	l.withMetadata(models.LogLevelWarn, msg)
}

// Errorf implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Errorf(format string, args ...interface{}) {
	l.DBLogger.Logger.Errorf(format, args...)
	msg := formatf(format, args...)
	l.withMetadata(models.LogLevelError, msg)
}

// Fatalf implementa la interfaz Logger
func (l *DBLoggerWithMetadata) Fatalf(format string, args ...interface{}) {
	l.DBLogger.Logger.Fatalf(format, args...)
	msg := formatf(format, args...)
	l.withMetadata(models.LogLevelFatal, msg)
} 