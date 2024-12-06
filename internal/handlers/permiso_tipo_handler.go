package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermisoTipoHandler struct {
	repo *repository.PermisoTipoRepository
}

func NewPermisoTipoHandler(repo *repository.PermisoTipoRepository) *PermisoTipoHandler {
	return &PermisoTipoHandler{repo: repo}
}

func (h *PermisoTipoHandler) Create(c *gin.Context) {
	var req models.CreatePermisoTipoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permisoTipo := &models.PermisoTipo{
		Codigo:      req.Codigo,
		Nombre:      req.Nombre,
		Descripcion: req.Descripcion,
	}

	if err := h.repo.Create(permisoTipo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, permisoTipo.ToResponse())
}

func (h *PermisoTipoHandler) GetAll(c *gin.Context) {
	permisoTipos, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]models.PermisoTipoResponse, len(permisoTipos))
	for i, pt := range permisoTipos {
		response[i] = pt.ToResponse()
	}

	c.JSON(http.StatusOK, response)
}

func (h *PermisoTipoHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	permisoTipo, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permisoTipo.ToResponse())
}
