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
