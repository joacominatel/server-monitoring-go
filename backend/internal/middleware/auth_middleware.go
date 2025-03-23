package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/internal/services"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// Constantes para cookies y contexto
const (
	AuthCookieName = "auth_token"
	UserContextKey = "user"
	UserIDKey      = "user_id"
	UserRoleKey    = "user_role"
)

// AuthMiddleware gestiona la autenticación mediante cookies JWT
type AuthMiddleware struct {
	authService *services.AuthService
	logger      logger.Logger
}

// NewAuthMiddleware crea una nueva instancia del middleware de autenticación
func NewAuthMiddleware(authService *services.AuthService, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware para verificar si el usuario está autenticado
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extraer token de la cookie
		tokenCookie, err := c.Cookie(AuthCookieName)
		if err != nil {
			m.logger.Warnf("Acceso no autorizado: cookie no encontrada")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
			c.Abort()
			return
		}

		// Verificar token
		claims, err := m.authService.VerifyToken(tokenCookie)
		if err != nil {
			m.logger.Warnf("Token inválido: %v", err)

			// Si el token expiró, eliminar la cookie
			if err == services.ErrTokenExpired {
				c.SetCookie(AuthCookieName, "", -1, "/", "", false, true)
			}

			c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
			c.Abort()
			return
		}

		// Almacenar información del usuario en el contexto
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)

		c.Next()
	}
}

// RequireRole middleware para verificar si el usuario tiene alguno de los roles requeridos
func (m *AuthMiddleware) RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Primero verificar autenticación
		userIDValue, exists := c.Get(UserIDKey)
		if !exists {
			m.logger.Warn("Verificación de rol fallida: usuario no autenticado")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
			c.Abort()
			return
		}

		userID, ok := userIDValue.(uint)
		if !ok {
			m.logger.Warn("Verificación de rol fallida: ID de usuario inválido")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
			c.Abort()
			return
		}

		// Obtener el rol del usuario del contexto
		userRoleValue, exists := c.Get(UserRoleKey)
		if !exists {
			m.logger.Warn("Verificación de rol fallida: rol no encontrado en contexto")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
			c.Abort()
			return
		}

		userRole, ok := userRoleValue.(models.Role)
		if !ok {
			m.logger.Warn("Verificación de rol fallida: tipo de rol inválido")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error interno del servidor"})
			c.Abort()
			return
		}

		// Verificar si el usuario tiene alguno de los roles requeridos
		hasRequiredRole := false
		for _, role := range roles {
			if userRole == role {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			m.logger.Warnf("Acceso denegado: usuario %d con rol %s no tiene los roles requeridos", userID, userRole)
			c.JSON(http.StatusForbidden, gin.H{"error": "Acceso denegado"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserID obtiene el ID del usuario autenticado desde el contexto
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUserRole obtiene el rol del usuario autenticado desde el contexto
func GetUserRole(c *gin.Context) (models.Role, bool) {
	userRole, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}

	role, ok := userRole.(models.Role)
	return role, ok
}
