package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	repo                 *repository.RoleRepository
	rolModuloPermisoRepo *repository.RolModuloPermisoRepository
}

func NewRoleHandler(repo *repository.RoleRepository, rmpRepo *repository.RolModuloPermisoRepository) *RoleHandler {
	return &RoleHandler{
		repo:                 repo,
		rolModuloPermisoRepo: rmpRepo,
	}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := &models.Role{
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
	}

	if err := h.repo.Create(role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.RoleResponse{
		ID:                 role.ID,
		Nombre:             role.Nombre,
		FechaCreacion:      role.FechaCreacion,
		FechaActualizacion: role.FechaActualizacion,
	})
}

func (h *RoleHandler) GetAll(c *gin.Context) {
	roles, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]models.RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = role.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *RoleHandler) AssignModulePermission(c *gin.Context) {
	var req models.AssignRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.repo.AssignModulePermission(req.RoleID, req.ModuloID, req.PermisoTipoID)
	if err != nil {
		// Verificar si el error es sobre permisos no disponibles
		if strings.Contains(err.Error(), "permisos no están disponibles") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   err.Error(),
				"message": "Solo se pueden asignar permisos que estén disponibles en el módulo",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permisos asignados exitosamente al rol",
	})
}

func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	roleID := c.Param("id")
	id, err := strconv.Atoi(roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	permissions, err := h.repo.GetRolePermissions(id)
	if err != nil {
		if err.Error() == "rol no encontrado: record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rol no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func (h *RoleHandler) RemoveModulePermission(c *gin.Context) {
	var req struct {
		RoleID        int `json:"role_id" binding:"required"`
		ModuloID      int `json:"modulo_id" binding:"required"`
		PermisoTipoID int `json:"permiso_tipo_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.RemoveModulePermission(req.RoleID, req.ModuloID, req.PermisoTipoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permiso removido exitosamente del rol",
	})
}

func (h *RoleHandler) RemoveModuleFromRole(c *gin.Context) {
	var req struct {
		RoleID   int `json:"role_id" binding:"required"`
		ModuloID int `json:"modulo_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.RemoveModuleFromRole(req.RoleID, req.ModuloID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Módulo removido exitosamente del rol",
	})
}
