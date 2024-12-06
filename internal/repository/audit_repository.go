package repository

import (
	"auth-service/internal/models"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create registra un nuevo log de auditoría
func (r *AuditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

// GetAll obtiene todos los logs con paginación
func (r *AuditRepository) GetAll(page, size int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * size

	// Obtener el total de registros
	if err := r.db.Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Obtener los registros paginados
	if err := r.db.
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUserID obtiene los logs de un usuario específico
func (r *AuditRepository) GetByUserID(userID int, page, size int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("user_id = ?", userID).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByModuleName obtiene los logs de un módulo específico
func (r *AuditRepository) GetByModuleName(moduleName string, page, size int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.AuditLog{}).
		Where("module_name = ?", moduleName).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("module_name = ?", moduleName).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByDateRange obtiene logs dentro de un rango de fechas
func (r *AuditRepository) GetByDateRange(startDate, endDate string, page, size int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	offset := (page - 1) * size

	if err := r.db.Model(&models.AuditLog{}).
		Where("fecha_creacion BETWEEN ? AND ?", startDate, endDate).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.
		Where("fecha_creacion BETWEEN ? AND ?", startDate, endDate).
		Order("fecha_creacion DESC").
		Offset(offset).
		Limit(size).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
