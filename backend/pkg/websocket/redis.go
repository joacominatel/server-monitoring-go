package websocket

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// RedisConfig contiene la configuración para el cliente Redis
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Enabled  bool
}

// NewRedisClient crea un nuevo cliente Redis
func NewRedisClient(cfg *RedisConfig, log logger.Logger) (*redis.Client, error) {
	if !cfg.Enabled {
		log.Info("Redis está deshabilitado, no se utilizará Pub/Sub")
		return nil, nil
	}

	log.Info("Iniciando conexión a Redis para Pub/Sub")
	
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Host + ":" + cfg.Port,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		log.Errorf("Error al conectar con Redis: %v", err)
		return nil, err
	}

	log.Info("Conexión a Redis establecida correctamente")
	return client, nil
}

// PublishMetric publica una métrica en Redis
func PublishMetric(client *redis.Client, serverID uint, metric interface{}) error {
	if client == nil {
		return nil
	}

	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	channel := getServerChannel(serverID)
	return client.Publish(context.Background(), channel, data).Err()
} 