package repository

import (
	"auth-service/internal/models"
	"time"

	"gorm.io/gorm"
)

type RolModuloPermisoRepository struct {
	db *gorm.DB
}

func NewRolModuloPermisoRepository(db *gorm.DB) *RolModuloPermisoRepository {
	return &RolModuloPermisoRepository{db: db}
}

func (r *RolModuloPermisoRepository) Create(rolModuloPermiso *models.RolModuloPermiso) error {
	return r.db.Create(rolModuloPermiso).Error
}

func (r *RolModuloPermisoRepository) GetAll() ([]models.RolModuloPermiso, error) {
	var rolModuloPermisos []models.RolModuloPermiso
	err := r.db.Where("fecha_eliminacion IS NULL").
		Preload("Role").
		Preload("Modulo", "fecha_eliminacion IS NULL").
		Preload("PermisoTipo").
		Find(&rolModuloPermisos).Error
	return rolModuloPermisos, err
}

func (r *RolModuloPermisoRepository) GetByID(id int) (*models.RolModuloPermiso, error) {
	var rolModuloPermiso models.RolModuloPermiso
	err := r.db.Where("fecha_eliminacion IS NULL").
		Preload("Role").
		Preload("Modulo", "fecha_eliminacion IS NULL").
		Preload("PermisoTipo").
		First(&rolModuloPermiso, id).Error
	if err != nil {
		return nil, err
	}
	return &rolModuloPermiso, nil
}

func (r *RolModuloPermisoRepository) GetByRoleID(roleID int) ([]models.RolModuloPermiso, error) {
	var rolModuloPermisos []models.RolModuloPermiso
	err := r.db.Where("id_rol = ? AND fecha_eliminacion IS NULL", roleID).
		Preload("Role").
		Preload("Modulo", "fecha_eliminacion IS NULL").
		Preload("PermisoTipo").
		Find(&rolModuloPermisos).Error
	return rolModuloPermisos, err
}

func (r *RolModuloPermisoRepository) Delete(id int) error {
	now := time.Now()
	return r.db.Model(&models.RolModuloPermiso{}).
		Where("id = ? AND fecha_eliminacion IS NULL", id).
		Update("fecha_eliminacion", now).Error
}

func (r *RolModuloPermisoRepository) DeleteByRoleID(roleID int) error {
	now := time.Now()
	return r.db.Model(&models.RolModuloPermiso{}).
		Where("id_rol = ? AND fecha_eliminacion IS NULL", roleID).
		Update("fecha_eliminacion", now).Error
}
