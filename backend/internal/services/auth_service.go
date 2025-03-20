package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
)

// Errores relacionados con autenticación
var (
	ErrInvalidCredentials = errors.New("credenciales inválidas")
	ErrUserNotFound       = errors.New("usuario no encontrado")
	ErrTokenInvalid       = errors.New("token inválido o expirado")
	ErrTokenExpired       = errors.New("token expirado")
	ErrUserDisabled       = errors.New("usuario deshabilitado")
	ErrInsufficientRole   = errors.New("permisos insuficientes")
)

// JWTClaims contiene los claims del token JWT
type JWTClaims struct {
	UserID   uint        `json:"user_id"`
	Username string      `json:"username"`
	Email    string      `json:"email"`
	Role     models.Role `json:"role"`
	jwt.RegisteredClaims
}

// AuthService maneja la autenticación de usuarios
type AuthService struct {
	db         *gorm.DB
	logger     logger.Logger
	jwtSecret  []byte
	jwtExpires time.Duration
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(db *gorm.DB, logger logger.Logger, jwtSecret string, jwtExpireHours int) *AuthService {
	if jwtExpireHours <= 0 {
		jwtExpireHours = 24 // Por defecto 24 horas
	}
	
	return &AuthService{
		db:         db,
		logger:     logger,
		jwtSecret:  []byte(jwtSecret),
		jwtExpires: time.Duration(jwtExpireHours) * time.Hour,
	}
}

// Login autentica un usuario y devuelve un token JWT
func (s *AuthService) Login(username, password string) (string, *models.User, error) {
	var user models.User
	
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Intento de login para usuario inexistente: %s", username)
			return "", nil, ErrInvalidCredentials
		}
		s.logger.Errorf("Error al buscar usuario en BD: %v", err)
		return "", nil, err
	}
	
	if !user.CheckPassword(password) {
		s.logger.Warnf("Contraseña incorrecta para usuario: %s", username)
		return "", nil, ErrInvalidCredentials
	}
	
	// Actualizar último login
	now := time.Now()
	user.LastLogin = &now
	if err := s.db.Save(&user).Error; err != nil {
		s.logger.Warnf("Error al actualizar último login: %v", err)
		// No devolver error para no interrumpir el login
	}
	
	// Generar token JWT
	token, err := s.GenerateToken(&user)
	if err != nil {
		s.logger.Errorf("Error al generar token JWT: %v", err)
		return "", nil, err
	}
	
	s.logger.Infof("Login exitoso para usuario: %s", username)
	return token, &user, nil
}

// GenerateToken genera un token JWT para un usuario
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(s.jwtExpires)
	
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

// VerifyToken verifica y analiza un token JWT
func (s *AuthService) VerifyToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	
	if !token.Valid {
		return nil, ErrTokenInvalid
	}
	
	return claims, nil
}

// GetUserByID obtiene un usuario por su ID
func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	
	if err := s.db.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// CheckUserRole verifica si un usuario tiene el rol mínimo requerido
func (s *AuthService) CheckUserRole(userID uint, minRole models.Role) error {
	user, err := s.GetUserByID(userID)
	if err != nil {
		return err
	}
	
	if !user.CanAccess(minRole) {
		return ErrInsufficientRole
	}
	
	return nil
} 