package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/config"
	"github.com/jminat01/dashboard-servers-go/backend/internal/handlers"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/database"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

func main() {
	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error al cargar configuración: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger básico para el arranque
	log := logger.NewLogger(cfg.Server.Env)
	log.Info("Iniciando servidor de monitoreo...")

	// Establecer modo de Gin según el entorno
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Conectar a la base de datos
	db, err := database.NewDatabase(cfg.Database.GetDSN(), log)
	if err != nil {
		log.Fatalf("Error al conectar a la base de datos: %v", err)
	}

	// Migrar modelos
	if err := db.AutoMigrate(
		&models.Server{}, 
		&models.Metric{},
		&models.Log{}, // Añadir tabla de logs
	); err != nil {
		log.Fatalf("Error en la migración automática: %v", err)
	}

	// Inicializar servicios
	logService := services.NewLogService(db.DB, log)
	
	// Una vez que tenemos el servicio de logs, podemos crear un logger con persistencia en BD
	log = logger.NewDBLogger(log, logService, "system")
	log.Info("Sistema de logging en base de datos inicializado")

	// Inicializar resto de servicios con el nuevo logger
	serverService := services.NewServerService(db.DB, log)
	metricService := services.NewMetricService(db.DB, log)

	// Inicializar handlers
	serverHandler := handlers.NewServerHandler(serverService, log)
	metricHandler := handlers.NewMetricHandler(metricService, serverService, log)
	logHandler := handlers.NewLogHandler(logService, log)

	// Configurar router
	router := gin.Default()

	// Middleware para CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Endpoint de salud
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// Registrar rutas
	serverHandler.RegisterRoutes(router)
	metricHandler.RegisterRoutes(router)
	logHandler.RegisterRoutes(router) // Registrar rutas para logs

	// Manejar señales para apagado graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar el servidor en una goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Infof("Servidor iniciado en http://localhost%s", addr)
		
		if err := router.Run(addr); err != nil {
			log.Fatalf("Error al iniciar servidor: %v", err)
		}
	}()

	// Bloquear hasta que se reciba una señal de terminación
	<-quit
	log.Info("Apagando servidor...")

	// Cerrar conexión a la base de datos
	if err := db.Close(); err != nil {
		log.Errorf("Error al cerrar la conexión a la base de datos: %v", err)
	}

	log.Info("Servidor apagado exitosamente")
} 