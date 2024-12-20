package models

const (
	PermisoVer        = "R" // Ver
	PermisoCreateEdit = "W" // Crear/Editar
	PermisoExportar   = "X" // Exportar
	PermisoEliminar   = "D" // Eliminar
)

type PermisoTipo struct {
	ID          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Codigo      string `json:"codigo" gorm:"type:varchar(1);not null;unique"`
	Nombre      string `json:"nombre" gorm:"type:varchar(50);not null"`
	Descripcion string `json:"descripcion,omitempty" gorm:"type:varchar(255)"`
}

func (PermisoTipo) TableName() string {
	return "permiso_tipos"
}

func (pt *PermisoTipo) ToResponse() PermisoTipoResponse {
	return PermisoTipoResponse{
		ID:          pt.ID,
		Codigo:      pt.Codigo,
		Nombre:      pt.Nombre,
		Descripcion: pt.Descripcion,
	}
}

type CreatePermisoTipoRequest struct {
	Codigo      string `json:"codigo" binding:"required"`
	Nombre      string `json:"nombre" binding:"required"`
	Descripcion string `json:"descripcion"`
}

type PermisoTipoResponse struct {
	ID          int    `json:"id"`
	Codigo      string `json:"codigo"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion,omitempty"`
}
