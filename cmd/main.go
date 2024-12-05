package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/repository"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Setup database connection
	db, err := config.SetupDatabase()
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Initialize repositories
	roleRepo := repository.NewRoleRepository(db)
	permisoTipoRepo := repository.NewPermisoTipoRepository(db)
	moduleRepo := repository.NewModuleRepository(db)
	userRepo := repository.NewUserRepository(db)
	rolModuloPermisoRepo := repository.NewRolModuloPermisoRepository(db)

	// Initialize handlers
	roleHandler := handlers.NewRoleHandler(roleRepo, rolModuloPermisoRepo)
	permisoTipoHandler := handlers.NewPermisoTipoHandler(permisoTipoRepo)
	moduleHandler := handlers.NewModuleHandler(moduleRepo)
	userHandler := handlers.NewUserHandler(userRepo, roleRepo)

	// Setup Gin router
	r := gin.Default()

	// User routes
	userRoutes := r.Group("/users")
	{
		userRoutes.POST("", userHandler.Create)
		userRoutes.GET("", userHandler.GetAll)
		userRoutes.GET("/:id", userHandler.GetByID)
		userRoutes.PUT("/:id", userHandler.Update)                   // Actualización general
		userRoutes.POST("/:id/password", userHandler.ChangePassword) // Cambio de contraseña
		userRoutes.DELETE("/:id", userHandler.Delete)
		userRoutes.GET("/permissions", userHandler.GetAllUsersWithPermissions)
		userRoutes.GET("/:id/permissions", userHandler.GetUserPermissions)
	}

	// Role routes
	roleRoutes := r.Group("/roles")
	{
		roleRoutes.POST("", roleHandler.Create)
		roleRoutes.GET("", roleHandler.GetAll)
		roleRoutes.POST("/assign-permission", roleHandler.AssignModulePermission)
		roleRoutes.GET("/:id/permissions", roleHandler.GetRolePermissions)
		roleRoutes.DELETE("/remove-permission", roleHandler.RemoveModulePermission)
		// Nueva ruta para eliminar un módulo completo de un rol
		roleRoutes.DELETE("/remove-module", roleHandler.RemoveModuleFromRole)
	}

	// Permiso Tipo routes
	permisoTipoRoutes := r.Group("/permiso-tipos")
	{
		permisoTipoRoutes.POST("", permisoTipoHandler.Create)
		permisoTipoRoutes.GET("", permisoTipoHandler.GetAll)
		permisoTipoRoutes.GET("/:id", permisoTipoHandler.GetByID)
	}

	// Module routes
	moduleRoutes := r.Group("/modules")
	{
		moduleRoutes.POST("", moduleHandler.Create)
		moduleRoutes.GET("", moduleHandler.GetAll)
		moduleRoutes.GET("/:id/permissions", moduleHandler.GetModuleWithPermissions)
		moduleRoutes.POST("/assign-permissions", moduleHandler.AssignPermissions)
		// Nuevas rutas para módulos
		moduleRoutes.DELETE("/:id", moduleHandler.Delete)
		moduleRoutes.DELETE("/remove-permission", moduleHandler.RemovePermission)
		moduleRoutes.POST("/:id/restore", moduleHandler.Restore)
		moduleRoutes.GET("/deleted", moduleHandler.GetDeletedModules)
	}

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
