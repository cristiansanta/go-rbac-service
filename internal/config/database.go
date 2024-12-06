package config

import (
	"auth-service/internal/models"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func SetupDatabase() (*gorm.DB, error) {
	dsn := "host=localhost user=authuser password=authpass dbname=authdb port=5432 sslmode=disable TimeZone=America/Bogota"

	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			TablePrefix:   "testing.",
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Crear schema si no existe
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS testing").Error; err != nil {
		return nil, fmt.Errorf("error creando schema: %v", err)
	}

	// Establecer schema por defecto
	if err := db.Exec("SET search_path TO testing").Error; err != nil {
		return nil, fmt.Errorf("error estableciendo schema: %v", err)
	}

	// Función de actualización
	if err := db.Exec(`
        CREATE OR REPLACE FUNCTION testing.update_fecha_actualizacion()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.fecha_actualizacion = CURRENT_TIMESTAMP;
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
    `).Error; err != nil {
		log.Printf("Error creando función update_fecha_actualizacion: %v", err)
	}

	// Automigrate
	if err := db.AutoMigrate(
		&models.Role{},
		&models.PermisoTipo{},
		&models.Module{},
		&models.ModuloPermiso{},
		&models.RolModuloPermiso{},
		&models.User{},
		&models.TokenBlacklist{},
		&models.AuditLog{}, // Nueva tabla
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	// Seed data
	if err := SeedPermisos(db); err != nil {
		log.Printf("Error en SeedPermisos: %v", err)
		return nil, fmt.Errorf("error al crear permisos: %v", err)
	}

	if err := SeedSuperAdmin(db); err != nil {
		log.Printf("Error en SeedSuperAdmin: %v", err)
		return nil, fmt.Errorf("error al crear SuperAdmin: %v", err)
	}

	return db, nil
}
