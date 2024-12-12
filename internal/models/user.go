package models

import (
	"auth-service/internal/constants"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                 int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Nombre             string    `json:"nombre" gorm:"type:varchar(100);not null"`
	Apellidos          string    `json:"apellidos" gorm:"type:varchar(100);not null"`
	TipoDocumento      string    `json:"tipo_documento" gorm:"type:varchar(20);not null"`
	NumeroDocumento    string    `json:"numero_documento" gorm:"type:varchar(20);not null;unique"`
	Sede               string    `json:"sede" gorm:"type:varchar(100);not null"`
	IdRol              int       `json:"id_rol" gorm:"not null"`
	Role               Role      `json:"role" gorm:"foreignKey:IdRol"`
	Regional           string    `json:"regional" gorm:"type:varchar(100);not null"`
	Correo             string    `json:"correo" gorm:"type:varchar(100);not null;unique"`
	Telefono           string    `json:"telefono" gorm:"type:varchar(20)"`
	Contraseña         string    `json:"-" gorm:"column:contraseña;type:varchar(255);not null"`
	FechaCreacion      time.Time `json:"fecha_creacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	FechaActualizacion time.Time `json:"fecha_actualizacion" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (User) TableName() string {
	return "usuarios"
}

// BeforeCreate se ejecuta antes de crear un usuario
func (u *User) BeforeCreate(tx *gorm.DB) error {
	log.Printf("BeforeCreate: Hasheando contraseña para usuario: %s", u.Correo)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Contraseña), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hasheando contraseña: %v", err)
		return err
	}
	u.Contraseña = string(hashedPassword)
	log.Printf("Contraseña hasheada exitosamente para usuario: %s", u.Correo)
	return nil
}

// ComparePassword compara la contraseña proporcionada con el hash almacenado
func (u *User) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Contraseña), []byte(password))
	return err == nil
}

// Requests structs se mantienen igual que en el código original
type CreateUserRequest struct {
	Nombre          string `json:"nombre" binding:"required"`
	Apellidos       string `json:"apellidos" binding:"required"`
	TipoDocumento   string `json:"tipo_documento" binding:"required"`
	NumeroDocumento string `json:"numero_documento" binding:"required"`
	Sede            string `json:"sede" binding:"required"`
	IdRol           int    `json:"id_rol" binding:"required"`
	Regional        string `json:"regional" binding:"required"`
	Correo          string `json:"correo" binding:"required,email"`
	Telefono        string `json:"telefono" binding:"required"`
	Contraseña      string `json:"contraseña" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Nombre          string `json:"nombre,omitempty"`
	Apellidos       string `json:"apellidos,omitempty"`
	TipoDocumento   string `json:"tipo_documento,omitempty"`
	NumeroDocumento string `json:"numero_documento,omitempty"`
	Sede            string `json:"sede,omitempty"`
	Regional        string `json:"regional,omitempty"`
	Correo          string `json:"correo,omitempty"`
	Telefono        string `json:"telefono,omitempty"`
	IdRol           int    `json:"id_rol,omitempty"`
}

type UserResponse struct {
	ID                 int       `json:"id"`
	Nombre             string    `json:"nombre"`
	Apellidos          string    `json:"apellidos"`
	TipoDocumento      string    `json:"tipo_documento"`
	NumeroDocumento    string    `json:"numero_documento"`
	Sede               string    `json:"sede"`
	IdRol              int       `json:"id_rol"`
	Role               Role      `json:"role"`
	Regional           string    `json:"regional"`
	Correo             string    `json:"correo"`
	Telefono           string    `json:"telefono"`
	FechaCreacion      time.Time `json:"fecha_creacion"`
	FechaActualizacion time.Time `json:"fecha_actualizacion"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                 u.ID,
		Nombre:             u.Nombre,
		Apellidos:          u.Apellidos,
		TipoDocumento:      u.TipoDocumento,
		NumeroDocumento:    u.NumeroDocumento,
		Sede:               u.Sede,
		IdRol:              u.IdRol,
		Role:               u.Role,
		Regional:           u.Regional,
		Correo:             u.Correo,
		Telefono:           u.Telefono,
		FechaCreacion:      u.FechaCreacion,
		FechaActualizacion: u.FechaActualizacion,
	}
}
func (u *User) IsSuperAdmin() bool {
	return u.Role.Nombre == constants.RoleSuperAdmin
}

// HasPermission verifica si el usuario tiene un permiso específico
func (u *User) HasPermission(moduleCode string, permissionCode string) bool {
	if u.IsSuperAdmin() {
		return true
	}

	// Implementaremos esto más adelante cuando tengamos la lógica de permisos
	return false
}
