package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Database      DatabaseConfig
	Server        ServerConfig
	Auth          AuthConfig
	Redis         RedisConfig
	WebSocket     WebSocketConfig
	Notifications NotificationsConfig
}

// DatabaseConfig contiene la configuración de la base de datos
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ServerConfig contiene la configuración del servidor
type ServerConfig struct {
	Port string
	Env  string
}

// AuthConfig contiene la configuración de autenticación
type AuthConfig struct {
	JWTSecret            string
	DefaultAdminPassword string
}

// RedisConfig contiene la configuración de Redis para Pub/Sub
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Enabled  bool
}

// WebSocketConfig contiene la configuración del servidor WebSocket
type WebSocketConfig struct {
	PingInterval   int // Intervalo de ping en segundos
	AllowedOrigins []string
}

// NotificationsConfig contiene la configuración para las notificaciones
type NotificationsConfig struct {
	EmailEnabled      bool
	EmailFrom         string
	EmailSMTP         string
	EmailPort         int
	EmailUser         string
	EmailPassword     string
	DiscordEnabled    bool
	DiscordWebhookURL string
}

// LoadConfig carga la configuración desde el archivo .env
func LoadConfig() (*Config, error) {
	// Intentar cargar archivo .env
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar el archivo .env, utilizando variables de entorno del sistema")
	}

	viper.AutomaticEnv()

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "server_monitoring"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Auth: AuthConfig{
			JWTSecret:            getEnv("JWT_SECRET", "mi_clave_secreta_jwt_para_desarrollo"),
			DefaultAdminPassword: getEnv("ADMIN_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnvAsBool("REDIS_ENABLED", false),
		},
		WebSocket: WebSocketConfig{
			PingInterval:   getEnvAsInt("WS_PING_INTERVAL", 30),
			AllowedOrigins: getEnvAsStringSlice("WS_ALLOWED_ORIGINS", []string{"*"}),
		},
		Notifications: NotificationsConfig{
			EmailEnabled:      getEnvAsBool("EMAIL_ENABLED", false),
			EmailFrom:         getEnv("EMAIL_FROM", "alertas@sistema.local"),
			EmailSMTP:         getEnv("EMAIL_SMTP", "smtp.example.com"),
			EmailPort:         getEnvAsInt("EMAIL_PORT", 587),
			EmailUser:         getEnv("EMAIL_USER", ""),
			EmailPassword:     getEnv("EMAIL_PASSWORD", ""),
			DiscordEnabled:    getEnvAsBool("DISCORD_ENABLED", false),
			DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		},
	}

	return config, nil
}

// GetDSN retorna la cadena de conexión para PostgreSQL
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Función auxiliar para obtener variables de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt obtiene variable de entorno como entero
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	if err != nil || intValue != defaultValue {
		return defaultValue
	}

	return intValue
}

// getEnvAsBool obtiene variable de entorno como booleano
func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value == "true" || value == "1" || value == "yes" || value == "y"
}

// getEnvAsStringSlice obtiene variable de entorno como slice de strings
func getEnvAsStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Implementación simple para separar por comas
	return strings.Split(value, ",")
}
