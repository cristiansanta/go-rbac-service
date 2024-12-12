package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ModuleHandler struct {
	repo *repository.ModuleRepository
}

func NewModuleHandler(repo *repository.ModuleRepository) *ModuleHandler {
	return &ModuleHandler{repo: repo}
}

func (h *ModuleHandler) GetAll(c *gin.Context) {
	modules, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]models.ModuleResponse, len(modules))
	for i, module := range modules {
		response[i] = module.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *ModuleHandler) GetModuleWithPermissions(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	moduleWithPermissions, err := h.repo.GetModuleWithPermissions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, moduleWithPermissions)
}

func (h *ModuleHandler) Create(c *gin.Context) {
	var req models.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	module := &models.Module{
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
	}

	if err := h.repo.Create(module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, module.ToResponse())
}

func (h *ModuleHandler) AssignPermissions(c *gin.Context) {
	var req models.AssignModulePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.AssignPermissions(req.ModuloID, req.PermisoTipoIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	moduleWithPermissions, err := h.repo.GetModuleWithPermissions(req.ModuloID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, moduleWithPermissions)
}

func (h *ModuleHandler) RemovePermission(c *gin.Context) {
	var req struct {
		ModuloID      int `json:"modulo_id" binding:"required"`
		PermisoTipoID int `json:"permiso_tipo_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.RemovePermission(req.ModuloID, req.PermisoTipoID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Permiso removido exitosamente del módulo",
	})
}

func (h *ModuleHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Módulo eliminado exitosamente",
	})
}
func (h *ModuleHandler) Restore(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.repo.Restore(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	moduleWithPermissions, err := h.repo.GetModuleWithPermissions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Módulo restaurado exitosamente",
		"module":  moduleWithPermissions,
	})
}
func (h *ModuleHandler) GetDeletedModules(c *gin.Context) {
	modules, err := h.repo.GetDeletedModules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]models.ModuleResponse, len(modules))
	for i, module := range modules {
		response[i] = models.ModuleResponse{
			ID:                 module.ID,
			Nombre:             module.Nombre,
			Descripcion:        module.Descripcion,
			FechaCreacion:      module.FechaCreacion,
			FechaActualizacion: module.FechaActualizacion,
			FechaEliminacion:   module.FechaEliminacion,
		}
	}

	c.JSON(http.StatusOK, response)
}
