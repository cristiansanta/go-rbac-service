package config

import (
	"auth-service/internal/constants"
	"auth-service/internal/models"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func SeedSuperAdmin(db *gorm.DB) error {
	// Crear rol SuperAdmin si no existe
	superAdminRole := &models.Role{
		Nombre:      constants.RoleSuperAdmin,
		Descripcion: "Rol con acceso completo al sistema",
	}

	var existingRole models.Role
	if err := db.Where("UPPER(nombre) = ?", constants.RoleSuperAdmin).First(&existingRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(superAdminRole).Error; err != nil {
				return fmt.Errorf("error creando rol SuperAdmin: %v", err)
			}
		} else {
			return err
		}
	} else {
		superAdminRole = &existingRole
	}

	// Crear usuario SuperAdmin si no existe
	superAdmin := &models.User{
		Nombre:          "Super",
		Apellidos:       "Admin",
		TipoDocumento:   "CC",
		NumeroDocumento: "0000000000",
		Sede:            "Principal",
		IdRol:           superAdminRole.ID,
		Regional:        "Nacional",
		Correo:          "superadmin@example.com",
		Telefono:        "0000000000",
		Contraseña:      "superadmin123", // La contraseña se hasheará automáticamente por BeforeCreate
	}

	var existingUser models.User
	if err := db.Where("correo = ?", superAdmin.Correo).First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(superAdmin).Error; err != nil {
				return fmt.Errorf("error creando usuario SuperAdmin: %v", err)
			}
			log.Println("SuperAdmin creado exitosamente")
			log.Println("Email: superadmin@example.com")
			log.Println("Contraseña: superadmin123")
		} else {
			return err
		}
	}

	return nil
}
