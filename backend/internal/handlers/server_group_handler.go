package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/internal/middleware"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// ServerGroupHandler manejador para las rutas de grupos de servidores
type ServerGroupHandler struct {
	service *services.ServerGroupService
	logger  logger.Logger
}

// NewServerGroupHandler crea un nuevo manejador para grupos de servidores
func NewServerGroupHandler(service *services.ServerGroupService, log logger.Logger) *ServerGroupHandler {
	return &ServerGroupHandler{
		service: service,
		logger:  log,
	}
}

// RegisterRoutes registra las rutas relacionadas con grupos
func (h *ServerGroupHandler) RegisterRoutes(router gin.IRouter, authMiddleware *middleware.AuthMiddleware) {
	groups := router.Group("/server-groups")
	{
		// Rutas accesibles a todos los usuarios autenticados
		groups.GET("", h.GetAllGroups)
		groups.GET("/tree", h.GetGroupTree)
		groups.GET("/:id", h.GetGroupByID)
		groups.GET("/:id/servers", h.GetServersInGroup)

		// Rutas que requieren rol de admin o user
		adminOrUser := groups.Group("")
		adminOrUser.Use(authMiddleware.RequireRole(models.RoleAdmin, models.RoleUser))
		{
			adminOrUser.POST("", h.CreateGroup)
			adminOrUser.PUT("/:id", h.UpdateGroup)
			adminOrUser.DELETE("/:id", h.DeleteGroup)
			adminOrUser.POST("/:id/servers/:server_id", h.AddServerToGroup)
			adminOrUser.DELETE("/:id/servers/:server_id", h.RemoveServerFromGroup)
		}
	}
}

// GetAllGroups obtiene todos los grupos
func (h *ServerGroupHandler) GetAllGroups(c *gin.Context) {
	includeChildren := c.Query("include_children") == "true"
	includeServers := c.Query("include_servers") == "true"

	groups, err := h.service.GetAllGroups(includeChildren, includeServers)
	if err != nil {
		h.logger.Errorf("Error al obtener grupos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener grupos"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// GetGroupTree obtiene el árbol jerárquico de grupos
func (h *ServerGroupHandler) GetGroupTree(c *gin.Context) {
	includeServers := c.Query("include_servers") == "true"

	// Obtener solo los grupos raíz con sus hijos
	groups, err := h.service.GetRootGroups(true, includeServers)
	if err != nil {
		h.logger.Errorf("Error al obtener árbol de grupos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener árbol de grupos"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

// GetGroupByID obtiene un grupo por su ID
func (h *ServerGroupHandler) GetGroupByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	includeChildren := c.Query("include_children") == "true"
	includeServers := c.Query("include_servers") == "true"

	group, err := h.service.GetGroup(uint(id), includeChildren, includeServers)
	if err != nil {
		h.logger.Errorf("Error al obtener grupo %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Grupo no encontrado"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// CreateGroup crea un nuevo grupo
func (h *ServerGroupHandler) CreateGroup(c *gin.Context) {
	var group models.ServerGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		h.logger.Warnf("Error en binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	// Obtener el ID del usuario del contexto
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Usuario no identificado"})
		return
	}
	group.CreatedBy = userID.(uint)

	if err := h.service.CreateGroup(&group); err != nil {
		h.logger.Errorf("Error al crear grupo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// UpdateGroup actualiza un grupo existente
func (h *ServerGroupHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var group models.ServerGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		h.logger.Warnf("Error en binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	group.ID = uint(id)

	if err := h.service.UpdateGroup(&group); err != nil {
		h.logger.Errorf("Error al actualizar grupo %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup elimina un grupo
func (h *ServerGroupHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.service.DeleteGroup(uint(id)); err != nil {
		h.logger.Errorf("Error al eliminar grupo %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grupo eliminado correctamente"})
}

// GetServersInGroup obtiene los servidores en un grupo
func (h *ServerGroupHandler) GetServersInGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	group, err := h.service.GetGroup(uint(id), false, true)
	if err != nil {
		h.logger.Errorf("Error al obtener servidores del grupo %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Grupo no encontrado"})
		return
	}

	c.JSON(http.StatusOK, group.Servers)
}

// AddServerToGroup añade un servidor a un grupo
func (h *ServerGroupHandler) AddServerToGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de grupo inválido"})
		return
	}

	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	if err := h.service.AddServerToGroup(uint(groupID), uint(serverID)); err != nil {
		h.logger.Errorf("Error al añadir servidor %d al grupo %d: %v", serverID, groupID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Servidor añadido al grupo correctamente"})
}

// RemoveServerFromGroup elimina un servidor de un grupo
func (h *ServerGroupHandler) RemoveServerFromGroup(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de grupo inválido"})
		return
	}

	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de servidor inválido"})
		return
	}

	if err := h.service.RemoveServerFromGroup(uint(groupID), uint(serverID)); err != nil {
		h.logger.Errorf("Error al eliminar servidor %d del grupo %d: %v", serverID, groupID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Servidor eliminado del grupo correctamente"})
}
