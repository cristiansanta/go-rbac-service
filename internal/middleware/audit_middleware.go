package middleware

import (
	"auth-service/internal/models"
	"auth-service/internal/services"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// bodyLogWriter estructura para capturar la respuesta
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditMiddleware middleware para registrar las acciones
func AuditMiddleware(auditService *services.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Capturar el body original
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Preparar writer para capturar la respuesta
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Verificar si es ruta de login
		if c.Request.URL.Path == "/login" {
			// Continuar con la request para que se procese
			c.Next()

			// Extraer email del intento de login
			var loginData struct {
				Email string `json:"email"`
			}
			if err := json.Unmarshal(bodyBytes, &loginData); err != nil {
				log.Printf("Error unmarshaling login data: %v", err)
			}

			auditLog := &models.RegistroAuditoria{
				Correo:          loginData.Email,
				NombreModulo:    "auth",
				Accion:          "LOGIN",
				PermisoUsado:    "W",
				DireccionIP:     c.ClientIP(),
				AgenteUsuario:   c.Request.UserAgent(),
				CodigoEstado:    c.Writer.Status(),
				RutaSolicitud:   c.Request.URL.Path,
				MetodoSolicitud: c.Request.Method,
				FechaCreacion:   startTime,
			}

			// Si el login fue exitoso (código 200)
			if c.Writer.Status() == 200 {
				// Extraer información del usuario de la respuesta
				var loginResponse struct {
					User struct {
						ID       int    `json:"id"`
						Regional string `json:"regional"`
						Role     struct {
							Nombre string `json:"nombre"`
							ID     int    `json:"id"`
						} `json:"role"`
					} `json:"user"`
				}

				responseBody := blw.body.String()
				if err := json.Unmarshal([]byte(responseBody), &loginResponse); err == nil {
					auditLog.IdUsuario = loginResponse.User.ID
					auditLog.Regional = loginResponse.User.Regional
					auditLog.Rol = loginResponse.User.Role.Nombre
					auditLog.IdRol = loginResponse.User.Role.ID
					auditLog.ValorNuevo = models.JsonMap{"mensaje": "Login exitoso"}
				} else {
					log.Printf("Error unmarshaling login response: %v", err)
				}
			} else {
				auditLog.Accion = "LOGIN_FALLIDO"
				auditLog.ValorNuevo = models.JsonMap{"error": "Credenciales inválidas"}
			}

			go func(registro *models.RegistroAuditoria) {
				if err := auditService.CreateRegistro(registro); err != nil {
					log.Printf("Error al crear registro de auditoría: %v", err)
				}
			}(auditLog)

			return
		}

		// Para las demás rutas, verificar autenticación
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			auditLog := &models.RegistroAuditoria{
				IdUsuario:       0,
				Correo:          "INTENTO_NO_AUTORIZADO",
				Regional:        "",
				NombreModulo:    getModuleFromPath(c.Request.URL.Path),
				Accion:          "ACCESO_DENEGADO",
				PermisoUsado:    "",
				Rol:             "NO_AUTENTICADO",
				IdRol:           0,
				ValorAnterior:   nil,
				ValorNuevo:      models.JsonMap{"error": "Intento de acceso sin token de autenticación"},
				DireccionIP:     c.ClientIP(),
				AgenteUsuario:   c.Request.UserAgent(),
				CodigoEstado:    401,
				RutaSolicitud:   c.Request.URL.Path,
				MetodoSolicitud: c.Request.Method,
				FechaCreacion:   startTime,
			}

			go func(registro *models.RegistroAuditoria) {
				if err := auditService.CreateRegistro(registro); err != nil {
					log.Printf("Error al registrar intento no autorizado: %v", err)
				}
			}(auditLog)

			c.Next()
			return
		}

		c.Next()

		// Si el estado es 401 o 403, registrar como intento fallido
		if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
			auditLog := &models.RegistroAuditoria{
				IdUsuario:       c.GetInt("user_id"),
				Correo:          c.GetString("user_email"),
				Regional:        c.GetString("user_regional"),
				NombreModulo:    getModuleFromPath(c.Request.URL.Path),
				Accion:          "ACCESO_DENEGADO",
				PermisoUsado:    getPermissionFromContext(c),
				Rol:             c.GetString("user_role"),
				IdRol:           c.GetInt("user_role_id"),
				ValorAnterior:   nil,
				ValorNuevo:      models.JsonMap{"error": "Acceso denegado - Token inválido o permisos insuficientes"},
				DireccionIP:     c.ClientIP(),
				AgenteUsuario:   c.Request.UserAgent(),
				CodigoEstado:    c.Writer.Status(),
				RutaSolicitud:   c.Request.URL.Path,
				MetodoSolicitud: c.Request.Method,
				FechaCreacion:   startTime,
			}

			go func(registro *models.RegistroAuditoria) {
				if err := auditService.CreateRegistro(registro); err != nil {
					log.Printf("Error al registrar acceso denegado: %v", err)
				}
			}(auditLog)

			return
		}

		// Registrar acciones autorizadas normales
		userID := c.GetInt("user_id")
		username := c.GetString("user_email")
		userRegional := c.GetString("user_regional")
		userRole := c.GetString("user_role")
		userRoleID := c.GetInt("user_role_id")

		moduleName := getModuleFromPath(c.Request.URL.Path)
		action := getActionFromMethod(c.Request.Method)
		permissionUsed := getPermissionFromContext(c)

		var oldValue, newValue models.JsonMap
		if len(bodyBytes) > 0 && isWriteOperation(c.Request.Method) {
			if err := json.Unmarshal(bodyBytes, &newValue); err != nil {
				log.Printf("Error unmarshaling request body: %v", err)
			}
		}

		auditLog := &models.RegistroAuditoria{
			IdUsuario:       userID,
			Correo:          username,
			Regional:        userRegional,
			NombreModulo:    moduleName,
			Accion:          action,
			PermisoUsado:    permissionUsed,
			Rol:             userRole,
			IdRol:           userRoleID,
			ValorAnterior:   oldValue,
			ValorNuevo:      newValue,
			DireccionIP:     c.ClientIP(),
			AgenteUsuario:   c.Request.UserAgent(),
			CodigoEstado:    c.Writer.Status(),
			RutaSolicitud:   c.Request.URL.Path,
			MetodoSolicitud: c.Request.Method,
			FechaCreacion:   startTime,
		}

		go func(registro *models.RegistroAuditoria) {
			if err := auditService.CreateRegistro(registro); err != nil {
				log.Printf("Error al crear registro de auditoría: %v", err)
			}
		}(auditLog)
	}
}

