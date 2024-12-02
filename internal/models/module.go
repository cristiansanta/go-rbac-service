package models

import (
	"time"
)

type Module struct {
	ID                 int           `json:"id" gorm:"primaryKey;autoIncrement"`
	Nombre             string        `json:"nombre" gorm:"type:varchar(255);not null"`
	Descripcion        string        `json:"descripcion" gorm:"type:text"`
	Estado             int16         `json:"estado" gorm:"type:int2;not null;default:1"`
	FechaCreacion      time.Time     `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	FechaActualizacion time.Time     `json:"fecha_actualizacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Permisos           []PermisoTipo `json:"permisos,omitempty" gorm:"many2many:modulo_permisos;foreignKey:ID;joinForeignKey:id_modulo;References:ID;joinReferences:id_permiso_tipo"`
}

func (Module) TableName() string {
	return "modulos"
}

type CreateModuleRequest struct {
	Nombre      string `json:"nombre" binding:"required"`
	Descripcion string `json:"descripcion"`
	Estado      int16  `json:"estado" binding:"required"`
}

type ModuleResponse struct {
	ID                 int       `json:"id"`
	Nombre             string    `json:"nombre"`
	Descripcion        string    `json:"descripcion"`
	Estado             int16     `json:"estado"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaActualizacion time.Time `json:"fecha_actualizacion"`
}

type ModuleWithPermissions struct {
	ID                 int           `json:"id"`
	Nombre             string        `json:"nombre"`
	Descripcion        string        `json:"descripcion"`
	Estado             int16         `json:"estado"`
	Permisos           []PermisoTipo `json:"permisos"`
	FechaCreacion      time.Time     `json:"fecha_creacion"`
	FechaActualizacion time.Time     `json:"fecha_actualizacion"`
}

type AssignModulePermissionsRequest struct {
	ModuloID       int   `json:"modulo_id" binding:"required"`
	PermisoTipoIDs []int `json:"permiso_tipo_ids" binding:"required"`
}
