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

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Obtiene el estado anterior de la entidad antes de la operación
func getResourceState(c *gin.Context, db *gorm.DB) models.JsonMap {
	id := c.Param("id")
	if id == "" {
		return nil
	}

	entityType := getEntityTypeFromPath(c.Request.URL.Path)
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil
	}

	return getEntityState(db, entityType, idInt)
}

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

		// Obtener estado anterior para operaciones de modificación
		var oldValue models.JsonMap
		if c.Request.Method == "PUT" || c.Request.Method == "PATCH" || c.Request.Method == "DELETE" {
			oldValue = getResourceState(c, auditService.GetDB())
		}

		// Lógica existente para login
		if c.Request.URL.Path == "/login" {
			c.Next()

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

			if c.Writer.Status() == 200 {
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

		// Verificación de autenticación
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

		// Manejo de errores 401/403
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

		// Preparar el valor nuevo según el método HTTP
		var newValue models.JsonMap
		switch c.Request.Method {
		case "POST":
			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &newValue); err != nil {
					log.Printf("Error unmarshaling request body: %v", err)
				}
			}
		case "PUT", "PATCH":
			if len(bodyBytes) > 0 {
				if err := json.Unmarshal(bodyBytes, &newValue); err != nil {
					log.Printf("Error unmarshaling request body: %v", err)
				}
			}
		case "DELETE":
			newValue = models.JsonMap{"mensaje": "Recurso eliminado"}
		case "GET":
			if len(c.Request.URL.RawQuery) > 0 {
				queryParams := make(map[string]string)
				for key, values := range c.Request.URL.Query() {
					if len(values) > 0 {
						queryParams[key] = values[0]
					}
				}
				if len(queryParams) > 0 {
					newValue = models.JsonMap{"criterios_busqueda": queryParams}
				}
			}
		}

		// Registrar la acción
		auditLog := &models.RegistroAuditoria{
			IdUsuario:       c.GetInt("user_id"),
			Correo:          c.GetString("user_email"),
			Regional:        c.GetString("user_regional"),
			NombreModulo:    getModuleFromPath(c.Request.URL.Path),
			Accion:          getActionFromMethod(c.Request.Method),
			PermisoUsado:    getPermissionFromContext(c),
			Rol:             c.GetString("user_role"),
			IdRol:           c.GetInt("user_role_id"),
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

// Mantener todas las funciones auxiliares existentes
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

	delete(result, "contraseña")
	delete(result, "password")

	return result
}

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
	if perm, exists := c.Get("permission_used"); exists {
		return perm.(string)
	}
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
