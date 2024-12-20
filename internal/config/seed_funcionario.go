package config

import (
	"auth-service/internal/constants"
	"auth-service/internal/models"
	"fmt"
	"log"

	"gorm.io/gorm"
)

func SeedFuncionario(db *gorm.DB) error {
	// Crear rol Funcionario si no existe
	funcionarioRole := &models.Role{
		Nombre:      constants.RoleFuncionario,
		Descripcion: "Rol para funcionarios del sistema",
	}

	var existingRole models.Role
	if err := db.Where("UPPER(nombre) = ?", constants.RoleFuncionario).First(&existingRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(funcionarioRole).Error; err != nil {
				return fmt.Errorf("error creando rol Funcionario: %v", err)
			}
			log.Println("Rol Funcionario creado exitosamente")
		} else {
			return err
		}
	} else {
		log.Println("Rol Funcionario ya existe")
	}

	return nil
}
