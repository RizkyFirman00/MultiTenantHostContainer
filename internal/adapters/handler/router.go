package handler

import (
	"github.com/damantine/multi-tenant-hosting/internal/core/services"
	"github.com/gin-gonic/gin"
)

func NewRouter(authSvc *services.AuthService, projectSvc *services.ProjectService) *gin.Engine {
	r := gin.Default()

	authHandler := NewAuthHandler(authSvc)
	projectHandler := NewProjectHandler(projectSvc)

	// Public routes
	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(AuthMiddleware(authSvc))
	{
		api.GET("/auth/me", authHandler.Me) // New Me endpoint
		api.POST("/projects", projectHandler.Create)
		api.POST("/projects/:id/deploy", projectHandler.Deploy)
		api.POST("/projects/:id/start", projectHandler.Start)
		api.POST("/projects/:id/stop", projectHandler.Stop)
		api.GET("/projects", projectHandler.List)
		api.GET("/projects/:id", projectHandler.Get)
		api.PUT("/projects/:id", projectHandler.Update)
		api.DELETE("/projects/:id", projectHandler.Delete)
	}

	return r
}
