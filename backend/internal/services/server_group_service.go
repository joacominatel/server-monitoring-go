package services

import (
	"fmt"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
	"gorm.io/gorm"
)

// ServerGroupService servicio para gestionar grupos de servidores
type ServerGroupService struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewServerGroupService crea un nuevo servicio de grupos de servidores
func NewServerGroupService(db *gorm.DB, log logger.Logger) *ServerGroupService {
	return &ServerGroupService{
		db:     db,
		logger: log,
	}
}

// CreateGroup crea un nuevo grupo de servidores
func (sgs *ServerGroupService) CreateGroup(group *models.ServerGroup) error {
	// Verificar recursividad en grupos si se especifica un padre
	if group.ParentID != nil {
		if err := sgs.checkGroupRecursion(*group.ParentID, 0); err != nil {
			return err
		}
	}

	if err := sgs.db.Create(group).Error; err != nil {
		sgs.logger.Errorf("Error al crear grupo de servidores: %v", err)
		return err
	}

	sgs.logger.Infof("Grupo de servidores creado: %s (ID: %d)", group.Name, group.ID)
	return nil
}

// UpdateGroup actualiza un grupo existente
func (sgs *ServerGroupService) UpdateGroup(group *models.ServerGroup) error {
	// Verificar que el grupo existe
	var existingGroup models.ServerGroup
	if err := sgs.db.First(&existingGroup, group.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("grupo no encontrado")
		}
		return err
	}

	// Verificar recursividad en grupos si se cambia el padre
	if group.ParentID != nil && (existingGroup.ParentID == nil || *existingGroup.ParentID != *group.ParentID) {
		if err := sgs.checkGroupRecursion(*group.ParentID, group.ID); err != nil {
			return err
		}
	}

	if err := sgs.db.Save(group).Error; err != nil {
		sgs.logger.Errorf("Error al actualizar grupo de servidores: %v", err)
		return err
	}

	sgs.logger.Infof("Grupo de servidores actualizado: %s (ID: %d)", group.Name, group.ID)
	return nil
}

// DeleteGroup elimina un grupo de servidores
func (sgs *ServerGroupService) DeleteGroup(id uint) error {
	// Verificar si existen grupos hijos
	var childCount int64
	if err := sgs.db.Model(&models.ServerGroup{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		sgs.logger.Errorf("Error al verificar grupos hijos: %v", err)
		return err
	}

	if childCount > 0 {
		return fmt.Errorf("no se puede eliminar un grupo con subgrupos (%d subgrupos encontrados)", childCount)
	}

	// Iniciar transacción
	tx := sgs.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Eliminar relaciones con servidores
	if err := tx.Exec("DELETE FROM server_group_servers WHERE server_group_id = ?", id).Error; err != nil {
		tx.Rollback()
		sgs.logger.Errorf("Error al eliminar relaciones con servidores: %v", err)
		return err
	}

	// Eliminar el grupo
	if err := tx.Delete(&models.ServerGroup{}, id).Error; err != nil {
		tx.Rollback()
		sgs.logger.Errorf("Error al eliminar grupo: %v", err)
		return err
	}

	// Confirmar transacción
	if err := tx.Commit().Error; err != nil {
		sgs.logger.Errorf("Error al confirmar eliminación de grupo: %v", err)
		return err
	}

	sgs.logger.Infof("Grupo de servidores eliminado: %d", id)
	return nil
}

// GetGroup obtiene un grupo por ID con sus relaciones
func (sgs *ServerGroupService) GetGroup(id uint, includeChildren bool, includeServers bool) (*models.ServerGroup, error) {
	var group models.ServerGroup

	// Construir query con preload condicional
	query := sgs.db.Model(&models.ServerGroup{})

	if includeChildren {
		query = query.Preload("Children")
	}

	if includeServers {
		query = query.Preload("Servers")
	}

	if err := query.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("grupo no encontrado")
		}
		sgs.logger.Errorf("Error al obtener grupo de servidores: %v", err)
		return nil, err
	}

	return &group, nil
}

