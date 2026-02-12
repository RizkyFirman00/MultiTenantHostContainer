package services

import (
	"context"
	"fmt"
	"os"
	"regexp"

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
	// Ambil Base Domain dari environment variables (default: localhost)
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "localhost"
	}

	// Format Label Traefik v2/v3 untuk subdomain routing
	// "traefik.http.routers.my-app.rule=Host(`subdomain.domain.com`)"
	labels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.rule", project.Subdomain): fmt.Sprintf("Host(`%s.%s`)", project.Subdomain, baseDomain), 
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", project.Subdomain): fmt.Sprintf("%d", project.ContainerPort),
		"com.multitenant.project_id": project.ID.String(), // Label for robust cleanup
	}
	
	// Convert EnvVars domain ke []string format "KEY=VALUE"
	var envs []string
	for _, env := range project.EnvVars {
		envs = append(envs, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}

	// 2. Cleanup Existing Containers (if any, to avoid duplicates or name conflicts)
	// Iterate deployments and attempt to remove old containers.
	// In a production system, we might want to keep history, but for now we cleanup to keep resources tidy.
	for _, d := range project.Deployments {
		// Attempt stop (ignore error if not running)
		_ = s.dockerRuntime.StopContainer(ctx, d.ContainerID)
		// Attempt remove (ignore error if not found)
		_ = s.dockerRuntime.RemoveContainer(ctx, d.ContainerID)
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
	
	if err := s.repo.SaveDeployment(ctx, deployment); err != nil {
		// Log error but don't fail deployment completely? Or maybe fail?
		// Better to fail or warn. For now let's just log print (since we don't have logger injected)
		// and proceed to update project status.
		fmt.Printf("Failed to save deployment history: %v\n", err)
	}
	
	// Update status project
	project.Status = "running"
	// Also append to in-memory deployments to ensure immediate consistency if reused in same context (though context is usually per request)
	project.Deployments = append(project.Deployments, *deployment)
	
	s.repo.Update(ctx, project)

	return deployment, nil
}

// CreateProject hanya menyimpan metadata ke DB
func (s *ProjectService) CreateProject(ctx context.Context, userID uuid.UUID, name, image, subdomain string, port int) (*domain.Project, error) {
	// Validate Subdomain (RFC 1123 compliant for DNS labels)
	// Only lowercase alphanumeric and hyphens, cannot start or end with hyphen.
	var subdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !subdomainRegex.MatchString(subdomain) {
		return nil, fmt.Errorf("subdomain must consist of lowercase alphanumeric characters or hyphens, and cannot start or end with a hyphen")
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

	// Validate Subdomain (RFC 1123 compliant for DNS labels)
	// Only lowercase alphanumeric and hyphens, cannot start or end with hyphen.
	var subdomainRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	if !subdomainRegex.MatchString(subdomain) {
		return nil, fmt.Errorf("subdomain must consist of lowercase alphanumeric characters or hyphens, and cannot start or end with a hyphen")
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
			// Always attempt to stop and remove, regardless of stored status.
			// The container might be running even if DB says otherwise, or stopped but exists.
			_ = s.dockerRuntime.StopContainer(ctx, d.ContainerID)
			_ = s.dockerRuntime.RemoveContainer(ctx, d.ContainerID)
		}
	}
	
	// 3. Robust Cleanup: Find any wandering containers by label "com.multitenant.project_id"
	// This handles cases where deployments are missing or not synced.
	orphans, err := s.dockerRuntime.ListContainers(ctx, map[string]string{
		"com.multitenant.project_id": project.ID.String(),
	})
	if err == nil {
		for _, o := range orphans {
			_ = s.dockerRuntime.StopContainer(ctx, o.ID)
			_ = s.dockerRuntime.RemoveContainer(ctx, o.ID)
		}
	} else {
	    // Log error technically but proceed to delete project
	    fmt.Printf("Failed to list orphan containers: %v\n", err)
	}

	// 4. Remove from DB
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
	if err := s.dockerRuntime.StartContainer(ctx, latestDeployment.ContainerID); err != nil {
		return err
	}

	// Update project status
	project.Status = "running"
	return s.repo.Update(ctx, project)
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
	if err := s.dockerRuntime.StopContainer(ctx, latestDeployment.ContainerID); err != nil {
		return err
	}

	// Update project status
	project.Status = "stopped"
	return s.repo.Update(ctx, project)
}
