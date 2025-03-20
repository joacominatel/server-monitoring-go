package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/internal/middleware"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// Duración de la cookie de autenticación
const (
	cookieDuration = 24 * time.Hour // 24 horas
)

// LoginRequest representa los datos para iniciar sesión
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest representa los datos para registrar un nuevo usuario
type RegisterRequest struct {
	Username string      `json:"username" binding:"required,min=3,max=50"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     models.Role `json:"role"`
}

// ChangePasswordRequest representa los datos para cambiar la contraseña
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// AuthHandler maneja las rutas relacionadas con autenticación
type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	logger      logger.Logger
}

// NewAuthHandler crea una nueva instancia del manejador de autenticación
func NewAuthHandler(authService *services.AuthService, userService *services.UserService, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

// RegisterRoutes registra las rutas del manejador en el router
func (h *AuthHandler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)
		auth.POST("/register", h.Register)
		
		// Rutas protegidas
		protected := auth.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/me", h.GetCurrentUser)
			protected.POST("/change-password", h.ChangePassword)
		}
	}
}

// Login maneja la autenticación de usuarios
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Datos de login inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de inicio de sesión inválidos"})
		return
	}
	
	// Autenticar usuario
	token, user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		statusCode := http.StatusInternalServerError
		
		if err == services.ErrInvalidCredentials {
			statusCode = http.StatusUnauthorized
		}
		
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	
	// Establecer cookie segura
	c.SetCookie(
		middleware.AuthCookieName,
		token,
		int(cookieDuration.Seconds()),
		"/",           // Path
		"",            // Domain (vacío = dominio actual)
		false,         // Secure (en producción debería ser true)
		true,          // HttpOnly (protege contra XSS)
	)
	
	// No enviar la contraseña en la respuesta
	c.JSON(http.StatusOK, gin.H{
		"message": "Login exitoso",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Logout cierra la sesión eliminando la cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	// Establecer una cookie expirada para eliminarla
	c.SetCookie(
		middleware.AuthCookieName,
		"",
		-1,   // MaxAge < 0 elimina la cookie
		"/",
		"",
		false,
		true,
	)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Sesión cerrada exitosamente",
	})
}

// Register registra un nuevo usuario
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Datos de registro inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de registro inválidos"})
		return
	}
	
	// Establecer rol por defecto si no se proporciona
	if req.Role == "" {
		req.Role = models.RoleUser
	}
	
	// Solo admins pueden crear otros admins
	if req.Role == models.RoleAdmin {
		// Obtener el token de la cookie para verificar si es admin
		tokenCookie, err := c.Cookie(middleware.AuthCookieName)
		if err != nil || tokenCookie == "" {
			h.logger.Warn("Intento de crear usuario admin sin autenticación")
			c.JSON(http.StatusForbidden, gin.H{"error": "No autorizado para crear usuarios administradores"})
			return
		}
		
		// Verificar token y rol
		claims, err := h.authService.VerifyToken(tokenCookie)
		if err != nil || claims.Role != models.RoleAdmin {
			h.logger.Warn("Intento de crear usuario admin por un no-admin")
			c.JSON(http.StatusForbidden, gin.H{"error": "No autorizado para crear usuarios administradores"})
			return
		}
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
		"message": "Usuario registrado exitosamente",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// GetCurrentUser obtiene información del usuario autenticado
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}
	
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.Errorf("Error al obtener usuario actual: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener información del usuario"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"last_login": user.LastLogin,
		"created_at": user.CreatedAt,
	})
}

// ChangePassword cambia la contraseña del usuario autenticado
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}
	
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Datos para cambio de contraseña inválidos: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}
	
	// Verificar contraseña actual
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.Errorf("Error al obtener usuario para cambio de contraseña: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
		return
	}
	
	if !user.CheckPassword(req.CurrentPassword) {
		h.logger.Warnf("Contraseña actual incorrecta para usuario ID %d", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contraseña actual incorrecta"})
		return
	}
	
	// Cambiar contraseña
	if err := h.userService.ChangePassword(userID, req.NewPassword); err != nil {
		h.logger.Errorf("Error al cambiar contraseña: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al cambiar contraseña"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Contraseña cambiada exitosamente",
	})
} 