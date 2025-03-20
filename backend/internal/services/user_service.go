package services

import (
	"errors"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
)

// UserService maneja la lógica de negocio relacionada con usuarios
type UserService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewUserService crea una nueva instancia del servicio de usuarios
func NewUserService(db *gorm.DB, logger logger.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// GetAllUsers obtiene todos los usuarios
func (s *UserService) GetAllUsers() ([]models.User, error) {
	var users []models.User
	
	if err := s.db.Find(&users).Error; err != nil {
		s.logger.Errorf("Error al obtener todos los usuarios: %v", err)
		return nil, err
	}
	
	return users, nil
}

// GetUserByID obtiene un usuario por su ID
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Usuario con ID %d no encontrado", id)
			return nil, errors.New("usuario no encontrado")
		}
		s.logger.Errorf("Error al obtener usuario por ID: %v", err)
		return nil, err
	}
	
	return &user, nil
}

// GetUserByUsername obtiene un usuario por su nombre de usuario
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Usuario con username '%s' no encontrado", username)
			return nil, errors.New("usuario no encontrado")
		}
		s.logger.Errorf("Error al obtener usuario por username: %v", err)
		return nil, err
	}
	
	return &user, nil
}

// GetUserByEmail obtiene un usuario por su email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Usuario con email '%s' no encontrado", email)
			return nil, errors.New("usuario no encontrado")
		}
		s.logger.Errorf("Error al obtener usuario por email: %v", err)
		return nil, err
	}
	
	return &user, nil
}

// CreateUser crea un nuevo usuario
func (s *UserService) CreateUser(user *models.User, plainPassword string) error {
	// Verificar si ya existe un usuario con el mismo username o email
	var count int64
	s.db.Model(&models.User{}).Where("username = ? OR email = ?", user.Username, user.Email).Count(&count)
	if count > 0 {
		s.logger.Warnf("Intento de crear usuario con username o email duplicado: %s, %s", user.Username, user.Email)
		return errors.New("el nombre de usuario o email ya está en uso")
	}
	
	// Establecer contraseña
	if err := user.SetPassword(plainPassword); err != nil {
		s.logger.Errorf("Error al cifrar contraseña: %v", err)
		return err
	}
	
	// Guardar usuario
	if err := s.db.Create(user).Error; err != nil {
		s.logger.Errorf("Error al crear usuario: %v", err)
		return err
	}
	
	s.logger.Infof("Usuario creado exitosamente: ID=%d, Username=%s", user.ID, user.Username)
	return nil
}

// UpdateUser actualiza un usuario existente
func (s *UserService) UpdateUser(user *models.User) error {
	// La contraseña se maneja en un método separado
	if err := s.db.Omit("password").Save(user).Error; err != nil {
		s.logger.Errorf("Error al actualizar usuario: %v", err)
		return err
	}
	
	s.logger.Infof("Usuario actualizado exitosamente: ID=%d, Username=%s", user.ID, user.Username)
	return nil
}

// DeleteUser elimina un usuario (soft delete)
func (s *UserService) DeleteUser(id uint) error {
	if err := s.db.Delete(&models.User{}, id).Error; err != nil {
		s.logger.Errorf("Error al eliminar usuario: %v", err)
		return err
	}
	
	s.logger.Infof("Usuario eliminado exitosamente: ID=%d", id)
	return nil
}

// ChangePassword cambia la contraseña de un usuario
func (s *UserService) ChangePassword(id uint, newPassword string) error {
	user, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	
	if err := user.SetPassword(newPassword); err != nil {
		s.logger.Errorf("Error al cifrar nueva contraseña: %v", err)
		return err
	}
	
	if err := s.db.Save(user).Error; err != nil {
		s.logger.Errorf("Error al guardar nueva contraseña: %v", err)
		return err
	}
	
	s.logger.Infof("Contraseña cambiada exitosamente para usuario ID=%d", id)
	return nil
} 