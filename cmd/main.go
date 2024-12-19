package main

import (
	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/services"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupTokenCleanup(tokenBlacklistRepo *repository.TokenBlacklistRepository) func() {
	ticker := time.NewTicker(24 * time.Hour)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := tokenBlacklistRepo.CleanupExpiredTokens(); err != nil {
					log.Printf("Error cleaning up expired tokens: %v", err)
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}

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
	tokenBlacklistRepo := repository.NewTokenBlacklistRepository(db)
	registroAuditoriaRepo := repository.NewRegistroAuditoriaRepository(db)

	// Iniciar la tarea de limpieza de tokens
	cleanup := setupTokenCleanup(tokenBlacklistRepo)
	defer cleanup()

	// Initialize services
	authService := services.NewAuthService(userRepo, tokenBlacklistRepo)
	auditService := services.NewAuditService(registroAuditoriaRepo)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Initialize handlers
	roleHandler := handlers.NewRoleHandler(roleRepo, rolModuloPermisoRepo)
	permisoTipoHandler := handlers.NewPermisoTipoHandler(permisoTipoRepo)
	moduleHandler := handlers.NewModuleHandler(moduleRepo)
	userHandler := handlers.NewUserHandler(userRepo, roleRepo)
	authHandler := handlers.NewAuthHandler(authService)
	auditHandler := handlers.NewAuditHandler(auditService)

	// Setup Gin router
	r := gin.Default()

	// Configurar CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"Accept",
		"X-Requested-With",
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	r.Use(cors.New(config))

	// Public routes
	r.POST("/login", authHandler.Login)
	r.POST("/logout", authMiddleware.Authentication(), authHandler.Logout)
	r.Use(middleware.AuditMiddleware(auditService))

	// Protected routes
	protected := r.Group("")
	protected.Use(authMiddleware.Authentication())

	// User routes
	userRoutes := protected.Group("/users")
	{
		userRoutes.POST("", authMiddleware.Authorization("lista_usuarios", "W"), userHandler.Create)
		userRoutes.GET("", authMiddleware.Authorization("lista_usuarios", "R"), userHandler.GetAll)
		userRoutes.GET("/:id", authMiddleware.Authorization("lista_usuarios", "R"), userHandler.GetByID)
		userRoutes.PUT("/:id", authMiddleware.Authorization("lista_usuarios", "W"), userHandler.Update)
		userRoutes.POST("/:id/password", authMiddleware.Authorization("lista_usuarios", "W"), userHandler.ChangePassword)
		userRoutes.DELETE("/:id", authMiddleware.Authorization("lista_usuarios", "D"), userHandler.Delete)
		userRoutes.GET("/permissions", authMiddleware.Authorization("roles_permisos", "R"), userHandler.GetAllUsersWithPermissions)
		userRoutes.GET("/:id/permissions", authMiddleware.Authorization("roles_permisos", "R"), userHandler.GetUserPermissions)
	}

	// Role routes
	roleRoutes := protected.Group("/roles")
	{
		roleRoutes.POST("", authMiddleware.Authorization("roles_permisos", "W"), roleHandler.Create)
		roleRoutes.GET("", authMiddleware.Authorization("roles_permisos", "R"), roleHandler.GetAll)
		roleRoutes.POST("/assign-permission", authMiddleware.Authorization("roles_permisos", "W"), roleHandler.AssignModulePermission)
		roleRoutes.GET("/:id/permissions", authMiddleware.Authorization("roles_permisos", "R"), roleHandler.GetRolePermissions)
		roleRoutes.DELETE("/remove-permission", authMiddleware.Authorization("roles_permisos", "D"), roleHandler.RemoveModulePermission)
		roleRoutes.DELETE("/remove-module", authMiddleware.Authorization("roles_permisos", "D"), roleHandler.RemoveModuleFromRole)
	}

	// Permiso Tipo routes
	permisoTipoRoutes := protected.Group("/permiso-tipos")
	{
		permisoTipoRoutes.POST("", authMiddleware.Authorization("roles_permisos", "W"), permisoTipoHandler.Create)
		permisoTipoRoutes.GET("", authMiddleware.Authorization("roles_permisos", "R"), permisoTipoHandler.GetAll)
		permisoTipoRoutes.GET("/:id", authMiddleware.Authorization("roles_permisos", "R"), permisoTipoHandler.GetByID)
	}

	// Module routes - Actualizado sin la ruta de creación
	moduleRoutes := protected.Group("/modules")
	{
		moduleRoutes.GET("", authMiddleware.Authorization("roles_permisos", "R"), moduleHandler.GetAll)
		moduleRoutes.GET("/:id", authMiddleware.Authorization("roles_permisos", "R"), moduleHandler.GetModuleWithPermissions)
		moduleRoutes.DELETE("/:id", authMiddleware.Authorization("roles_permisos", "D"), moduleHandler.Delete)
		moduleRoutes.DELETE("/remove-permission", authMiddleware.Authorization("roles_permisos", "D"), moduleHandler.RemovePermission)
		moduleRoutes.POST("/:id/restore", authMiddleware.Authorization("roles_permisos", "W"), moduleHandler.Restore)
		moduleRoutes.GET("/deleted", authMiddleware.Authorization("roles_permisos", "R"), moduleHandler.GetDeletedModules)
	}

	// Audit routes
	auditRoutes := protected.Group("/audit")
	{
		auditRoutes.GET("/logs", authMiddleware.Authorization("roles_permisos", "R"), auditHandler.GetLogs)
		auditRoutes.GET("/logs/user/:user_id", authMiddleware.Authorization("roles_permisos", "R"), auditHandler.GetLogsByUser)
		auditRoutes.GET("/logs/module/:module_name", authMiddleware.Authorization("roles_permisos", "R"), auditHandler.GetLogsByModule)
		auditRoutes.GET("/logs/date-range", authMiddleware.Authorization("roles_permisos", "R"), auditHandler.GetLogsByDateRange)
		auditRoutes.GET("/logs/filter", authMiddleware.Authorization("roles_permisos", "R"), auditHandler.GetLogsByFilters)
	}

	// Configurar el servidor HTTP con Graceful Shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Iniciar el servidor en una goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Esperar señal de término
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Cerrar el servidor gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
