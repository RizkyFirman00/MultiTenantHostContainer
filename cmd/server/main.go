package main

import (
	"context"
	"log"
	"os"

	"github.com/damantine/multi-tenant-hosting/internal/adapters/docker"
	"github.com/damantine/multi-tenant-hosting/internal/adapters/handler"
	"github.com/damantine/multi-tenant-hosting/internal/adapters/repository"
	"github.com/damantine/multi-tenant-hosting/internal/core/domain"
	"github.com/damantine/multi-tenant-hosting/internal/core/services"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=multitenant port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}
	
	var db *gorm.DB
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Running in Memory/Mock mode not implemented fully.", err)
	} else {
		log.Println("Database connected. Running migrations...")
		db.AutoMigrate(&domain.User{}, &domain.Project{}, &domain.EnvVar{}, &domain.Deployment{})
	}

	dockerClient, err := docker.NewDockerClient()
	if err != nil {
		log.Fatalf("Failed to init Docker client: %v", err)
	}

	projectRepo := repository.NewGormProjectRepository(db)

	authService := services.NewAuthService(db, "rahasia-negara-dont-use-in-prod")
	projectService := services.NewProjectService(projectRepo, dockerClient)

	r := handler.NewRouter(authService, projectService)
	
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runDemo(svc *services.ProjectService) {
	ctx := context.Background()
	log.Println("--- Starting Demo Scenario ---")
	
	fakeUserID := uuid.New()
	
	log.Println("1. Creating Project Metadata...")
	proj, err := svc.CreateProject(ctx, fakeUserID, "Demo App", "nginx:alpine", "demo-site", 80)
	if err != nil {
		log.Printf("Error creating project (DB might be down): %v", err)
		return
	}
	log.Printf("Project Created: ID=%s Subdomain=%s", proj.ID, proj.Subdomain)

	// Deploy
	log.Println("2. Deploying to Docker...")
	deployment, err := svc.DeployProject(ctx, proj.ID)
	if err != nil {
		log.Fatalf("Deployment Failed: %v", err)
	}
	
	log.Printf("SUCCESS! Container ID: %s. Status: %s", deployment.ContainerID, deployment.Status)
	log.Println("Try checking 'docker ps'")
}
