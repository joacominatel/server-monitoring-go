package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// ServerHandler maneja las rutas relacionadas con servidores
type ServerHandler struct {
	serverService *services.ServerService
	logger        logger.Logger
}

// NewServerHandler crea una nueva instancia del manejador de servidores
func NewServerHandler(serverService *services.ServerService, logger logger.Logger) *ServerHandler {
	return &ServerHandler{
		serverService: serverService,
		logger:        logger,
	}
}

// RegisterRoutes registra las rutas del manejador en el router
func (h *ServerHandler) RegisterRoutes(router gin.IRouter) {
	servers := router.Group("/servers")
	{
		// Rutas que cualquier usuario autenticado puede acceder
		servers.GET("", h.GetAllServers)
		servers.GET("/:id", h.GetServerByID)
		
		// Rutas que requieren rol de admin o usuario normal (no viewer)
		serverAdmin := servers.Group("")
		serverAdmin.Use(func(c *gin.Context) {
			// Aquí iría un middleware que verifica rol, pero esta 
			// lógica ahora está en main.go con RequireRole
			c.Next()
		})
		{
			serverAdmin.POST("", h.CreateServer)
			serverAdmin.PUT("/:id", h.UpdateServer)
			serverAdmin.DELETE("/:id", h.DeleteServer)
		}
	}
}

// GetAllServers obtiene todos los servidores
func (h *ServerHandler) GetAllServers(c *gin.Context) {
	servers, err := h.serverService.GetAllServers()
	if err != nil {
		h.logger.Errorf("Error al obtener servidores: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener servidores"})
		return
	}
	
	c.JSON(http.StatusOK, servers)
}

// GetServerByID obtiene un servidor por su ID
func (h *ServerHandler) GetServerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warnf("ID de servidor inválido: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}
	
	server, err := h.serverService.GetServerByID(uint(id))
	if err != nil {
		h.logger.Errorf("Error al obtener servidor: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Servidor no encontrado"})
		return
	}
	
	c.JSON(http.StatusOK, server)
}

// CreateServer crea un nuevo servidor
func (h *ServerHandler) CreateServer(c *gin.Context) {
	var server models.Server
	
	if err := c.ShouldBindJSON(&server); err != nil {
		h.logger.Warnf("Datos de servidor inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de servidor inválidos"})
		return
	}
	
	if err := h.serverService.CreateServer(&server); err != nil {
		h.logger.Errorf("Error al crear servidor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear servidor"})
		return
	}
	
	c.JSON(http.StatusCreated, server)
}

// UpdateServer actualiza un servidor existente
func (h *ServerHandler) UpdateServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warnf("ID de servidor inválido: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}
	
	// Primero verificar que el servidor existe
	_, err = h.serverService.GetServerByID(uint(id))
	if err != nil {
		h.logger.Errorf("Error al obtener servidor para actualización: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Servidor no encontrado"})
		return
	}
	
	// Vincular los datos de la solicitud
	var serverUpdate models.Server
	if err := c.ShouldBindJSON(&serverUpdate); err != nil {
		h.logger.Warnf("Datos de actualización de servidor inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de servidor inválidos"})
		return
	}
	
	// Asegurar que el ID sea el correcto
	serverUpdate.ID = uint(id)
	
	if err := h.serverService.UpdateServer(&serverUpdate); err != nil {
		h.logger.Errorf("Error al actualizar servidor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar servidor"})
		return
	}
	
	c.JSON(http.StatusOK, serverUpdate)
}

// DeleteServer elimina un servidor
func (h *ServerHandler) DeleteServer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warnf("ID de servidor inválido: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}
	
	if err := h.serverService.DeleteServer(uint(id)); err != nil {
		h.logger.Errorf("Error al eliminar servidor: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al eliminar servidor"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Servidor eliminado exitosamente"})
} 