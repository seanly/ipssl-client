package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"ipssl-client/internal/logger"
)

// Client represents a Docker API client
type Client struct {
	client *client.Client
	logger *logger.Logger
}

// NewClient creates a new Docker client
func NewClient(logger *logger.Logger) (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &Client{
		client: cli,
		logger: logger,
	}, nil
}

// ReloadContainer reloads a Docker container by sending a SIGHUP signal
func (c *Client) ReloadContainer(ctx context.Context, containerName string) error {
	c.logger.Info("Reloading container", "container", containerName)

	// Get container information
	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var targetContainer types.Container
	found := false
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName || name == containerName {
				targetContainer = container
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("container %s not found", containerName)
	}

	// Check if container is running
	if targetContainer.State != "running" {
		return fmt.Errorf("container %s is not running (state: %s)", containerName, targetContainer.State)
	}

	// Send SIGHUP signal to reload configuration
	err = c.client.ContainerKill(ctx, targetContainer.ID, "SIGHUP")
	if err != nil {
		return fmt.Errorf("failed to send SIGHUP signal to container %s: %w", containerName, err)
	}

	c.logger.Info("Successfully sent reload signal to container", "container", containerName)
	return nil
}

// RestartContainer restarts a Docker container
func (c *Client) RestartContainer(ctx context.Context, containerName string) error {
	c.logger.Info("Restarting container", "container", containerName)

	// Get container information
	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	var targetContainer types.Container
	found := false
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName || name == containerName {
				targetContainer = container
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("container %s not found", containerName)
	}

	// Restart the container
	timeout := 30
	err = c.client.ContainerRestart(ctx, targetContainer.ID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to restart container %s: %w", containerName, err)
	}

	c.logger.Info("Successfully restarted container", "container", containerName)
	return nil
}

// GetContainerStatus gets the status of a Docker container
func (c *Client) GetContainerStatus(ctx context.Context, containerName string) (string, error) {
	containers, err := c.client.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName || name == containerName {
				return container.State, nil
			}
		}
	}

	return "", fmt.Errorf("container %s not found", containerName)
}
