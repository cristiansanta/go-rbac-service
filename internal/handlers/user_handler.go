package handlers

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo     *repository.UserRepository
	roleRepo *repository.RoleRepository
}

func NewUserHandler(repo *repository.UserRepository, roleRepo *repository.RoleRepository) *UserHandler {
	return &UserHandler{
		repo:     repo,
		roleRepo: roleRepo,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar si el rol existe
	_, err := h.roleRepo.GetByID(req.IdRol)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El rol especificado no existe"})
		return
	}

	// Verificar si ya existe un usuario con el mismo correo
	exists, err := h.repo.ExistsByEmail(req.Correo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Ya existe un usuario con este correo"})
		return
	}

	// Verificar si ya existe un usuario con el mismo documento
	exists, err = h.repo.ExistsByDocumento(req.TipoDocumento, req.NumeroDocumento)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Ya existe un usuario con este documento"})
		return
	}

	user := &models.User{
		Nombre:          req.Nombre,
		Apellidos:       req.Apellidos,
		TipoDocumento:   req.TipoDocumento,
		NumeroDocumento: req.NumeroDocumento,
		Sede:            req.Sede,
		IdRol:           req.IdRol,
		Regional:        req.Regional,
		Correo:          req.Correo,
		Telefono:        req.Telefono,
		Contraseña:      req.Contraseña,
	}

	if err := h.repo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Obtener el usuario con su rol para la respuesta
	createdUser, err := h.repo.GetByID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		ID:                 createdUser.ID,
		Nombre:             createdUser.Nombre,
		Apellidos:          createdUser.Apellidos,
		TipoDocumento:      createdUser.TipoDocumento,
		NumeroDocumento:    createdUser.NumeroDocumento,
		Sede:               createdUser.Sede,
		IdRol:              createdUser.IdRol,
		Role:               createdUser.Role,
		Regional:           createdUser.Regional,
		Correo:             createdUser.Correo,
		Telefono:           createdUser.Telefono,
		FechaCreacion:      createdUser.FechaCreacion,
		FechaActualizacion: createdUser.FechaActualizacion,
	})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]models.UserResponse, len(users))
	for i, user := range users {
		response[i] = models.UserResponse{
			ID:                 user.ID,
			Nombre:             user.Nombre,
			Apellidos:          user.Apellidos,
			TipoDocumento:      user.TipoDocumento,
			NumeroDocumento:    user.NumeroDocumento,
			Sede:               user.Sede,
			IdRol:              user.IdRol,
			Role:               user.Role,
			Regional:           user.Regional,
			Correo:             user.Correo,
			Telefono:           user.Telefono,
			FechaCreacion:      user.FechaCreacion,
			FechaActualizacion: user.FechaActualizacion,
		}
	}

	c.JSON(http.StatusOK, response)
}
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	user, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		ID:                 user.ID,
		Nombre:             user.Nombre,
		Apellidos:          user.Apellidos,
		TipoDocumento:      user.TipoDocumento,
		NumeroDocumento:    user.NumeroDocumento,
		Sede:               user.Sede,
		IdRol:              user.IdRol,
		Role:               user.Role,
		Regional:           user.Regional,
		Correo:             user.Correo,
		Telefono:           user.Telefono,
		FechaCreacion:      user.FechaCreacion,
		FechaActualizacion: user.FechaActualizacion,
	})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Actualizar solo los campos proporcionados
	if req.Nombre != "" {
		user.Nombre = req.Nombre
	}
	if req.Apellidos != "" {
		user.Apellidos = req.Apellidos
	}
	if req.TipoDocumento != "" {
		user.TipoDocumento = req.TipoDocumento
	}
	if req.NumeroDocumento != "" {
		user.NumeroDocumento = req.NumeroDocumento
	}
	if req.Sede != "" {
		user.Sede = req.Sede
	}
	if req.IdRol != 0 {
		user.IdRol = req.IdRol
	}
	if req.Regional != "" {
		user.Regional = req.Regional
	}
	if req.Correo != "" {
		user.Correo = req.Correo
	}
	if req.Telefono != "" {
		user.Telefono = req.Telefono
	}
	if req.Contraseña != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Contraseña), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al procesar la contraseña"})
			return
		}
		user.Contraseña = string(hashedPassword)
	}

	if err := h.repo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.UserResponse{
		ID:                 user.ID,
		Nombre:             user.Nombre,
		Apellidos:          user.Apellidos,
		TipoDocumento:      user.TipoDocumento,
		NumeroDocumento:    user.NumeroDocumento,
		Sede:               user.Sede,
		IdRol:              user.IdRol,
		Role:               user.Role,
		Regional:           user.Regional,
		Correo:             user.Correo,
		Telefono:           user.Telefono,
		FechaCreacion:      user.FechaCreacion,
		FechaActualizacion: user.FechaActualizacion,
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado exitosamente"})
}

func (h *UserHandler) GetAllUsersWithPermissions(c *gin.Context) {
	response, err := h.repo.GetAllUsersWithPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	permissions, err := h.repo.GetUserPermissions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}
