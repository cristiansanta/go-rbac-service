package repository

import (
	"auth-service/internal/models"

	"gorm.io/gorm"
)

type RegistroAuditoriaRepository struct {
	db *gorm.DB
}

func (r *RegistroAuditoriaRepository) GetByRol(rol string, page, size int) ([]models.RegistroAuditoria, int64, error) {
	var registros []models.RegistroAuditoria
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.RegistroAuditoria{}).
		Where("rol = ?", rol).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("rol = ?", rol).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&registros).Error; err != nil {
		return nil, 0, err
	}

	return registros, total, nil
}

func NewRegistroAuditoriaRepository(db *gorm.DB) *RegistroAuditoriaRepository {
	return &RegistroAuditoriaRepository{db: db}
}

// Create registra un nuevo registro de auditoría
func (r *RegistroAuditoriaRepository) Create(registro *models.RegistroAuditoria) error {
	return r.db.Create(registro).Error
}

// GetAll obtiene todos los registros con paginación
func (r *RegistroAuditoriaRepository) GetAll(page, size int) ([]models.RegistroAuditoria, int64, error) {
	var registros []models.RegistroAuditoria
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.RegistroAuditoria{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&registros).Error; err != nil {
		return nil, 0, err
	}

	return registros, total, nil
}

// GetByIdUsuario obtiene los registros de un usuario específico
func (r *RegistroAuditoriaRepository) GetByIdUsuario(idUsuario int, page, size int) ([]models.RegistroAuditoria, int64, error) {
	var registros []models.RegistroAuditoria
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.RegistroAuditoria{}).
		Where("id_usuario = ?", idUsuario).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("id_usuario = ?", idUsuario).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&registros).Error; err != nil {
		return nil, 0, err
	}

	return registros, total, nil
}

// GetByNombreModulo obtiene los registros de un módulo específico
func (r *RegistroAuditoriaRepository) GetByNombreModulo(nombreModulo string, page, size int) ([]models.RegistroAuditoria, int64, error) {
	var registros []models.RegistroAuditoria
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.RegistroAuditoria{}).
		Where("nombre_modulo = ?", nombreModulo).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("nombre_modulo = ?", nombreModulo).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&registros).Error; err != nil {
		return nil, 0, err
	}

	return registros, total, nil
}

// GetByRangoFechas obtiene registros dentro de un rango de fechas
func (r *RegistroAuditoriaRepository) GetByRangoFechas(fechaInicio, fechaFin string, page, size int) ([]models.RegistroAuditoria, int64, error) {
	var registros []models.RegistroAuditoria
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.RegistroAuditoria{}).
		Where("fecha_creacion BETWEEN ? AND ?", fechaInicio, fechaFin).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("fecha_creacion BETWEEN ? AND ?", fechaInicio, fechaFin).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&registros).Error; err != nil {
		return nil, 0, err
	}

	return registros, total, nil
}

// GetByFilters obtiene registros filtrados por correo y/o regional con paginación
func (r *RegistroAuditoriaRepository) GetByFilters(correo, regional, rol string, page, size int) ([]models.RegistroAuditoria, int64, error) {
	var registros []models.RegistroAuditoria
	var total int64
	offset := (page - 1) * size

	query := r.db.Model(&models.RegistroAuditoria{})

	if correo != "" {
		query = query.Where("correo ILIKE ?", "%"+correo+"%")
	}
	if regional != "" {
		query = query.Where("regional = ?", regional)
	}
	if rol != "" {
		query = query.Where("rol = ?", rol)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&registros).Error; err != nil {
		return nil, 0, err
	}

	return registros, total, nil
}
