package repository

import (
	"auth-service/internal/models"
	"fmt"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *models.Role) error {
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

func (r *RoleRepository) AssignModulePermission(roleID, moduleID, permisoTipoID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar que el rol existe
		var role models.Role
		if err := tx.First(&role, roleID).Error; err != nil {
			return fmt.Errorf("rol no encontrado: %v", err)
		}

		// Verificar que el módulo existe
		var module models.Module
		if err := tx.First(&module, moduleID).Error; err != nil {
			return fmt.Errorf("módulo no encontrado: %v", err)
		}

		// Verificar que el tipo de permiso existe
		var permisoTipo models.PermisoTipo
		if err := tx.First(&permisoTipo, permisoTipoID).Error; err != nil {
			return fmt.Errorf("tipo de permiso no encontrado: %v", err)
		}

		// Verificar que el permiso está disponible para el módulo
		var moduloPermiso models.ModuloPermiso
		if err := tx.Where("id_modulo = ? AND id_permiso_tipo = ?", moduleID, permisoTipoID).
			First(&moduloPermiso).Error; err != nil {
			return fmt.Errorf("el permiso no está disponible para este módulo")
		}

		// Verificar si ya existe la asignación
		var exists bool
		err := tx.Model(&models.RolModuloPermiso{}).
			Where("id_rol = ? AND id_modulo = ? AND id_permiso_tipo = ?", roleID, moduleID, permisoTipoID).
			Select("count(*) > 0").
			Scan(&exists).Error

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("esta asignación de permiso ya existe")
		}

		// Crear nueva asignación
		rolModuloPermiso := &models.RolModuloPermiso{
			IdRol:         roleID,
			IdModulo:      moduleID,
			IdPermisoTipo: permisoTipoID,
		}

		return tx.Create(rolModuloPermiso).Error
	})
}

func (r *RoleRepository) GetRolePermissions(roleID int) ([]models.RolModuloPermiso, error) {
    var permissions []models.RolModuloPermiso
    
    // Primero verifica si el rol existe
    var role models.Role
    if err := r.db.First(&role, roleID).Error; err != nil {
        return nil, fmt.Errorf("rol no encontrado: %v", err)
    }
    
    // Obtiene los permisos con todos los preloads necesarios
    err := r.db.Where("id_rol = ?", roleID).
        Preload("Role", func(db *gorm.DB) *gorm.DB {
            return db.Select("id, nombre, fecha_creacion, fecha_actualizacion")
        }).
        Preload("Modulo").
        Preload("PermisoTipo").
        Find(&permissions).Error

    if err != nil {
        return nil, err
    }

    // Asegura que cada permiso tenga la información correcta del rol
    for i := range permissions {
        permissions[i].Role = role
    }

    return permissions, nil
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