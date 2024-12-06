package constants

import "strings"

const (
	// Roles principales (pero no limitantes)
	RoleSuperAdmin  = "SUPERADMIN"
	RoleFuncionario = "FUNCIONARIO"
	RoleOperario    = "OPERARIO"

	// JWT
	JWTSecret          = "tu_secret_key_muy_segura" // En producción, usar variables de entorno
	JWTExpirationHours = 24

	// Auth errors
	ErrInvalidCredentials = "credenciales inválidas"
	ErrUnauthorized       = "no autorizado"
	ErrForbidden          = "acceso denegido"
)

// ModuleNames define los nombres de los módulos del sistema
var ModuleNames = map[string]string{
	"inicio":           "Inicio",
	"mapa":             "Mapa",
	"busqueda":         "Búsqueda del Ciudadano",
	"cargar_archivo":   "Cargar Archivo",
	"generar_reportes": "Generar Reportes",
	"lista_usuarios":   "Lista de Usuarios",
	"linea_atencion":   "Línea de Atención",
	"roles_permisos":   "Roles y Permisos",
}

// DefaultPermissions define los permisos por defecto para el SuperAdmin
var DefaultPermissions = map[string][]string{
	"inicio":           {"R", "W", "X", "D"},
	"mapa":             {"R", "W", "X", "D"},
	"busqueda":         {"R", "W", "X", "D"},
	"cargar_archivo":   {"R", "W", "X", "D"},
	"generar_reportes": {"R", "W", "X", "D"},
	"lista_usuarios":   {"R", "W", "X", "D"},
	"linea_atencion":   {"R", "W", "X", "D"},
	"roles_permisos":   {"R", "W", "X", "D"},
}

func IsSuperAdmin(roleName string) bool {
	return strings.ToUpper(roleName) == RoleSuperAdmin
}
