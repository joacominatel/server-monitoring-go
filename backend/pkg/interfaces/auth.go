package interfaces

// TokenClaims representa las claims en un token JWT
type TokenClaims struct {
	Subject string
}

// AuthServiceInterface define los métodos que debe implementar un servicio de autenticación
type AuthServiceInterface interface {
	ValidateToken(token string) (*TokenClaims, error)
}

// ServerServiceInterface define los métodos que debe implementar un servicio de servidores
type ServerServiceInterface interface {
	GetServerByID(id uint) (interface{}, error)
}
