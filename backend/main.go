package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jminat01/dashboard-servers-go/backend/config"
	"github.com/jminat01/dashboard-servers-go/backend/internal/handlers"
	"github.com/jminat01/dashboard-servers-go/backend/internal/middleware"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/database"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/notifications"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/websocket"
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
		&models.Log{},            // Añadir tabla de logs
		&models.User{},           // Añadir tabla de usuarios
		&models.Alert{},          // Añadir tabla de alertas
		&models.AlertThreshold{}, // Añadir tabla de umbrales de alertas
	); err != nil {
		log.Fatalf("Error en la migración automática: %v", err)
	}

	// Inicializar servicios
	logService := services.NewLogService(db.DB, log)

	// Una vez que tenemos el servicio de logs, podemos crear un logger con persistencia en BD
	log = logger.NewDBLogger(log, logService, "system")
	log.Info("Sistema de logging en base de datos inicializado")

	// Configurar cliente Redis si está habilitado
	var redisClient *redis.Client
	if cfg.Redis.Enabled {
		redisConfig := &websocket.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			Enabled:  cfg.Redis.Enabled,
		}

		redisClient, err = websocket.NewRedisClient(redisConfig, log)
		if err != nil {
			log.Warnf("Error al conectar con Redis: %v. Continuando sin soporte de Redis.", err)
		}
	}

	// Inicializar el Hub de WebSockets
	wsHub := websocket.NewHub(log, redisClient)
	go wsHub.Run() // Iniciar en una goroutine

	// Inicializar manejador de notificaciones (para alertas)
	notifyConfig := &notifications.NotificationConfig{
		DiscordEnabled:    cfg.Notifications.DiscordEnabled,
		DiscordWebhookURL: cfg.Notifications.DiscordWebhookURL,
		EmailEnabled:      cfg.Notifications.EmailEnabled,
		SMTPServer:        cfg.Notifications.EmailSMTP,
		SMTPPort:          cfg.Notifications.EmailPort,
		SMTPUser:          cfg.Notifications.EmailUser,
		SMTPPassword:      cfg.Notifications.EmailPassword,
		EmailFrom:         cfg.Notifications.EmailFrom,
	}
	notificationManager := notifications.NewNotificationManager(notifyConfig, log)

	// Inicializar resto de servicios con el nuevo logger
	serverService := services.NewServerService(db.DB, log)
	metricService := services.NewMetricService(db.DB, log, wsHub, redisClient)
	userService := services.NewUserService(db.DB, log)
	authService := services.NewAuthService(db.DB, log, cfg.Auth.JWTSecret, 86400) // Token válido por 24 horas en segundos
	alertService := services.NewAlertService(db.DB, log, notificationManager)

	// Configurar las dependencias circulares entre servicios
	metricService.SetAlertService(alertService)
	alertService.SetMetricService(metricService)

	// Crear usuario admin por defecto si no existe
	createDefaultAdmin(userService, log, cfg)

	// Inicializar middleware de autenticación
	authMiddleware := middleware.NewAuthMiddleware(authService, log)
	wsAuthMiddleware := websocket.NewWSAuthMiddleware(authService, log)

	// Inicializar handlers
	serverHandler := handlers.NewServerHandler(serverService, log)
	metricHandler := handlers.NewMetricHandler(metricService, serverService, log, wsAuthMiddleware, wsHub)
	logHandler := handlers.NewLogHandler(logService, log)
	authHandler := handlers.NewAuthHandler(authService, userService, log)
	userHandler := handlers.NewUserHandler(userService, log)
	alertHandler := handlers.NewAlertHandler(alertService, log)

	// Configurar router
	router := gin.Default()

	// Middleware para CORS
	router.Use(func(c *gin.Context) {
		// Usar el origen específico en lugar de "*"
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5500")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

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
	authHandler.RegisterRoutes(router, authMiddleware)
	userHandler.RegisterRoutes(router, authMiddleware)

	// Registrar rutas protegidas con autenticación
	serverRoutes := router.Group("/api")
	serverRoutes.Use(authMiddleware.RequireAuth())

	// Registrar rutas de servidores y métricas (requieren autenticación)
	serverHandler.RegisterRoutes(serverRoutes)
	metricHandler.RegisterRoutes(serverRoutes)

	// Ruta de logs (requiere rol de admin)
	logRoutes := router.Group("/api")
	logRoutes.Use(authMiddleware.RequireAuth())
	logRoutes.Use(authMiddleware.RequireRole(models.RoleAdmin))
	logHandler.RegisterRoutes(logRoutes)

	// Registrar rutas de alertas
	alertRoutes := router.Group("/api")
	alertRoutes.Use(authMiddleware.RequireAuth())
	alertHandler.RegisterRoutes(alertRoutes, authMiddleware)

	// Manejar señales para apagado graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar el servidor en una goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Infof("Servidor iniciado en http://localhost%s", addr)
		log.Infof("WebSockets disponibles en ws://localhost%s/api/metrics/live/{server_id}", addr)

		if err := router.Run(addr); err != nil {
			log.Fatalf("Error al iniciar servidor: %v", err)
		}
	}()

	// Bloquear hasta que se reciba una señal de terminación
	<-quit
	log.Info("Apagando servidor...")

	// Detener el hub de WebSockets
	if wsHub != nil {
		wsHub.Stop()
	}

	// Cerrar conexión a la base de datos
	if err := db.Close(); err != nil {
		log.Errorf("Error al cerrar la conexión a la base de datos: %v", err)
	}

	// Cerrar conexión a Redis si estaba activa
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			log.Errorf("Error al cerrar la conexión a Redis: %v", err)
		}
	}

	log.Info("Servidor apagado exitosamente")
}

// createDefaultAdmin crea un usuario administrador por defecto si no existe
func createDefaultAdmin(userService *services.UserService, log logger.Logger, cfg *config.Config) {
	_, err := userService.GetUserByUsername("admin")

	// Si el usuario no existe, crearlo
	if err != nil {
		log.Info("Creando usuario administrador por defecto...")

		// Usar contraseña de configuración o una por defecto
		password := cfg.Auth.DefaultAdminPassword
		if password == "" {
			password = "admin123" // Contraseña por defecto
			log.Warn("Usando contraseña por defecto para admin. Se recomienda cambiarla inmediatamente.")
		}

		admin := &models.User{
			Username: "admin",
			Email:    "admin@sistema.local",
			Role:     models.RoleAdmin,
		}

		if err := userService.CreateUser(admin, password); err != nil {
			log.Errorf("Error al crear usuario admin por defecto: %v", err)
			return
		}

		log.Info("Usuario administrador creado exitosamente. Username: admin")
	} else {
		log.Info("Usuario administrador ya existe, omitiendo creación")
	}
}
