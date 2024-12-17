// internal/models/module.go
package models

import (
	"time"
)

// Estructura para recibir el JSON de creación masiva
type CrearModulosRequest struct {
	Modules []InfoModulo `json:"modules" binding:"required,dive"`
}

type InfoModulo struct {
	Nombre      string `json:"nombre" binding:"required"`
	Descripcion string `json:"descripcion"`
}

type Module struct {
	ID                 int           `json:"id" gorm:"primaryKey;autoIncrement"`
	Nombre             string        `json:"nombre" gorm:"type:varchar(255);not null"`
	Descripcion        string        `json:"descripcion" gorm:"type:text"`
	FechaCreacion      time.Time     `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	FechaActualizacion time.Time     `json:"fecha_actualizacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	FechaEliminacion   *time.Time    `json:"fecha_eliminacion" gorm:"type:timestamp;default:null"`
	Permisos           []PermisoTipo `json:"permisos,omitempty" gorm:"many2many:modulo_permisos;foreignKey:ID;joinForeignKey:id_modulo;References:ID;joinReferences:id_permiso_tipo"`
}

func (Module) TableName() string {
	return "modulos"
}

type ModuleResponse struct {
	ID                 int                   `json:"id"`
	Nombre             string                `json:"nombre"`
	Descripcion        string                `json:"descripcion"`
	Permisos           []PermisoTipoResponse `json:"permisos"`
	FechaCreacion      time.Time             `json:"fecha_creacion"`
	FechaActualizacion time.Time             `json:"fecha_actualizacion"`
	FechaEliminacion   *time.Time            `json:"fecha_eliminacion,omitempty"`
}

type ModuleWithPermissions struct {
	ID                 int           `json:"id"`
	Nombre             string        `json:"nombre"`
	Descripcion        string        `json:"descripcion"`
	Permisos           []PermisoTipo `json:"permisos"`
	FechaCreacion      time.Time     `json:"fecha_creacion"`
	FechaActualizacion time.Time     `json:"fecha_actualizacion"`
	FechaEliminacion   *time.Time    `json:"fecha_eliminacion,omitempty"`
}

type AssignModulePermissionsRequest struct {
	ModuloID       int   `json:"modulo_id" binding:"required"`
	PermisoTipoIDs []int `json:"permiso_tipo_ids" binding:"required"`
}

func (m *Module) ToResponse() ModuleResponse {
	permisosResponse := make([]PermisoTipoResponse, len(m.Permisos))
	for i, permiso := range m.Permisos {
		permisosResponse[i] = permiso.ToResponse()
	}

	return ModuleResponse{
		ID:                 m.ID,
		Nombre:             m.Nombre,
		Descripcion:        m.Descripcion,
		Permisos:           permisosResponse,
		FechaCreacion:      m.FechaCreacion,
		FechaActualizacion: m.FechaActualizacion,
		FechaEliminacion:   m.FechaEliminacion,
	}
}

func (m *Module) ToResponseWithPermissions() ModuleWithPermissions {
	return ModuleWithPermissions{
		ID:                 m.ID,
		Nombre:             m.Nombre,
		Descripcion:        m.Descripcion,
		Permisos:           m.Permisos,
		FechaCreacion:      m.FechaCreacion,
		FechaActualizacion: m.FechaActualizacion,
		FechaEliminacion:   m.FechaEliminacion,
	}
}

func (m *Module) HasPermission(permisoID int) bool {
	for _, permiso := range m.Permisos {
		if permiso.ID == permisoID {
			return true
		}
	}
	return false
}

func (m *Module) GetAvailablePermissions() []int {
	permisos := make([]int, len(m.Permisos))
	for i, permiso := range m.Permisos {
		permisos[i] = permiso.ID
	}
	return permisos
}
