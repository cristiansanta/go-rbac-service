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
	dsn := "host=localhost user=authuser password=authpass dbname=authdb port=5433 sslmode=disable TimeZone=America/Bogota"

	// Primero conectar sin configuración específica de schema
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

	// Reconectar con la configuración correcta
	db, err = gorm.Open(postgres.Open(dsn+" search_path=testing"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect with schema: %v", err)
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

	// Automigrate con el schema correcto
	if err := db.AutoMigrate(
		&models.Role{},
		&models.PermisoTipo{},
		&models.Module{},
		&models.ModuloPermiso{},
		&models.RolModuloPermiso{},
		&models.User{},
		&models.TokenBlacklist{},
		&models.RegistroAuditoria{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	// Seed data
	if err := SeedPermisos(db); err != nil {
		log.Printf("Error en SeedPermisos: %v", err)
		return nil, fmt.Errorf("error al crear permisos: %v", err)
	}

	if err := SeedModules(db); err != nil {
		log.Printf("Error en SeedModules: %v", err)
		return nil, fmt.Errorf("error al crear módulos: %v", err)
	}

	if err := SeedSuperAdmin(db); err != nil {
		log.Printf("Error en SeedSuperAdmin: %v", err)
		return nil, fmt.Errorf("error al crear SuperAdmin: %v", err)
	}

	// Agregar el nuevo seeder de Funcionario antes del SuperAdmin
	if err := SeedFuncionario(db); err != nil {
		log.Printf("Error en SeedFuncionario: %v", err)
		return nil, fmt.Errorf("error al crear rol Funcionario: %v", err)
	}

	if err := SeedSuperAdmin(db); err != nil {
		log.Printf("Error en SeedSuperAdmin: %v", err)
		return nil, fmt.Errorf("error al crear SuperAdmin: %v", err)
	}

	return db, nil
}
