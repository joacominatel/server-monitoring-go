package websocket

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/interfaces"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// Configuración para el upgrade de WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // En producción, esto debería ser más restrictivo
	},
}

// WSAuthMiddleware proporciona autenticación para conexiones WebSocket
type WSAuthMiddleware struct {
	authService interfaces.AuthServiceInterface
	log         logger.Logger
}

// NewWSAuthMiddleware crea un nuevo middleware de autenticación WebSocket
func NewWSAuthMiddleware(authService interfaces.AuthServiceInterface, log logger.Logger) *WSAuthMiddleware {
	return &WSAuthMiddleware{
		authService: authService,
		log:         log,
	}
}

// Authenticate verifica el token JWT en el header de la petición
func (m *WSAuthMiddleware) Authenticate(c *gin.Context) (uint, bool) {
	// Extraer token del header Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		// Intentar extraer token de query param para WebSockets
		token := c.Query("token")
		if token == "" {
			m.log.Warn("No se encontró token de autenticación")
			return 0, false
		}
		authHeader = "Bearer " + token
	}

	// Verificar formato correcto
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		m.log.Warn("Formato de token inválido")
		return 0, false
	}

	// Validar token
	claims, err := m.authService.ValidateToken(parts[1])
	if err != nil {
		m.log.Warnf("Token inválido: %v", err)
		return 0, false
	}

	// Extraer ID de usuario
	userID, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil {
		m.log.Warnf("Error al extraer userID del token: %v", err)
		return 0, false
	}

	return uint(userID), true
}

// HandleWSConnection gestiona una conexión WebSocket autenticada
func HandleWSConnection(c *gin.Context, hub *Hub, authMiddleware *WSAuthMiddleware, serverService interfaces.ServerServiceInterface) {
	// Autenticar al usuario
	userID, authenticated := authMiddleware.Authenticate(c)
	if !authenticated {
		c.String(http.StatusUnauthorized, "No autorizado")
		return
	}

	// Extraer serverID de los parámetros
	serverIDStr := c.Param("server_id")
	serverID, err := strconv.ParseUint(serverIDStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID de servidor inválido")
		return
	}

	// Verificar que el servidor existe
	_, err = serverService.GetServerByID(uint(serverID))
	if err != nil {
		c.String(http.StatusNotFound, "Servidor no encontrado")
		return
	}

	// Actualizar a conexión WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		authMiddleware.log.Errorf("Error al actualizar a WebSocket: %v", err)
		return
	}

	// Crear nuevo cliente
	serverID64 := uint(serverID)
	client := NewClient(hub, conn, serverID64, userID, authMiddleware.log)
	hub.register <- client

	// Permitir recolección del cliente cuando las goroutines terminen
	go client.WritePump()
	go client.ReadPump()
} 