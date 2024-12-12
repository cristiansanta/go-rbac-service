package models

import (
	"time"
)

type RolModuloPermiso struct {
	ID               int         `json:"id" gorm:"primaryKey;autoIncrement"`
	IdRol            int         `json:"id_rol" gorm:"not null"`
	IdModulo         int         `json:"id_modulo" gorm:"not null"`
	IdPermisoTipo    int         `json:"id_permiso_tipo" gorm:"not null"`
	FechaCreacion    time.Time   `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	FechaEliminacion *time.Time  `json:"fecha_eliminacion" gorm:"type:timestamp;default:null"`
	Role             Role        `json:"role" gorm:"foreignKey:IdRol"`
	Modulo           Module      `json:"modulo" gorm:"foreignKey:IdModulo"`
	PermisoTipo      PermisoTipo `json:"permiso_tipo" gorm:"foreignKey:IdPermisoTipo"`
}

func (RolModuloPermiso) TableName() string {
	return "rol_modulo_permisos"
}

type RolModuloPermisoResponse struct {
	ID            int            `json:"id"`
	IdRol         int            `json:"id_rol"`
	IdModulo      int            `json:"id_modulo"`
	IdPermisoTipo int            `json:"id_permiso_tipo"`
	FechaCreacion time.Time      `json:"fecha_creacion"`
	Role          Role           `json:"role"`
	Modulo        ModuleResponse `json:"modulo"`
	PermisoTipo   PermisoTipo    `json:"permiso_tipo"`
}
