package database

import (
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Database es el objeto que representa la conexión a la base de datos
type Database struct {
	DB     *gorm.DB
	Logger logger.Logger
}

// NewDatabase crea una nueva conexión a la base de datos
func NewDatabase(dsn string, log logger.Logger) (*Database, error) {
	log.Info("Conectando a la base de datos PostgreSQL...")
	
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: NewGormLogger(log),
	}
	
	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		log.Errorf("Error al conectar a la base de datos: %v", err)
		return nil, err
	}
	
	log.Info("Conexión a la base de datos establecida exitosamente")
	
	return &Database{
		DB:     db,
		Logger: log,
	}, nil
}

// AutoMigrate realiza la migración automática de los modelos
func (d *Database) AutoMigrate(models ...interface{}) error {
	d.Logger.Info("Ejecutando migración automática...")
	
	if err := d.DB.AutoMigrate(models...); err != nil {
		d.Logger.Errorf("Error al ejecutar migración automática: %v", err)
		return err
	}
	
	d.Logger.Info("Migración automática completada exitosamente")
	return nil
}

// Close cierra la conexión a la base de datos
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		d.Logger.Errorf("Error al obtener la conexión SQL: %v", err)
		return err
	}
	
	if err := sqlDB.Close(); err != nil {
		d.Logger.Errorf("Error al cerrar la conexión a la base de datos: %v", err)
		return err
	}
	
	d.Logger.Info("Conexión a la base de datos cerrada exitosamente")
	return nil
} 