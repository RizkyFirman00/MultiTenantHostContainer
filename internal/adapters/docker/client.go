package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/damantine/multi-tenant-hosting/internal/core/ports"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient struct {
	cli *client.Client
}

// NewDockerClient inisialisasi koneksi ke Docker Daemon
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{cli: cli}, nil
}

// EnsureImage memastikan image tersedia (pull jika belum ada)
func (d *DockerClient) EnsureImage(ctx context.Context, imageName string) error {
	reader, err := d.cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	// Discard output agar tidak memenuhi buffer, di real app bisa di stream ke logs
	io.Copy(os.Stdout, reader) 
	return nil
}

// CreateContainer implementasi ports.ContainerRuntime
func (d *DockerClient) CreateContainer(ctx context.Context, config ports.ContainerConfig) (string, error) {
	// 1. Pastikan Image ada
	if err := d.EnsureImage(ctx, config.Image); err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}

	// 2. Konfigurasi Port Binding (Expose port container ke host dynamic port atau internal network)
	// Untuk kasus Traefik dan Single Node, biasanya kita tidak perlu bind ke Host Port jika dalam satu network.
	// Namun untuk debug, kita bisa set up variable.
	// Di sini kita asumsikan Traefik route via Docker Network, jadi tidak perlu publish ports ke Host (User -> Traefik -> Container IP).
	
	containerConfig := &container.Config{
		Image: config.Image,
		Env:   config.Env,
		Labels: config.Labels, // Traefik labels masuk sini
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", config.Port)): {},
		},
	}

	hostConfig := &container.HostConfig{
		NetworkMode: "traefik-net",
	}

	resp, err := d.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, config.Name)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	return d.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (d *DockerClient) StopContainer(ctx context.Context, containerID string) error {
	// Timeout default 10s
	return d.cli.ContainerStop(ctx, containerID, container.StopOptions{})
}

func (d *DockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	return d.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func (d *DockerClient) InspectContainer(ctx context.Context, containerID string) (*ports.ContainerStatus, error) {
	json, err := d.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, err
	}
	
	return &ports.ContainerStatus{
		ID: json.ID,
		State: json.State.Status, // running, paused, etc
		Status: json.State.Status,
	}, nil
}
