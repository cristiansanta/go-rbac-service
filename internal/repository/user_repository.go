package repository

import (
	"auth-service/internal/models"
	"fmt"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Role").First(&user, id).Error
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	// Trigger de actualización automática de fecha_actualizacion se maneja en la base de datos
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id int) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("correo = ?", email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) ExistsByDocumento(tipoDoc, numDoc string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).
		Where("tipo_documento = ? AND numero_documento = ?", tipoDoc, numDoc).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("correo = ?", email).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByDocumento(tipoDoc, numDoc string) (*models.User, error) {
	var user models.User
	err := r.db.Where("tipo_documento = ? AND numero_documento = ?", tipoDoc, numDoc).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %v", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByRoleID(roleID int) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id_rol = ?", roleID).Find(&users).Error
	return users, err
}

func (r *UserRepository) UpdatePassword(id int, hashedPassword string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("contraseña", hashedPassword).Error
}

// Nuevo método para obtener un usuario con sus permisos
func (r *UserRepository) GetUserWithPermissions(id int) (*models.User, []models.RolModuloPermiso, error) {
	var user models.User
	err := r.db.Preload("Role").First(&user, id).Error
	if err != nil {
		return nil, nil, fmt.Errorf("usuario no encontrado: %v", err)
	}

	var permissions []models.RolModuloPermiso
	err = r.db.Where("id_rol = ?", user.IdRol).
		Preload("Modulo").
		Preload("PermisoTipo").
		Find(&permissions).Error
	if err != nil {
		return nil, nil, fmt.Errorf("error al obtener permisos: %v", err)
	}

	return &user, permissions, nil
}

func (r *UserRepository) GetAllUsersWithPermissions() (*models.UsersPermissionsListResponse, error) {
	var users []models.User
	if err := r.db.Preload("Role").Find(&users).Error; err != nil {
		return nil, err
	}

	response := &models.UsersPermissionsListResponse{
		Total:    len(users),
		Usuarios: make([]models.UserPermissionsResponse, 0, len(users)),
	}

	for _, user := range users {
		userPerms, err := r.GetUserPermissions(user.ID)
		if err != nil {
			continue
		}
		response.Usuarios = append(response.Usuarios, *userPerms)
	}

	return response, nil
}

func (r *UserRepository) GetUserPermissions(userID int) (*models.UserPermissionsResponse, error) {
	var user models.User
	if err := r.db.Preload("Role").First(&user, userID).Error; err != nil {
		return nil, err
	}

	// Obtener permisos por módulo
	var rolModuloPermisos []models.RolModuloPermiso
	if err := r.db.Where("id_rol = ?", user.IdRol).
		Preload("Modulo").
		Preload("PermisoTipo").
		Find(&rolModuloPermisos).Error; err != nil {
		return nil, err
	}

	// Organizar permisos por módulo
	moduloPermisos := make(map[int]*models.ModuloPermissions)
	for _, rmp := range rolModuloPermisos {
		if _, exists := moduloPermisos[rmp.IdModulo]; !exists {
			moduloPermisos[rmp.IdModulo] = &models.ModuloPermissions{
				ID:       rmp.Modulo.ID,
				Nombre:   rmp.Modulo.Nombre,
				Permisos: make([]string, 0),
			}
		}
		moduloPermisos[rmp.IdModulo].Permisos = append(
			moduloPermisos[rmp.IdModulo].Permisos,
			rmp.PermisoTipo.Codigo,
		)
	}

	// Convertir map a slice
	modulePermsList := make([]models.ModuloPermissions, 0, len(moduloPermisos))
	for _, mp := range moduloPermisos {
		modulePermsList = append(modulePermsList, *mp)
	}

	return &models.UserPermissionsResponse{
		ID:              user.ID,
		Nombre:          user.Nombre,
		Apellidos:       user.Apellidos,
		TipoDocumento:   user.TipoDocumento,
		NumeroDocumento: user.NumeroDocumento,
		Correo:          user.Correo,
		Sede:            user.Sede,
		Regional:        user.Regional,
		Role: models.RolePermissions{
			ID:             user.Role.ID,
			Nombre:         user.Role.Nombre,
			ModuloPermisos: modulePermsList,
		},
	}, nil
}
