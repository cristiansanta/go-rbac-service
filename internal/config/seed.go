package config

import (
	"auth-service/internal/models"

	"gorm.io/gorm"
)

func SeedPermisos(db *gorm.DB) error {
	permisos := []models.PermisoTipo{
		{
			Codigo:      "R",
			Nombre:      "Ver",
			Descripcion: "Permiso de lectura/visualización",
		},
		{
			Codigo:      "W",
			Nombre:      "Crear/Editar",
			Descripcion: "Permiso de creación y edición",
		},
		{
			Codigo:      "X",
			Nombre:      "Exportar",
			Descripcion: "Permiso de exportación",
		},
		{
			Codigo:      "D",
			Nombre:      "Eliminar",
			Descripcion: "Permiso de eliminación",
		},
	}

	for _, permiso := range permisos {
		var exists models.PermisoTipo
		if err := db.Where("codigo = ?", permiso.Codigo).First(&exists).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&permiso).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}
