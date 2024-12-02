package config

import (
	"auth-service/internal/models"
	"fmt"

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

	// Create schema if it doesn't exist
	db.Exec("CREATE SCHEMA IF NOT EXISTS testing")

	// Set default schema
	db.Exec("SET search_path TO testing")

	// Create update_fecha_actualizacion function if it doesn't exist
	db.Exec(`
        CREATE OR REPLACE FUNCTION testing.update_fecha_actualizacion()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.fecha_actualizacion = CURRENT_TIMESTAMP;
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;
    `)

	// Automigrate the models
	if err := db.AutoMigrate(
		&models.Role{},
		&models.PermisoTipo{},
		&models.Module{},
		&models.ModuloPermiso{},
		&models.RolModuloPermiso{},
		&models.User{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}
