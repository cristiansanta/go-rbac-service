package models

import (
	"time"
)

type Role struct {
	ID                 int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Nombre             string    `json:"nombre" gorm:"type:varchar(255);not null;unique"`
	FechaCreacion      time.Time `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	FechaActualizacion time.Time `json:"fecha_actualizacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (Role) TableName() string {
	return "roles"
}

type CreateRoleRequest struct {
	Nombre string `json:"nombre" binding:"required"`
}

type RoleResponse struct {
	ID                 int       `json:"id"`
	Nombre             string    `json:"nombre"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaActualizacion time.Time `json:"fecha_actualizacion"`
}

type AssignRolePermissionsRequest struct {
	RoleID        int `json:"role_id" binding:"required"`
	ModuloID      int `json:"modulo_id" binding:"required"`
	PermisoTipoID int `json:"permiso_tipo_id" binding:"required"`
}
