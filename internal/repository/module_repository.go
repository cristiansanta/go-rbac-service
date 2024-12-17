// internal/repository/module_repository.go
package repository

import (
	"auth-service/internal/models"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ModuleRepository struct {
	db *gorm.DB
}

func NewModuleRepository(db *gorm.DB) *ModuleRepository {
	return &ModuleRepository{db: db}
}

// Método para crear múltiples módulos
func (r *ModuleRepository) CrearModulos(modulosInfo []models.InfoModulo) ([]models.Module, error) {
	var modulosCreados []models.Module

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Obtener todos los permisos base
		var permisos []models.PermisoTipo
		if err := tx.Find(&permisos).Error; err != nil {
			return fmt.Errorf("error obteniendo permisos base: %v", err)
		}

		if len(permisos) != 4 {
			return fmt.Errorf("no se encontraron los 4 permisos base necesarios")
		}

		// Verificar nombres duplicados (case insensitive)
		for i, modulo := range modulosInfo {
			// Verificar contra módulos existentes en la base de datos
			var exists bool
			if err := tx.Model(&models.Module{}).
				Where("LOWER(nombre) = LOWER(?)", modulo.Nombre).
				Where("fecha_eliminacion IS NULL").
				Select("count(*) > 0").
				Scan(&exists).Error; err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("ya existe un módulo con el nombre: %s", modulo.Nombre)
			}

			// Verificar contra otros módulos en el mismo request
			for j := 0; j < i; j++ {
				if strings.EqualFold(modulosInfo[j].Nombre, modulo.Nombre) {
					return fmt.Errorf("módulos duplicados en la petición: %s", modulo.Nombre)
				}
			}
		}

		// Crear los módulos y asignar permisos
		for _, moduloInfo := range modulosInfo {
			modulo := models.Module{
				Nombre:      moduloInfo.Nombre,
				Descripcion: moduloInfo.Descripcion,
			}

			if err := tx.Create(&modulo).Error; err != nil {
				return fmt.Errorf("error creando módulo %s: %v", moduloInfo.Nombre, err)
			}

			// Asignar los permisos base
			for _, permiso := range permisos {
				moduloPermiso := models.ModuloPermiso{
					IdModulo:      modulo.ID,
					IdPermisoTipo: permiso.ID,
				}
				if err := tx.Create(&moduloPermiso).Error; err != nil {
					return fmt.Errorf("error asignando permiso al módulo %s: %v", modulo.Nombre, err)
				}
			}

			// Cargar el módulo con sus permisos para la respuesta
			if err := tx.Preload("Permisos").First(&modulo, modulo.ID).Error; err != nil {
				return fmt.Errorf("error cargando módulo con permisos: %v", err)
			}

			modulosCreados = append(modulosCreados, modulo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return modulosCreados, nil
}

func (r *ModuleRepository) GetAll() ([]models.Module, error) {
	var modules []models.Module
	// Modificamos para incluir los permisos en la consulta
	err := r.db.Where("fecha_eliminacion IS NULL").
		Preload("Permisos"). // Agregamos esto para cargar los permisos
		Find(&modules).Error
	return modules, err
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

func (r *ModuleRepository) Delete(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// Verificar si el módulo existe y no está ya eliminado
		var module models.Module
		if err := tx.Where("id = ? AND fecha_eliminacion IS NULL", id).First(&module).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("módulo no encontrado o ya está eliminado")
			}
			return err
		}

		// Actualizar fecha_eliminacion del módulo
		if err := tx.Model(&module).Update("fecha_eliminacion", now).Error; err != nil {
			return err
		}

		// Marcar como eliminados los registros relacionados
		if err := tx.Model(&models.RolModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", now).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.ModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", now).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *ModuleRepository) Restore(id int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

		// Restaurar registros relacionados
		if err := tx.Model(&models.RolModuloPermiso{}).
			Where("id_modulo = ?", id).
			Update("fecha_eliminacion", nil).Error; err != nil {
			return err
		}

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

// AssignPermissions asigna permisos específicos a un módulo
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

		return nil
	})
}

// RemovePermission elimina un permiso específico de un módulo
func (r *ModuleRepository) RemovePermission(moduleID int, permisoTipoID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Verificar que el módulo existe
		if err := tx.First(&models.Module{}, moduleID).Error; err != nil {
			return fmt.Errorf("módulo no encontrado: %v", err)
		}

		// Verificar que el permiso existe
		if err := tx.First(&models.PermisoTipo{}, permisoTipoID).Error; err != nil {
			return fmt.Errorf("permiso no encontrado: %v", err)
		}

		// Eliminar la relación
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