// getEntityState obtiene el estado actual de una entidad
func getEntityState(db *gorm.DB, entityType string, entityID int) models.JsonMap {
	if entityID <= 0 {
		return nil
	}

	var result models.JsonMap

	switch entityType {
	case "Rol":
		var entity models.Role
		if err := db.First(&entity, entityID).Error; err == nil {
			result = convertToJsonMap(entity)
		}
	case "Usuario":
		var entity models.User
		if err := db.Preload("Role").First(&entity, entityID).Error; err == nil {
			result = convertToJsonMap(entity)
		}
	case "Módulo":
		var entity models.Module
		if err := db.Preload("Permisos").First(&entity, entityID).Error; err == nil {
			result = convertToJsonMap(entity)
		}
	case "TipoPermiso":
		var entity models.PermisoTipo
		if err := db.First(&entity, entityID).Error; err == nil {
			result = convertToJsonMap(entity)
		}
	}

	return result
}
func convertToJsonMap(v interface{}) models.JsonMap {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}

	var result models.JsonMap
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil
	}

	// Eliminar campos sensibles
	delete(result, "contraseña")
	delete(result, "password")

	return result
}

// Funciones auxiliares
func getModuleFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return "unknown"
}

func getActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "LEER"
	case "POST":
		return "CREAR"
	case "PUT", "PATCH":
		return "ACTUALIZAR"
	case "DELETE":
		return "ELIMINAR"
	default:
		return "DESCONOCIDO"
	}
}

func getPermissionFromContext(c *gin.Context) string {
	// Obtener el permiso del contexto si fue establecido por el middleware de autorización
	if perm, exists := c.Get("permission_used"); exists {
		return perm.(string)
	}
	// Inferir el permiso basado en el método HTTP
	switch c.Request.Method {
	case "GET":
		return "R"
	case "POST", "PUT", "PATCH":
		return "W"
	case "DELETE":
		return "D"
	default:
		return ""
	}
}

func getEntityTypeFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		switch parts[1] {
		case "users":
			return "Usuario"
		case "roles":
			return "Rol"
		case "modules":
			return "Módulo"
		case "permiso-tipos":
			return "TipoPermiso"
		}
	}
	return "desconocido"
}

func getEntityIDFromPath(params gin.Params) int {
	if id := params.ByName("id"); id != "" {
		if idInt, err := strconv.Atoi(id); err == nil {
			return idInt
		}
	}
	return 0
}

func isWriteOperation(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH"
}
