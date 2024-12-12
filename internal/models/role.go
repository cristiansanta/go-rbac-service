package models

import "time"

type Role struct {
	ID                 int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Nombre             string    `json:"nombre" gorm:"type:varchar(255);not null;unique"`
	Descripcion        string    `json:"descripcion" gorm:"type:text"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaActualizacion time.Time `json:"fecha_actualizacion"`
}

func (Role) TableName() string {
	return "roles"
}

type CreateRoleRequest struct {
	Nombre      string `json:"nombre" binding:"required"`
	Descripcion string `json:"descripcion"`
}

type RoleResponse struct {
	ID                 int       `json:"id"`
	Nombre             string    `json:"nombre"`
	Descripcion        string    `json:"descripcion"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaActualizacion time.Time `json:"fecha_actualizacion"`
}

type AssignRolePermissionsRequest struct {
	RoleID        int   `json:"role_id" binding:"required"`
	ModuloID      int   `json:"modulo_id" binding:"required"`
	PermisoTipoID []int `json:"permiso_tipo_id" binding:"required"`
}

func (r *Role) ToResponse() RoleResponse {
	return RoleResponse{
		ID:                 r.ID,
		Nombre:             r.Nombre,
		Descripcion:        r.Descripcion,
		FechaCreacion:      r.FechaCreacion,
		FechaActualizacion: r.FechaActualizacion,
	}
}