// GetAllGroups obtiene todos los grupos
func (sgs *ServerGroupService) GetAllGroups(includeChildren bool, includeServers bool) ([]models.ServerGroup, error) {
	var groups []models.ServerGroup

	// Construir query con preload condicional
	query := sgs.db.Model(&models.ServerGroup{})

	if includeChildren {
		query = query.Preload("Children")
	}

	if includeServers {
		query = query.Preload("Servers")
	}

	if err := query.Find(&groups).Error; err != nil {
		sgs.logger.Errorf("Error al obtener grupos de servidores: %v", err)
		return nil, err
	}

	return groups, nil
}

// GetRootGroups obtiene los grupos raíz (sin padre)
func (sgs *ServerGroupService) GetRootGroups(includeChildren bool, includeServers bool) ([]models.ServerGroup, error) {
	var groups []models.ServerGroup

	// Construir query con preload condicional
	query := sgs.db.Model(&models.ServerGroup{}).Where("parent_id IS NULL")

	if includeChildren {
		query = query.Preload("Children")

		// Si queremos hijos recursivamente, tenemos que hacer preload de Children.Children
		if includeChildren {
			query = query.Preload("Children.Children")
		}
	}

	if includeServers {
		query = query.Preload("Servers")

		// Si queremos servidores en los hijos
		if includeChildren {
			query = query.Preload("Children.Servers")
		}
	}

	if err := query.Find(&groups).Error; err != nil {
		sgs.logger.Errorf("Error al obtener grupos raíz: %v", err)
		return nil, err
	}

	return groups, nil
}

// AddServerToGroup añade un servidor a un grupo
func (sgs *ServerGroupService) AddServerToGroup(groupID, serverID uint) error {
	// Comprobar que el grupo existe
	if err := sgs.db.First(&models.ServerGroup{}, groupID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("grupo no encontrado")
		}
		return err
	}

	// Comprobar que el servidor existe
	if err := sgs.db.First(&models.Server{}, serverID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("servidor no encontrado")
		}
		return err
	}

	// Añadir relación
	if err := sgs.db.Exec("INSERT INTO server_group_servers (server_group_id, server_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		groupID, serverID).Error; err != nil {
		sgs.logger.Errorf("Error al añadir servidor a grupo: %v", err)
		return err
	}

	sgs.logger.Infof("Servidor %d añadido al grupo %d", serverID, groupID)
	return nil
}

// RemoveServerFromGroup elimina un servidor de un grupo
func (sgs *ServerGroupService) RemoveServerFromGroup(groupID, serverID uint) error {
	if err := sgs.db.Exec("DELETE FROM server_group_servers WHERE server_group_id = ? AND server_id = ?",
		groupID, serverID).Error; err != nil {
		sgs.logger.Errorf("Error al eliminar servidor del grupo: %v", err)
		return err
	}

	sgs.logger.Infof("Servidor %d eliminado del grupo %d", serverID, groupID)
	return nil
}

// GetGroupsByServer obtiene los grupos a los que pertenece un servidor
func (sgs *ServerGroupService) GetGroupsByServer(serverID uint) ([]models.ServerGroup, error) {
	var groups []models.ServerGroup

	// Verificar que el servidor existe
	if err := sgs.db.First(&models.Server{}, serverID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("servidor no encontrado")
		}
		return nil, err
	}

	// Obtener grupos a los que pertenece
	if err := sgs.db.Joins("JOIN server_group_servers ON server_group_servers.server_group_id = server_groups.id").
		Where("server_group_servers.server_id = ?", serverID).
		Find(&groups).Error; err != nil {
		sgs.logger.Errorf("Error al obtener grupos del servidor: %v", err)
		return nil, err
	}

	return groups, nil
}

// checkGroupRecursion verifica que no haya recursión en la jerarquía de grupos
func (sgs *ServerGroupService) checkGroupRecursion(parentID, groupID uint) error {
	if parentID == groupID {
		return fmt.Errorf("un grupo no puede ser su propio padre")
	}

	var parent models.ServerGroup
	if err := sgs.db.First(&parent, parentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("grupo padre no encontrado")
		}
		return err
	}

	// Si el padre tiene padre, verificar recursivamente
	if parent.ParentID != nil {
		return sgs.checkGroupRecursion(*parent.ParentID, groupID)
	}

	return nil
}
