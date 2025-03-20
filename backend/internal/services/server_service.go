package services

import (
	"errors"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
)

// ServerService maneja la lógica de negocio relacionada con servidores
type ServerService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewServerService crea una nueva instancia del servicio de servidores
func NewServerService(db *gorm.DB, logger logger.Logger) *ServerService {
	return &ServerService{
		db:     db,
		logger: logger,
	}
}

// GetAllServers obtiene todos los servidores activos
func (s *ServerService) GetAllServers() ([]models.Server, error) {
	var servers []models.Server
	
	if err := s.db.Where("is_active = ?", true).Find(&servers).Error; err != nil {
		s.logger.Errorf("Error al obtener todos los servidores: %v", err)
		return nil, err
	}
	
	return servers, nil
}

// GetServerByID obtiene un servidor por su ID
func (s *ServerService) GetServerByID(id uint) (*models.Server, error) {
	var server models.Server
	
	if err := s.db.First(&server, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Servidor con ID %d no encontrado", id)
			return nil, errors.New("servidor no encontrado")
		}
		s.logger.Errorf("Error al obtener servidor con ID %d: %v", id, err)
		return nil, err
	}
	
	return &server, nil
}

// GetServerByHostname obtiene un servidor por su nombre de host
func (s *ServerService) GetServerByHostname(hostname string) (*models.Server, error) {
	var server models.Server
	
	if err := s.db.Where("hostname = ?", hostname).First(&server).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warnf("Servidor con hostname %s no encontrado", hostname)
			return nil, errors.New("servidor no encontrado")
		}
		s.logger.Errorf("Error al obtener servidor con hostname %s: %v", hostname, err)
		return nil, err
	}
	
	return &server, nil
}

// CreateServer crea un nuevo servidor
func (s *ServerService) CreateServer(server *models.Server) error {
	if err := s.db.Create(server).Error; err != nil {
		s.logger.Errorf("Error al crear servidor: %v", err)
		return err
	}
	
	s.logger.Infof("Servidor creado exitosamente: ID=%d, Hostname=%s", server.ID, server.Hostname)
	return nil
}

// UpdateServer actualiza un servidor existente
func (s *ServerService) UpdateServer(server *models.Server) error {
	if err := s.db.Save(server).Error; err != nil {
		s.logger.Errorf("Error al actualizar servidor con ID %d: %v", server.ID, err)
		return err
	}
	
	s.logger.Infof("Servidor actualizado exitosamente: ID=%d, Hostname=%s", server.ID, server.Hostname)
	return nil
}

// DeleteServer elimina un servidor (soft delete)
func (s *ServerService) DeleteServer(id uint) error {
	if err := s.db.Delete(&models.Server{}, id).Error; err != nil {
		s.logger.Errorf("Error al eliminar servidor con ID %d: %v", id, err)
		return err
	}
	
	s.logger.Infof("Servidor eliminado exitosamente: ID=%d", id)
	return nil
} 