package services

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
)

type AuditService struct {
	registroRepo *repository.RegistroAuditoriaRepository
}

func NewAuditService(registroRepo *repository.RegistroAuditoriaRepository) *AuditService {
	return &AuditService{
		registroRepo: registroRepo,
	}
}

func (s *AuditService) CreateRegistro(registro *models.RegistroAuditoria) error {
	return s.registroRepo.Create(registro)
}

func (s *AuditService) ObtenerRegistros(page, size int) ([]models.RegistroAuditoriaResponse, int64, error) {
	registros, total, err := s.registroRepo.GetAll(page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.RegistroAuditoriaResponse, len(registros))
	for i, registro := range registros {
		response[i] = registro.ToResponse()
	}

	return response, total, nil
}

func (s *AuditService) ObtenerRegistrosPorUsuario(idUsuario int, page, size int) ([]models.RegistroAuditoriaResponse, int64, error) {
	registros, total, err := s.registroRepo.GetByIdUsuario(idUsuario, page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.RegistroAuditoriaResponse, len(registros))
	for i, registro := range registros {
		response[i] = registro.ToResponse()
	}

	return response, total, nil
}

func (s *AuditService) ObtenerRegistrosPorModulo(nombreModulo string, page, size int) ([]models.RegistroAuditoriaResponse, int64, error) {
	registros, total, err := s.registroRepo.GetByNombreModulo(nombreModulo, page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.RegistroAuditoriaResponse, len(registros))
	for i, registro := range registros {
		response[i] = registro.ToResponse()
	}

	return response, total, nil
}

func (s *AuditService) ObtenerRegistrosPorRangoFechas(fechaInicio, fechaFin string, page, size int) ([]models.RegistroAuditoriaResponse, int64, error) {
	registros, total, err := s.registroRepo.GetByRangoFechas(fechaInicio, fechaFin, page, size)
	if err != nil {
		return nil, 0, err
	}

	response := make([]models.RegistroAuditoriaResponse, len(registros))
	for i, registro := range registros {
		response[i] = registro.ToResponse()
	}

	return response, total, nil
}
