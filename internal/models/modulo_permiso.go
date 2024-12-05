package models

import "time"

type ModuloPermiso struct {
	ID               int         `json:"id" gorm:"primaryKey;autoIncrement;type:serial"`
	IdModulo         int         `json:"id_modulo" gorm:"not null"`
	IdPermisoTipo    int         `json:"id_permiso_tipo" gorm:"not null"`
	FechaEliminacion *time.Time  `json:"fecha_eliminacion" gorm:"type:timestamp;default:null"`
	Modulo           Module      `json:"modulo" gorm:"foreignKey:IdModulo"`
	PermisoTipo      PermisoTipo `json:"permiso_tipo" gorm:"foreignKey:IdPermisoTipo"`
}

func (ModuloPermiso) TableName() string {
	return "modulo_permisos"
}
