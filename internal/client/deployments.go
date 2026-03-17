package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// Deployment represents a Dagster+ deployment.
type Deployment struct {
	IntID  int64
	Name   string
	Type   string
	Status string
}

// CreateDeployment creates a new serverless deployment in the organization.
func (c *Client) CreateDeployment(ctx context.Context, name string) (*Deployment, error) {
	resp, err := schema.CreateDeployment(ctx, c.gqlClient(""), name, schema.DeploymentAgentTypeServerless)
	if err != nil {
		return nil, fmt.Errorf("CreateDeployment: %w", err)
	}

	switch r := resp.CreateDeployment.(type) {
	case *schema.CreateDeploymentCreateDeploymentDagsterCloudDeployment:
		return &Deployment{
			IntID:  int64(r.DeploymentFields.DeploymentId),
			Name:   r.DeploymentFields.DeploymentName,
			Type:   string(r.DeploymentFields.DeploymentType),
			Status: string(r.DeploymentFields.DeploymentStatus),
		}, nil
	default:
		return nil, fmt.Errorf("CreateDeployment: unexpected result type %T", resp.CreateDeployment)
	}
}

// GetDeployment retrieves a deployment by name.
func (c *Client) GetDeployment(ctx context.Context, name string) (*Deployment, error) {
	deployments, err := c.ListDeployments(ctx)
	if err != nil {
		return nil, err
	}
	for i := range deployments {
		if deployments[i].Name == name {
			return &deployments[i], nil
		}
	}
	return nil, fmt.Errorf("GetDeployment: deployment %q not found", name)
}

// ListDeployments returns all deployments in the organization.
func (c *Client) ListDeployments(ctx context.Context) ([]Deployment, error) {
	resp, err := schema.ListDeployments(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListDeployments: %w", err)
	}

	result := make([]Deployment, len(resp.Deployments))
	for i, d := range resp.Deployments {
		result[i] = Deployment{
			IntID:  int64(d.DeploymentFields.DeploymentId),
			Name:   d.DeploymentFields.DeploymentName,
			Type:   string(d.DeploymentFields.DeploymentType),
			Status: string(d.DeploymentFields.DeploymentStatus),
		}
	}
	return result, nil
}

// GetDeploymentIntID returns the integer deployment ID for a given deployment name.
func (c *Client) GetDeploymentIntID(ctx context.Context, name string) (int64, error) {
	d, err := c.GetDeployment(ctx, name)
	if err != nil {
		return 0, err
	}
	return d.IntID, nil
}

// GetDeploymentNameByIntID returns the deployment name for a given integer deployment ID.
func (c *Client) GetDeploymentNameByIntID(ctx context.Context, id int64) (string, error) {
	deployments, err := c.ListDeployments(ctx)
	if err != nil {
		return "", err
	}
	for _, d := range deployments {
		if d.IntID == id {
			return d.Name, nil
		}
	}
	return "", fmt.Errorf("GetDeploymentNameByIntID: deployment with id %d not found", id)
}

// DeleteDeployment deletes a deployment by name.
func (c *Client) DeleteDeployment(ctx context.Context, name string) error {
	d, err := c.GetDeployment(ctx, name)
	if err != nil {
		return fmt.Errorf("DeleteDeployment: %w", err)
	}

	_, err = schema.DeleteDeployment(ctx, c.gqlClient(""), int(d.IntID))
	if err != nil {
		return fmt.Errorf("DeleteDeployment: %w", err)
	}
	return nil
}
