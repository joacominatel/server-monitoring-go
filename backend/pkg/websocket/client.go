package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

const (
	// Tiempo para escribir un mensaje al cliente
	writeWait = 10 * time.Second

	// Tiempo para leer el próximo pong del cliente
	pongWait = 60 * time.Second

	// Enviar pings al cliente con esta periodicidad
	pingPeriod = (pongWait * 9) / 10

	// Tamaño máximo del mensaje
	maxMessageSize = 512
)

// Cliente representa una conexión WebSocket individual
type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	serverID  uint
	userID    uint
	mu        sync.Mutex
	log       logger.Logger
	closed    bool
	closeChan chan struct{}
}

// NewClient crea un nuevo cliente WebSocket
func NewClient(hub *Hub, conn *websocket.Conn, serverID, userID uint, log logger.Logger) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		serverID:  serverID,
		userID:    userID,
		log:       log,
		closed:    false,
		closeChan: make(chan struct{}),
	}
}

// ReadPump bombea mensajes desde la conexión WebSocket al hub
func (c *Client) ReadPump() {
	defer func() {
		c.Close()
	}()
	
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Errorf("Error de lectura WebSocket: %v", err)
			}
			break
		}
		// Actualmente no necesitamos manejar mensajes entrantes de los clientes
	}
}

// WritePump bombea mensajes desde el hub a la conexión WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// El hub cerró el canal
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Añadir mensajes en cola al actual
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.closeChan:
			return
		}
	}
}

// SendMetric envía una métrica al cliente
func (c *Client) SendMetric(metric interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	data, err := json.Marshal(metric)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
	default:
		c.log.Warnf("Buffer de cliente lleno, descartando mensaje para serverID %d", c.serverID)
	}
	
	return nil
}

// Close cierra el cliente y limpia recursos
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.closed {
		c.closed = true
		close(c.closeChan)
		c.hub.unregister <- c
		c.conn.Close()
	}
} 