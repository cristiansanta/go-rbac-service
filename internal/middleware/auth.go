package middleware

import (
	"auth-service/internal/constants"
	"auth-service/internal/models"
	"auth-service/internal/services"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authentication verifica el token JWT y establece el usuario en el contexto
func (m *AuthMiddleware) Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token no proporcionado"})
			c.Abort()
			return
		}

		// Extraer el token del header "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "formato de token inválido"})
			c.Abort()
			return
		}

		tokenMetadata, err := m.authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			c.Abort()
			return
		}

		// Obtener información adicional del usuario
		user, err := m.authService.GetUserByID(tokenMetadata.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no encontrado"})
			c.Abort()
			return
		}

		// Guardar los datos del usuario en el contexto
		c.Set("user_id", tokenMetadata.UserID)
		c.Set("user_email", tokenMetadata.Email)
		c.Set("user_role", tokenMetadata.Role)
		c.Set("user_regional", user.Regional) // Añadir el regional al contexto

		c.Next()
	}
}

// Authorization verifica los permisos del usuario para acceder a un recurso
func (m *AuthMiddleware) Authorization(module string, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")

		// Si es SuperAdmin, permitir todo excepto modificar su propio rol
		if strings.ToUpper(userRole) == constants.RoleSuperAdmin {
			if isModifyingOwnRole(c) {
				c.JSON(http.StatusForbidden, gin.H{"error": "no puede modificar su propio rol"})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Verificar permisos específicos para otros roles
		userID := c.GetInt("user_id")
		authorized, err := m.authService.CheckPermission(userID, module, permission)
		if err != nil || !authorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "no tiene permiso para realizar esta acción"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isModifyingOwnRole verifica si un usuario está intentando modificar su propio rol
// Modificar esta función
func isModifyingOwnRole(c *gin.Context) bool {
	userID := c.GetInt("user_id")
	targetUserID := c.Param("id")
	userRole := c.GetString("user_role")

	// Si es SuperAdmin y está intentando actualizarse a sí mismo
	if strings.ToUpper(userRole) == constants.RoleSuperAdmin && userID == parseInt(targetUserID) {
		var updateReq models.UpdateUserRequest
		// Crear una copia del body
		bodyBytes, _ := c.GetRawData()
		// Restaurar el body para futuras lecturas
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := json.Unmarshal(bodyBytes, &updateReq); err != nil {
			return false
		}

		// Solo retornar true si intenta modificar el id_rol
		return updateReq.IdRol != 0
	}

	return false
}

// parseInt convierte un string a int de manera segura
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}
