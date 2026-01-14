package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/damantine/multi-tenant-hosting/internal/core/domain"
	"github.com/damantine/multi-tenant-hosting/internal/core/ports"
	"github.com/google/uuid"
)

type ProjectService struct {
	repo          ports.ProjectRepository
	dockerRuntime ports.ContainerRuntime
}

func NewProjectService(repo ports.ProjectRepository, docker ports.ContainerRuntime) *ProjectService {
	return &ProjectService{
		repo:          repo,
		dockerRuntime: docker,
	}
}

// DeployProject menghandle logika deployment aplikasi user
func (s *ProjectService) DeployProject(ctx context.Context, projectID uuid.UUID) (*domain.Deployment, error) {
	// 1. Ambil data project
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// 2. Siapkan config container
	// Format Label Traefik v2/v3 untuk subdomain routing
	// "traefik.http.routers.my-app.rule=Host(`subdomain.domain.com`)"
	labels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.rule", project.Subdomain): fmt.Sprintf("Host(`%s.localhost`)", project.Subdomain), // Pakai localhost utk dev
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", project.Subdomain): fmt.Sprintf("%d", project.ContainerPort),
	}
	
	// Convert EnvVars domain ke []string format "KEY=VALUE"
	var envs []string
	for _, env := range project.EnvVars {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}

	config := ports.ContainerConfig{
		Name:   fmt.Sprintf("%s-%s", project.Subdomain, uuid.NewString()[:8]), // Uniq name
		Image:  project.ImageName,
		Env:    envs,
		Labels: labels,
		Port:   project.ContainerPort,
	}

	// 3. Panggil Docker Adapter
	containerID, err := s.dockerRuntime.CreateContainer(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("docker create failed: %w", err)
	}

	if err := s.dockerRuntime.StartContainer(ctx, containerID); err != nil {
		return nil, fmt.Errorf("docker start failed: %w", err)
	}

	// 4. Record deployment history
	deployment := &domain.Deployment{
		ProjectID:   project.ID,
		ContainerID: containerID,
		Status:      "running",
	}
	// Di real app, simpan deployment ke DB via repo (belum diimplementasi di interface repo contoh ini)
	// s.repo.SaveDeployment(deployment)
	
	// Update status project
	project.Status = "running"
	s.repo.Update(ctx, project)

	return deployment, nil
}

// CreateProject hanya menyimpan metadata ke DB
func (s *ProjectService) CreateProject(ctx context.Context, userID uuid.UUID, name, image, subdomain string, port int) (*domain.Project, error) {
    if strings.Contains(subdomain, " ") {
        return nil, fmt.Errorf("subdomain cannot contain spaces")
    }

	project := &domain.Project{
		UserID:        userID,
		Name:          name,
		ImageName:     image,
		Subdomain:     subdomain,
		ContainerPort: port,
		Status:        "created",
	}

	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, userID uuid.UUID) ([]domain.Project, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ProjectService) GetProject(ctx context.Context, projectID uuid.UUID) (*domain.Project, error) {
	return s.repo.GetByID(ctx, projectID)
}

func (s *ProjectService) UpdateProject(ctx context.Context, projectID uuid.UUID, name, image, subdomain string, port int) (*domain.Project, error) {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Update fields
	project.Name = name
	project.ImageName = image
	project.Subdomain = subdomain
	project.ContainerPort = port
	
	// Reset status if critical config builds changes (optional, but good practice)
	// For now we keep it simple.

	if err := s.repo.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// 1. Remove Container if exists
	// We need to find container ID. Usually stored in Deployments.
	// For simplicity, we check deployments or just try to remove by name if we knew it?
	// Better: Check latest deployment or loop through deployments.
	// In this simple version, let's assume we try to cleanup resources based on potential container names or just skip if complex.
	// Actually, we should check active deployment.
    // Let's use List to find deployments if not loaded.
    // Repo GetByID loads deployments.
    
    if len(project.Deployments) > 0 {
		for _, d := range project.Deployments {
			if d.Status == "running" {
				// Try to stop and remove
				_ = s.dockerRuntime.StopContainer(ctx, d.ContainerID)
				_ = s.dockerRuntime.RemoveContainer(ctx, d.ContainerID)
			}
		}
	} else {
		// Fallback cleanup try (best effort)
		// Try to find container by name? Not implemented in runtime yet.
		// Skip for now.
	}

	// 2. Remove from DB
	return s.repo.Delete(ctx, projectID)
}

func (s *ProjectService) StartProject(ctx context.Context, projectID uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Find running or stopped container
	// For simplicity, we assume the latest deployment contains the relevant container ID
	// Real implementation might need to handle multiple deployments or look up by name.
	if len(project.Deployments) == 0 {
		return fmt.Errorf("no deployments found for this project")
	}
	
	// Get latest deployment
	latestDeployment := project.Deployments[len(project.Deployments)-1]
	
	// Start container
	return s.dockerRuntime.StartContainer(ctx, latestDeployment.ContainerID)
}

func (s *ProjectService) StopProject(ctx context.Context, projectID uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	if len(project.Deployments) == 0 {
		return fmt.Errorf("no deployments found for this project")
	}

	latestDeployment := project.Deployments[len(project.Deployments)-1]
	return s.dockerRuntime.StopContainer(ctx, latestDeployment.ContainerID)
}
