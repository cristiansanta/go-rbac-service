package repository

import (
	"auth-service/internal/constants"
	"auth-service/internal/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *models.Role) error {
	var exists bool
	if err := r.db.Model(&models.Role{}).
		Where("LOWER(nombre) = LOWER(?)", role.Nombre).
		Select("count(*) > 0").
		Scan(&exists).Error; err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("ya existe un rol con este nombre")
	}

	return r.db.Create(role).Error
}

func (r *RoleRepository) GetAll() ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) GetByID(id int) (*models.Role, error) {
	var role models.Role
	err := r.db.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Primero eliminamos todos los permisos asociados al rol
		if err := tx.Where("id_rol = ?", id).Delete(&models.RolModuloPermiso{}).Error; err != nil {
			return err
		}
		// Luego eliminamos el rol
		return tx.Delete(&models.Role{}, id).Error
	})
}

func (r *RoleRepository) AssignModulePermission(roleID, moduleID int, permisoTipoIDs []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar rol
		var role models.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return fmt.Errorf("rol no encontrado: %v", err)
		}

		// Verificar si es SuperAdmin
		if strings.ToUpper(role.Nombre) == constants.RoleSuperAdmin {
			return fmt.Errorf("no se pueden modificar los permisos del rol SuperAdmin")
		}

		// Verificar módulo
		var module models.Module
		if err := tx.First(&module, moduleID).Error; err != nil {
			return fmt.Errorf("módulo no encontrado: %v", err)
		}

		// Verificar permisos
		var count int64
		if err := tx.Model(&models.PermisoTipo{}).Where("id IN ?", permisoTipoIDs).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(permisoTipoIDs) {
			return fmt.Errorf("algunos permisos no existen")
		}

		// Eliminar permisos existentes
		if err := tx.Where("id_rol = ? AND id_modulo = ?", roleID, moduleID).
			Delete(&models.RolModuloPermiso{}).Error; err != nil {
			return err
		}

		// Crear nuevos permisos
		for _, permisoID := range permisoTipoIDs {
			rolModuloPermiso := &models.RolModuloPermiso{
				IdRol:         roleID,
				IdModulo:      moduleID,
				IdPermisoTipo: permisoID,
			}
			if err := tx.Create(rolModuloPermiso).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *RoleRepository) GetRolePermissions(roleID int) ([]models.RolModuloPermisoResponse, error) {
	var role models.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return nil, fmt.Errorf("rol no encontrado: %v", err)
	}

	var permissions []models.RolModuloPermiso
	err := r.db.Where("id_rol = ? AND fecha_eliminacion IS NULL", roleID).
		Preload("Role", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, nombre, descripcion, fecha_creacion, fecha_actualizacion")
		}).
		Preload("Modulo").
		Preload("PermisoTipo").
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	response := make([]models.RolModuloPermisoResponse, len(permissions))
	for i, p := range permissions {
		response[i] = models.RolModuloPermisoResponse{
			ID:            p.ID,
			IdRol:         p.IdRol,
			IdModulo:      p.IdModulo,
			IdPermisoTipo: p.IdPermisoTipo,
			FechaCreacion: p.FechaCreacion,
			Role:          p.Role,
			Modulo: models.ModuleResponse{
				ID:                 p.Modulo.ID,
				Nombre:             p.Modulo.Nombre,
				Descripcion:        p.Modulo.Descripcion,
				FechaCreacion:      p.Modulo.FechaCreacion,
				FechaActualizacion: p.Modulo.FechaActualizacion,
			},
			PermisoTipo: p.PermisoTipo,
		}
	}

	return response, nil
}

func (r *RoleRepository) RemoveModulePermission(roleID, moduleID, permisoTipoID int) error {
	result := r.db.Where(
		"id_rol = ? AND id_modulo = ? AND id_permiso_tipo = ?",
		roleID, moduleID, permisoTipoID,
	).Delete(&models.RolModuloPermiso{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no se encontró el permiso especificado")
	}

	return nil
}

func (r *RoleRepository) RemoveModuleFromRole(roleID, moduleID int) error {
	result := r.db.Where(
		"id_rol = ? AND id_modulo = ?",
		roleID, moduleID,
	).Delete(&models.RolModuloPermiso{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no se encontraron permisos para el módulo especificado")
	}

	return nil
}

func (r *RoleRepository) GetUsersByRoleID(roleID int) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id_rol = ?", roleID).Find(&users).Error
	return users, err
}
