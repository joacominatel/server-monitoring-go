package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role representa el rol de un usuario en el sistema
type Role string

// Constantes para los roles de usuario
const (
	RoleAdmin  Role = "ADMIN"
	RoleUser   Role = "USER"
	RoleViewer Role = "VIEWER"
)

// User representa un usuario del sistema
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Email     string         `gorm:"size:100;not null;uniqueIndex" json:"email"`
	Password  string         `gorm:"size:100;not null" json:"-"` // No se devuelve en JSON
	Role      Role           `gorm:"size:20;not null" json:"role"`
	LastLogin *time.Time     `json:"last_login,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// SetPassword cifra y establece la contraseña del usuario
func (u *User) SetPassword(plainPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifica si la contraseña proporcionada coincide con la almacenada
func (u *User) CheckPassword(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))
	return err == nil
}

// CanAccess determina si el usuario tiene acceso según su rol
func (u *User) CanAccess(minRole Role) bool {
	// Mapa de jerarquía de roles (menor valor = mayor privilegio)
	roleHierarchy := map[Role]int{
		RoleAdmin:  1,
		RoleUser:   2,
		RoleViewer: 3,
	}
	
	userLevel, userExists := roleHierarchy[u.Role]
	minLevel, minExists := roleHierarchy[minRole]
	
	if !userExists || !minExists {
		return false
	}
	
	return userLevel <= minLevel
} 