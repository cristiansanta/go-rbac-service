package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login maneja la autenticación de usuarios
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
func (h *AuthHandler) Logout(c *gin.Context) {
	// Obtener el token del header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token no proporcionado"})
		return
	}

	// Extraer el token del formato "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato de token inválido"})
		return
	}

	// Invalidar el token
	if err := h.authService.Logout(parts[1]); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cerrar sesión"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sesión cerrada exitosamente"})
}
