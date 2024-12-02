package repository

import (
	"auth-service/internal/models"

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
	err := r.db.Preload("Role").Preload("Modulo").Preload("PermisoTipo").Find(&rolModuloPermisos).Error
	return rolModuloPermisos, err
}

func (r *RolModuloPermisoRepository) GetByID(id int) (*models.RolModuloPermiso, error) {
	var rolModuloPermiso models.RolModuloPermiso
	err := r.db.Preload("Role").Preload("Modulo").Preload("PermisoTipo").First(&rolModuloPermiso, id).Error
	if err != nil {
		return nil, err
	}
	return &rolModuloPermiso, nil
}

func (r *RolModuloPermisoRepository) GetByRoleID(roleID int) ([]models.RolModuloPermiso, error) {
	var rolModuloPermisos []models.RolModuloPermiso
	err := r.db.Where("id_rol = ?", roleID).
		Preload("Role").
		Preload("Modulo").
		Preload("PermisoTipo").
		Find(&rolModuloPermisos).Error
	return rolModuloPermisos, err
}

func (r *RolModuloPermisoRepository) Delete(id int) error {
	return r.db.Delete(&models.RolModuloPermiso{}, id).Error
}

func (r *RolModuloPermisoRepository) DeleteByRoleID(roleID int) error {
	return r.db.Where("id_rol = ?", roleID).Delete(&models.RolModuloPermiso{}).Error
}
