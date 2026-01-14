package ports

import (
	"context"

	"github.com/damantine/multi-tenant-hosting/internal/core/domain"
	"github.com/google/uuid"
)

// ProjectRepository mendefinisikan operasi database untuk Project
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ContainerRuntime mendefinisikan interaksi dengan Docker Engine
// Ini adalah "Port" yang akan diimplementasikan oleh adapter Docker
type ContainerRuntime interface {
	// CreateContainer membuat container baru tanpa menjalankannya
	// Mengembalikan containerID jika sukses
	CreateContainer(ctx context.Context, config ContainerConfig) (string, error)

	// StartContainer menjalankan container yang sudah dibuat
	StartContainer(ctx context.Context, containerID string) error

	// StopContainer menghentikan container berjalan
	StopContainer(ctx context.Context, containerID string) error

	// RemoveContainer menghapus container
	RemoveContainer(ctx context.Context, containerID string) error
	
	// InspectContainer mendapatkan status terkini
	InspectContainer(ctx context.Context, containerID string) (*ContainerStatus, error)
}

// ContainerConfig structDTO untuk parameter pembuatan container
type ContainerConfig struct {
	Name      string
	Image     string
	Env       []string
	Labels    map[string]string
	Port      int
}

type ContainerStatus struct {
	ID     string
	State  string
	Status string
}
