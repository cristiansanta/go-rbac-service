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

		// Continuar con la request
		c.Next()

		// Extraer información necesaria para el log
		userID := c.GetInt("user_id")
		username := c.GetString("user_email")
		moduleName := getModuleFromPath(c.Request.URL.Path)
		action := getActionFromMethod(c.Request.Method)
		permissionUsed := getPermissionFromContext(c)

		// Preparar valores old y new para cambios
		var oldValue, newValue models.JsonMap
		if len(bodyBytes) > 0 && isWriteOperation(c.Request.Method) {
			if err := json.Unmarshal(bodyBytes, &newValue); err != nil {
				log.Printf("Error unmarshaling request body: %v", err)
			}
		}

		// Crear el log de auditoría
		auditLog := &models.RegistroAuditoria{
			IdUsuario:       userID,
			NombreUsuario:   username,
			NombreModulo:    moduleName,
			Accion:          action,
			PermisoUsado:    permissionUsed,
			TipoEntidad:     getEntityTypeFromPath(c.Request.URL.Path),
			IdEntidad:       getEntityIDFromPath(c.Params),
			ValorAnterior:   oldValue,
			ValorNuevo:      newValue,
			DireccionIP:     c.ClientIP(),
			AgenteUsuario:   c.Request.UserAgent(),
			CodigoEstado:    c.Writer.Status(),
			RutaSolicitud:   c.Request.URL.Path,
			MetodoSolicitud: c.Request.Method,
			FechaCreacion:   startTime,
		}

		// Guardar el log de manera asíncrona
		go func(registro *models.RegistroAuditoria) {
			if err := auditService.CreateRegistro(registro); err != nil {
				log.Printf("Error al crear registro de auditoría: %v", err)
			}
		}(auditLog)
	}
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
