package repository

import (
	"auth-service/internal/models"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type PermisoTipoRepository struct {
	db *gorm.DB
}

func NewPermisoTipoRepository(db *gorm.DB) *PermisoTipoRepository {
	return &PermisoTipoRepository{db: db}
}

func (r *PermisoTipoRepository) Create(permisoTipo *models.PermisoTipo) error {
	validCodigos := []string{
		models.PermisoVer,
		models.PermisoCreateEdit,
		models.PermisoExportar,
		models.PermisoEliminar,
	}

	codigoUpper := strings.ToUpper(permisoTipo.Codigo)
	isValid := false
	for _, codigo := range validCodigos {
		if codigoUpper == codigo {
			permisoTipo.Codigo = codigo
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("código de permiso inválido")
	}

	return r.db.Create(permisoTipo).Error
}

func (r *PermisoTipoRepository) GetAll() ([]models.PermisoTipo, error) {
	var permisoTipos []models.PermisoTipo
	err := r.db.Find(&permisoTipos).Error
	return permisoTipos, err
}

func (r *PermisoTipoRepository) GetByID(id int) (*models.PermisoTipo, error) {
	var permisoTipo models.PermisoTipo
	err := r.db.First(&permisoTipo, id).Error
	if err != nil {
		return nil, err
	}
	return &permisoTipo, nil
}

func (r *PermisoTipoRepository) GetByCodigo(codigo string) (*models.PermisoTipo, error) {
	var permisoTipo models.PermisoTipo
	err := r.db.Where("codigo = ?", codigo).First(&permisoTipo).Error
	if err != nil {
		return nil, err
	}
	return &permisoTipo, nil
}

func (r *PermisoTipoRepository) Update(permisoTipo *models.PermisoTipo) error {
	return r.db.Save(permisoTipo).Error
}

func (r *PermisoTipoRepository) Delete(id int) error {
	return r.db.Delete(&models.PermisoTipo{}, id).Error
}
