package client

import (
	"context"
	"fmt"
)

// Deployment represents a Dagster+ deployment.
type Deployment struct {
	Name string `json:"deploymentName"`
	Type string `json:"deploymentType"`
}

// CreateDeployment creates a new deployment in the organization.
func (c *Client) CreateDeployment(ctx context.Context, name, deploymentType string) (*Deployment, error) {
	const mutation = `
mutation CreateDeployment($deploymentName: String!, $deploymentAgentType: DeploymentAgentType!) {
  createDeployment(deploymentName: $deploymentName, deploymentAgentType: $deploymentAgentType) {
    __typename
    ... on DagsterCloudDeployment {
      deploymentName
      deploymentType
    }
  }
}`

	var result struct {
		CreateDeployment struct {
			Typename       string `json:"__typename"`
			DeploymentName string `json:"deploymentName"`
			DeploymentType string `json:"deploymentType"`
		} `json:"createDeployment"`
	}

	err := c.doGraphQL(ctx, "", mutation, map[string]any{
		"deploymentName":      name,
		"deploymentAgentType": deploymentType,
	}, &result)
	if err != nil {
		return nil, fmt.Errorf("CreateDeployment: %w", err)
	}

	if result.CreateDeployment.Typename != "DagsterCloudDeployment" {
		return nil, fmt.Errorf("CreateDeployment: unexpected result type %q (deployment may not have been created)", result.CreateDeployment.Typename)
	}

	return &Deployment{
		Name: result.CreateDeployment.DeploymentName,
		Type: result.CreateDeployment.DeploymentType,
	}, nil
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
	const query = `
query ListDeployments {
  deployments {
    deploymentName
    deploymentType
  }
}`

	var result struct {
		Deployments []Deployment `json:"deployments"`
	}

	if err := c.doGraphQL(ctx, "", query, nil, &result); err != nil {
		return nil, fmt.Errorf("ListDeployments: %w", err)
	}

	return result.Deployments, nil
}

// DeleteDeployment deletes a deployment by name.
func (c *Client) DeleteDeployment(ctx context.Context, name string) error {
	const mutation = `
mutation DeleteDeployment($deploymentName: String!) {
  deleteDeployment(deploymentName: $deploymentName) {
    success
  }
}`

	err := c.doGraphQL(ctx, "", mutation, map[string]any{
		"deploymentName": name,
	}, nil)
	if err != nil {
		return fmt.Errorf("DeleteDeployment: %w", err)
	}

	return nil
}
