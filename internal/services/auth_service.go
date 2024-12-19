package services

import (
	"auth-service/internal/constants"
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo           *repository.UserRepository
	tokenBlacklistRepo *repository.TokenBlacklistRepository
}

func NewAuthService(userRepo *repository.UserRepository, tokenBlacklistRepo *repository.TokenBlacklistRepository) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
	}
}

// GenerateToken genera un token JWT para un usuario
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Correo,
		"role":    user.Role.Nombre,
		"exp":     time.Now().Add(time.Hour * time.Duration(constants.JWTExpirationHours)).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(constants.JWTSecret))
}

func (s *AuthService) Logout(tokenString string) error {
	// Validar el token primero
	tokenMetadata, err := s.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	// Agregar el token a la blacklist
	expiresAt := time.Unix(tokenMetadata.Exp, 0)
	return s.tokenBlacklistRepo.AddToBlacklist(tokenString, expiresAt)
}

// ValidateToken valida un token JWT y retorna sus metadata
func (s *AuthService) ValidateToken(tokenString string) (*models.TokenMetadata, error) {
	// Verificar si el token está en la blacklist
	isBlacklisted, err := s.tokenBlacklistRepo.IsTokenBlacklisted(tokenString)
	if err != nil {
		return nil, fmt.Errorf("error checking token blacklist: %v", err)
	}
	if isBlacklisted {
		return nil, errors.New("token has been invalidated")
	}

	// Parsear y validar el token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea el correcto
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(constants.JWTSecret), nil
	})

	if err != nil {
		// Verificar si el error es por expiración usando jwt.v5
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expirado")
		}
		return nil, fmt.Errorf("error validando token: %v", err)
	}

	// Verificar que el token sea válido y obtener los claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	// El resto del código permanece igual...
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("user_id no encontrado en token")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("email no encontrado en token")
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("role no encontrado en token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, errors.New("exp no encontrado en token")
	}

	return &models.TokenMetadata{
		UserID: int(userID),
		Email:  email,
		Role:   role,
		Exp:    int64(exp),
	}, nil
}

// Login maneja el proceso de login
func (s *AuthService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	log.Printf("Intentando login para email: %s", req.Email)

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		log.Printf("Error al obtener usuario por email: %v", err)
		return nil, errors.New(constants.ErrInvalidCredentials)
	}
	log.Printf("Usuario encontrado con ID: %d", user.ID)

	// Cambiar esta línea
	if !user.ValidatePassword(req.Password) {
		log.Printf("Contraseña inválida para usuario: %d", user.ID)
		return nil, errors.New(constants.ErrInvalidCredentials)
	}
	log.Println("Contraseña validada correctamente")

	token, err := s.GenerateToken(user)
	if err != nil {
		log.Printf("Error al generar token: %v", err)
		return nil, err
	}

	return &models.LoginResponse{
		Token: token,
		User: models.UserResponse{
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
		},
	}, nil
}
func (s *AuthService) CheckPermission(userID int, module string, permission string) (bool, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, err
	}

	// Si es SuperAdmin, tiene todos los permisos
	if strings.EqualFold(user.Role.Nombre, constants.RoleSuperAdmin) {
		return true, nil
	}

	// Obtener los permisos del usuario para el módulo específico
	permissions, err := s.userRepo.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	// Buscar el módulo y permiso específico
	for _, modulo := range permissions.Role.ModuloPermisos {
		if strings.EqualFold(modulo.Nombre, module) {
			// Verificar si el permiso existe en el módulo
			for _, perm := range modulo.Permisos {
				if perm == permission {
					return true, nil
				}
			}
			break
		}
	}

	return false, nil
}
func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}
