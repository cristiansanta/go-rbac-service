package models

const (
	PermisoRead    = "R" // Ver
	PermisoWrite   = "W" // Crear/Editar
	PermisoExecute = "X" // Eliminar
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
