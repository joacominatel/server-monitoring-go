package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// Hub mantiene el conjunto de clientes activos y transmite mensajes
type Hub struct {
	// Clientes registrados
	clients sync.Map

	// Canal para registrar clientes
	register chan *Client

	// Canal para dar de baja clientes
	unregister chan *Client

	// Logger para eventos del hub
	log logger.Logger

	// Cliente Redis para Pub/Sub
	redisClient *redis.Client

	// Canal para detener el hub
	stopChan chan struct{}

	// Mutex para operaciones internas
	mu sync.Mutex

	// Contexto para Redis
	ctx context.Context
}

// NewHub crea un nuevo hub
func NewHub(log logger.Logger, redisClient *redis.Client) *Hub {
	hub := &Hub{
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		log:         log,
		redisClient: redisClient,
		stopChan:    make(chan struct{}),
		ctx:         context.Background(),
	}
	return hub
}

// Run inicia el hub y maneja operaciones en canales
func (h *Hub) Run() {
	h.log.Info("Iniciando hub de WebSockets")
	
	// Si tenemos Redis configurado, suscribirnos a los canales de métricas
	if h.redisClient != nil {
		go h.subscribeToRedis()
	}

	for {
		select {
		case client := <-h.register:
			// Almacenar clientes por serverID para envío eficiente
			serverClients, _ := h.clients.LoadOrStore(client.serverID, &sync.Map{})
			serverClients.(*sync.Map).Store(client, true)
			h.log.Infof("Cliente registrado para serverID: %d, userID: %d", client.serverID, client.userID)
		
		case client := <-h.unregister:
			h.mu.Lock()
			// Encontrar el mapa de clientes para este servidor
			if serverClients, ok := h.clients.Load(client.serverID); ok {
				// Eliminar este cliente del mapa
				serverClients.(*sync.Map).Delete(client)
				
				// Si no quedan clientes para este servidor, eliminar el mapa
				empty := true
				serverClients.(*sync.Map).Range(func(_, _ interface{}) bool {
					empty = false
					return false
				})
				
				if empty {
					h.clients.Delete(client.serverID)
				}
				
				close(client.send)
			}
			h.mu.Unlock()
			h.log.Infof("Cliente desregistrado para serverID: %d", client.serverID)
			
		case <-h.stopChan:
			return
		}
	}
}

// Stop detiene el hub
func (h *Hub) Stop() {
	h.log.Info("Deteniendo hub de WebSockets")
	close(h.stopChan)
}

// BroadcastToServer envía un mensaje a todos los clientes conectados a un servidor específico
func (h *Hub) BroadcastToServer(serverID uint, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.log.Errorf("Error al serializar mensaje: %v", err)
		return
	}

	// Si tenemos Redis configurado, publicar el mensaje
	if h.redisClient != nil {
		channel := getServerChannel(serverID)
		if err := h.redisClient.Publish(h.ctx, channel, data).Err(); err != nil {
			h.log.Errorf("Error al publicar mensaje en Redis: %v", err)
		}
		return
	}

	// Si no tenemos Redis, enviar directamente a los clientes
	h.broadcastToServerDirect(serverID, data)
}

// broadcastToServerDirect envía un mensaje directamente a los clientes conectados
func (h *Hub) broadcastToServerDirect(serverID uint, data []byte) {
	// Obtener todos los clientes para este servidor
	if serverClients, ok := h.clients.Load(serverID); ok {
		serverClients.(*sync.Map).Range(func(key, _ interface{}) bool {
			client := key.(*Client)
			select {
			case client.send <- data:
			default:
				h.mu.Lock()
				serverClients.(*sync.Map).Delete(client)
				close(client.send)
				h.mu.Unlock()
			}
			return true
		})
	}
}

// subscribeToRedis suscribe al hub a canales Redis para recibir métricas
func (h *Hub) subscribeToRedis() {
	pubsub := h.redisClient.PSubscribe(h.ctx, "metrics:server:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			// Extraer serverID del canal
			serverID := parseServerIDFromChannel(msg.Channel)
			if serverID == 0 {
				h.log.Errorf("No se pudo extraer serverID del canal: %s", msg.Channel)
				continue
			}

			// Enviar el mensaje a todos los clientes de este servidor
			h.broadcastToServerDirect(serverID, []byte(msg.Payload))

		case <-h.stopChan:
			return
		}
	}
}

// Funciones auxiliares para manejar nombres de canales Redis
func getServerChannel(serverID uint) string {
	return "metrics:server:" + uintToString(serverID)
}

func parseServerIDFromChannel(channel string) uint {
	// Implementar lógica para extraer serverID del nombre del canal
	// Por ejemplo, de "metrics:server:123" extraer 123
	var serverID uint
	_, err := fmt.Sscanf(channel, "metrics:server:%d", &serverID)
	if err != nil {
		return 0
	}
	return serverID
}

// uintToString convierte uint a string
func uintToString(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
} 