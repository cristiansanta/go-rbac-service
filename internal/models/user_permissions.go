package models

type UserPermissionsResponse struct {
	ID              int             `json:"id"`
	Nombre          string          `json:"nombre"`
	Apellidos       string          `json:"apellidos"`
	TipoDocumento   string          `json:"tipo_documento"`
	NumeroDocumento string          `json:"numero_documento"`
	Correo          string          `json:"correo"`
	Sede            string          `json:"sede"`
	Regional        string          `json:"regional"`
	Role            RolePermissions `json:"rol"`
}

type RolePermissions struct {
	ID             int                 `json:"id"`
	Nombre         string              `json:"nombre"`
	ModuloPermisos []ModuloPermissions `json:"modulos_permisos"`
}

type ModuloPermissions struct {
	ID       int      `json:"id"`
	Nombre   string   `json:"nombre"`
	Permisos []string `json:"permisos"` // ["R", "W", "X"]
}

type UsersPermissionsListResponse struct {
	Total    int                       `json:"total"`
	Usuarios []UserPermissionsResponse `json:"usuarios"`
}

func (r *RolePermissions) ToResponse() RolePermissions {
	return RolePermissions{
		ID:             r.ID,
		Nombre:         r.Nombre,
		ModuloPermisos: r.ModuloPermisos,
	}
}

func (mp *ModuloPermissions) ToResponse() ModuloPermissions {
	return ModuloPermissions{
		ID:       mp.ID,
		Nombre:   mp.Nombre,
		Permisos: mp.Permisos,
	}
}

func (up *UserPermissionsResponse) ToResponse() UserPermissionsResponse {
	return UserPermissionsResponse{
		ID:              up.ID,
		Nombre:          up.Nombre,
		Apellidos:       up.Apellidos,
		TipoDocumento:   up.TipoDocumento,
		NumeroDocumento: up.NumeroDocumento,
		Correo:          up.Correo,
		Sede:            up.Sede,
		Regional:        up.Regional,
		Role:            up.Role.ToResponse(),
	}
}
