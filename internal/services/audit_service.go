package services

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
)

type AuditService struct {
	auditRepo *repository.AuditRepository
}

func NewAuditService(auditRepo *repository.AuditRepository) *AuditService {
	return &AuditService{
		auditRepo: auditRepo,
	}
}

func (s *AuditService) CreateLog(log *models.AuditLog) error {
	return s.auditRepo.Create(log)
}

func (s *AuditService) GetLogs(page, size int) ([]models.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetAll(page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = log.ToResponse()
	}

	return response, total, nil
}

func (s *AuditService) GetLogsByUser(userID int, page, size int) ([]models.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetByUserID(userID, page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = log.ToResponse()
	}

	return response, total, nil
}

func (s *AuditService) GetLogsByModule(moduleName string, page, size int) ([]models.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetByModuleName(moduleName, page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = log.ToResponse()
	}

	return response, total, nil
}

func (s *AuditService) GetLogsByDateRange(startDate, endDate string, page, size int) ([]models.AuditLogResponse, int64, error) {
	logs, total, err := s.auditRepo.GetByDateRange(startDate, endDate, page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.AuditLogResponse, len(logs))
	for i, log := range logs {
		response[i] = log.ToResponse()
	}

	return response, total, nil
}
