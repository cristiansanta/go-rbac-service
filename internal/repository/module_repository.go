package repository

import (
	"auth-service/internal/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ModuleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

// repository/module_repository.go
func (r *ModuleRepository) Create(module *models.Module) error {
	var exists bool
	if err := r.db.Model(&models.Module{}).
		Where("LOWER(nombre) = LOWER(?)", module.Nombre).
		Select("count(*) > 0").
		Scan(&exists).Error; err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("ya existe un módulo con este nombre")
	}

	return r.db.Create(module).Error
}

func (r *ModuleRepository) GetAll() ([]models.Module, error) {
	var modules []models.Module
	err := r.db.Where("fecha_eliminacion IS NULL").Find(&modules).Error
	return modules, err
}

func (r *ModuleRepository) GetByID(id int) (*models.Module, error) {
	var module models.Module
	err := r.db.Where("fecha_eliminacion IS NULL").First(&module, id).Error
	if err != nil {
		return nil, err
	}
	return &module, nil
}

func (r *ModuleRepository) Update(module *models.Module) error {
	return r.db.Save(module).Error
}

func (r *ModuleRepository) Delete(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Primero, verificar si el módulo existe y no está ya eliminado
		var module models.Module
		if err := tx.Where("id = ? AND fecha_eliminacion IS NULL", id).First(&module).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("módulo no encontrado o ya está eliminado")
			}
			return err
		}

		// Obtener la fecha actual
		now := time.Now()

		// Actualizar fecha_eliminacion del módulo
		if err := tx.Model(&module).Update("fecha_eliminacion", now).Error; err != nil {
			return err
		}

		// Marcar como eliminados los registros relacionados en rol_modulo_permisos
		if err := tx.Model(&models.RolModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", now).Error; err != nil {
			return err
		}

		// Marcar como eliminados los registros relacionados en modulo_permisos
		if err := tx.Model(&models.ModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", now).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *ModuleRepository) AssignPermissions(moduleID int, permisoTipoIDs []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar que el módulo existe
		var module models.Module
		if err := tx.First(&module, moduleID).Error; err != nil {
			return fmt.Errorf("módulo no encontrado: %v", err)
		}

		// Verificar que todos los tipos de permisos existen
		var count int64
		if err := tx.Model(&models.PermisoTipo{}).Where("id IN ?", permisoTipoIDs).Count(&count).Error; err != nil {
			return err
		}
		if int(count) != len(permisoTipoIDs) {
			return fmt.Errorf("algunos tipos de permisos no existen")
		}

		// Eliminar permisos existentes
		if err := tx.Where("id_modulo = ?", moduleID).Delete(&models.ModuloPermiso{}).Error; err != nil {
			return err
		}

		// Insertar nuevos permisos
		for _, permisoTipoID := range permisoTipoIDs {
			moduloPermiso := &models.ModuloPermiso{
				IdModulo:      moduleID,
				IdPermisoTipo: permisoTipoID,
			}
			if err := tx.Create(moduloPermiso).Error; err != nil {
				return err
			}
		}

		// Eliminar asignaciones de roles que ya no son válidas
		var validPermisos []models.RolModuloPermiso
		err := tx.Where("id_modulo = ? AND id_permiso_tipo NOT IN ?", moduleID, permisoTipoIDs).
			Delete(&models.RolModuloPermiso{}).Error
		if err != nil {
			return err
		}

		return tx.Where("id_modulo = ?", moduleID).
			Preload("Modulo").
			Preload("PermisoTipo").
			Find(&validPermisos).Error
	})
}

func (r *ModuleRepository) GetModuleWithPermissions(moduleID int) (*models.ModuleWithPermissions, error) {
	var module models.Module
	err := r.db.Where("fecha_eliminacion IS NULL").
		Preload("Permisos").
		First(&module, moduleID).Error
	if err != nil {
		return nil, err
	}

	return &models.ModuleWithPermissions{
		ID:                 module.ID,
		Nombre:             module.Nombre,
		Descripcion:        module.Descripcion,
		Permisos:           module.Permisos,
		FechaCreacion:      module.FechaCreacion,
		FechaActualizacion: module.FechaActualizacion,
		FechaEliminacion:   module.FechaEliminacion,
	}, nil
}

func (r *ModuleRepository) RemovePermission(moduleID int, permisoTipoID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Primero eliminamos las asignaciones de roles que usan este permiso
		if err := tx.Where(
			"id_modulo = ? AND id_permiso_tipo = ?",
			moduleID, permisoTipoID,
		).Delete(&models.RolModuloPermiso{}).Error; err != nil {
			return err
		}

		// Luego eliminamos el permiso del módulo
		result := tx.Where(
			"id_modulo = ? AND id_permiso_tipo = ?",
			moduleID, permisoTipoID,
		).Delete(&models.ModuloPermiso{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("no se encontró el permiso especificado para este módulo")
		}

		return nil
	})
}
func (r *ModuleRepository) Restore(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar si el módulo existe y está eliminado
		var module models.Module
		if err := tx.Unscoped().First(&module, id).Error; err != nil {
			return fmt.Errorf("módulo no encontrado")
		}

		if module.FechaEliminacion == nil {
			return fmt.Errorf("el módulo no está eliminado")
		}

		// Restaurar el módulo
		if err := tx.Model(&module).Update("fecha_eliminacion", nil).Error; err != nil {
			return err
		}

		// Restaurar registros relacionados en rol_modulo_permisos
		if err := tx.Model(&models.RolModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", nil).Error; err != nil {
			return err
		}

		// Restaurar registros relacionados en modulo_permisos
		if err := tx.Model(&models.ModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", nil).Error; err != nil {
			return err
		}

		return nil
	})
}
func (r *ModuleRepository) GetDeletedModules() ([]models.Module, error) {
	var modules []models.Module
	err := r.db.Where("fecha_eliminacion IS NOT NULL").
		Order("fecha_eliminacion DESC").
		Find(&modules).Error
	return modules, err
}
