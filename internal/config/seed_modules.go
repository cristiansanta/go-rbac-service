package config

import (
	"auth-service/internal/models"
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Estructura para definir los módulos y sus permisos
type ModulePermission struct {
	Nombre      string
	Descripcion string
	Permisos    []string
}

// Mapa de módulos con sus permisos específicos
var DefaultModules = []ModulePermission{
	{
		Nombre:      "Inicio",
		Descripcion: "Módulo de Inicio",
		Permisos:    []string{"R"},
	},
	{
		Nombre:      "Mapa",
		Descripcion: "Módulo de Mapa",
		Permisos:    []string{"R"},
	},
	{
		Nombre:      "Búsqueda del Ciudadano",
		Descripcion: "Módulo de Búsqueda",
		Permisos:    []string{"R"},
	},
	{
		Nombre:      "Cargar Archivo",
		Descripcion: "Módulo de Carga de Archivos",
		Permisos:    []string{"R", "W"},
	},
	{
		Nombre:      "Generar Reportes",
		Descripcion: "Módulo de Reportes",
		Permisos:    []string{"R", "X"},
	},
	{
		Nombre:      "Lista de Usuarios",
		Descripcion: "Módulo de Usuarios",
		Permisos:    []string{"R", "W", "D"},
	},
	{
		Nombre:      "Linea de Atención",
		Descripcion: "Módulo de Atención",
		Permisos:    []string{"R"},
	},
	{
		Nombre:      "Roles y Permisos",
		Descripcion: "Módulo de Roles y Permisos",
		Permisos:    []string{"R", "W", "X", "D"},
	},
}

func SeedModules(db *gorm.DB) error {
	log.Println("Iniciando seed de módulos...")

	// Obtener todos los permisos existentes
	var permisos []models.PermisoTipo
	if err := db.Find(&permisos).Error; err != nil {
		return fmt.Errorf("error obteniendo permisos: %v", err)
	}

	// Crear mapa de permisos por código para fácil acceso
	permisosPorCodigo := make(map[string]models.PermisoTipo)
	for _, permiso := range permisos {
		permisosPorCodigo[permiso.Codigo] = permiso
	}

	// Crear cada módulo con sus permisos específicos
	for _, moduleInfo := range DefaultModules {
		// Verificar si el módulo ya existe
		var existingModule models.Module
		result := db.Where("LOWER(nombre) = LOWER(?)", moduleInfo.Nombre).
			Where("fecha_eliminacion IS NULL").
			First(&existingModule)

		if result.Error == nil {
			log.Printf("Módulo %s ya existe, continuando...", moduleInfo.Nombre)
			continue
		}

		// Crear el módulo
		module := models.Module{
			Nombre:      moduleInfo.Nombre,
			Descripcion: moduleInfo.Descripcion,
		}

		if err := db.Create(&module).Error; err != nil {
			return fmt.Errorf("error creando módulo %s: %v", moduleInfo.Nombre, err)
		}

		// Asignar permisos específicos al módulo
		for _, permisoCodigo := range moduleInfo.Permisos {
			permiso, exists := permisosPorCodigo[permisoCodigo]
			if !exists {
				log.Printf("Advertencia: Permiso %s no encontrado", permisoCodigo)
				continue
			}

			moduloPermiso := models.ModuloPermiso{
				IdModulo:      module.ID,
				IdPermisoTipo: permiso.ID,
			}

			if err := db.Create(&moduloPermiso).Error; err != nil {
				return fmt.Errorf("error asignando permiso %s al módulo %s: %v",
					permisoCodigo, moduleInfo.Nombre, err)
			}
		}

		log.Printf("Módulo %s creado exitosamente con sus permisos", moduleInfo.Nombre)
	}

	log.Println("Seed de módulos completado exitosamente")
	return nil
}
