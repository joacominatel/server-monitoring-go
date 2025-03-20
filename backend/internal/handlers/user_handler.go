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

// UserRequest representa los datos para crear o actualizar un usuario
type UserRequest struct {
	Username string      `json:"username" binding:"required,min=3,max=50"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password"` // No requerido en actualizaciones
	Role     models.Role `json:"role"`
}

// UserHandler maneja las rutas relacionadas con usuarios
type UserHandler struct {
	userService *services.UserService
	logger      logger.Logger
}

// NewUserHandler crea una nueva instancia del manejador de usuarios
func NewUserHandler(userService *services.UserService, logger logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registra las rutas del manejador en el router
func (h *UserHandler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	users := router.Group("/api/users")
	
	// Todas las rutas de usuarios requieren autenticación
	users.Use(authMiddleware.RequireAuth())
	
	// Rutas de administración de usuarios (solo admin)
	users.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		users.GET("", h.GetAllUsers)
		users.GET("/:id", h.GetUser)
		users.POST("", h.CreateUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}
}

// GetAllUsers obtiene todos los usuarios
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		h.logger.Errorf("Error al obtener usuarios: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener usuarios"})
		return
	}
	
	// No enviar contraseñas
	var response []gin.H
	for _, user := range users {
		response = append(response, gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"role":       user.Role,
			"last_login": user.LastLogin,
			"created_at": user.CreatedAt,
		})
	}
	
	c.JSON(http.StatusOK, response)
}

// GetUser obtiene un usuario por su ID
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warnf("ID de usuario inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}
	
	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		h.logger.Errorf("Error al obtener usuario %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"last_login": user.LastLogin,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

// CreateUser crea un nuevo usuario
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req UserRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Datos de usuario inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de usuario inválidos"})
		return
	}
	
	// Verificar que hay contraseña para nuevos usuarios
	if req.Password == "" {
		h.logger.Warn("Intento de crear usuario sin contraseña")
		c.JSON(http.StatusBadRequest, gin.H{"error": "La contraseña es obligatoria para nuevos usuarios"})
		return
	}
	
	// Establecer rol por defecto si no se proporciona
	if req.Role == "" {
		req.Role = models.RoleUser
	}
	
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}
	
	if err := h.userService.CreateUser(user, req.Password); err != nil {
		h.logger.Errorf("Error al crear usuario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Usuario creado exitosamente",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// UpdateUser actualiza un usuario existente
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warnf("ID de usuario inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}
	
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Datos de actualización inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de usuario inválidos"})
		return
	}
	
	// Verificar si el usuario existe
	existingUser, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		h.logger.Errorf("Usuario no encontrado para actualizar: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}
	
	// Actualizar campos del usuario
	existingUser.Username = req.Username
	existingUser.Email = req.Email
	if req.Role != "" {
		existingUser.Role = req.Role
	}
	
	// Actualizar el usuario
	if err := h.userService.UpdateUser(existingUser); err != nil {
		h.logger.Errorf("Error al actualizar usuario: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Si se proporcionó una contraseña, actualizarla
	if req.Password != "" {
		if err := h.userService.ChangePassword(uint(id), req.Password); err != nil {
			h.logger.Errorf("Error al actualizar contraseña: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar contraseña"})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario actualizado exitosamente",
		"user": gin.H{
			"id":       existingUser.ID,
			"username": existingUser.Username,
			"email":    existingUser.Email,
			"role":     existingUser.Role,
		},
	})
}

// DeleteUser elimina un usuario
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.logger.Warnf("ID de usuario inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}
	
	// Obtener ID del usuario autenticado
	currentUserID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}
	
	// Evitar que un usuario se elimine a sí mismo
	if currentUserID == uint(id) {
		h.logger.Warn("Intento de autoeliminar cuenta de usuario")
		c.JSON(http.StatusBadRequest, gin.H{"error": "No puedes eliminar tu propia cuenta"})
		return
	}
	
	if err := h.userService.DeleteUser(uint(id)); err != nil {
		h.logger.Errorf("Error al eliminar usuario %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Usuario eliminado exitosamente",
	})
} 