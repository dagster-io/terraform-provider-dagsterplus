package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// Deployment represents a Dagster+ deployment.
type Deployment struct {
	IntID     int64
	Name      string
	Type      string
	Status    string
	AgentType string
}

// CreateDeployment creates a new deployment in the organization. An empty
// agentType omits the agent type from the mutation, letting the organization's
// default apply.
func (c *Client) CreateDeployment(ctx context.Context, name, agentType string) (*Deployment, error) {
	var agentTypeArg *schema.DeploymentAgentType
	if agentType != "" {
		at := schema.DeploymentAgentType(agentType)
		agentTypeArg = &at
	}

	resp, err := schema.CreateDeployment(ctx, c.gqlClient(""), name, agentTypeArg)
	if err != nil {
		return nil, fmt.Errorf("CreateDeployment: %w", err)
	}

	switch r := resp.CreateDeployment.(type) {
	case *schema.CreateDeploymentCreateDeploymentDagsterCloudDeployment:
		return deploymentFromFields(&r.DeploymentFields), nil
	case *schema.CreateDeploymentCreateDeploymentDeploymentNotFoundError:
		return nil, fmt.Errorf("CreateDeployment: %s", r.Message)
	case *schema.CreateDeploymentCreateDeploymentUnauthorizedError:
		return nil, fmt.Errorf("CreateDeployment: %s", r.Message)
	case *schema.CreateDeploymentCreateDeploymentDuplicateDeploymentError:
		return nil, fmt.Errorf("CreateDeployment: %s", r.Message)
	case *schema.CreateDeploymentCreateDeploymentDeploymentLimitError:
		return nil, fmt.Errorf("CreateDeployment: %s", r.Message)
	case *schema.CreateDeploymentCreateDeploymentPythonError:
		return nil, fmt.Errorf("CreateDeployment: %s", r.Message)
	default:
		return nil, fmt.Errorf("CreateDeployment: unexpected result type %T", resp.CreateDeployment)
	}
}

// UpdateDeploymentAgentType switches an existing deployment between HYBRID and
// SERVERLESS, the equivalent of the "Switch to hybrid" action in the UI.
func (c *Client) UpdateDeploymentAgentType(ctx context.Context, name, agentType string) (*Deployment, error) {
	d, err := c.GetDeployment(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("UpdateDeploymentAgentType: %w", err)
	}

	resp, err := schema.UpdateDeploymentAgentType(ctx, c.gqlClient(""), int(d.IntID), schema.DeploymentAgentType(agentType))
	if err != nil {
		return nil, fmt.Errorf("UpdateDeploymentAgentType: %w", err)
	}

	switch r := resp.UpdateDeploymentAgentType.(type) {
	case *schema.UpdateDeploymentAgentTypeUpdateDeploymentAgentTypeDagsterCloudDeployment:
		return deploymentFromFields(&r.DeploymentFields), nil
	case *schema.UpdateDeploymentAgentTypeUpdateDeploymentAgentTypeDeploymentNotFoundError:
		return nil, fmt.Errorf("UpdateDeploymentAgentType: %s", r.Message)
	case *schema.UpdateDeploymentAgentTypeUpdateDeploymentAgentTypeUnauthorizedError:
		return nil, fmt.Errorf("UpdateDeploymentAgentType: %s", r.Message)
	case *schema.UpdateDeploymentAgentTypeUpdateDeploymentAgentTypePythonError:
		return nil, fmt.Errorf("UpdateDeploymentAgentType: %s", r.Message)
	default:
		return nil, fmt.Errorf("UpdateDeploymentAgentType: unexpected result type %T", resp.UpdateDeploymentAgentType)
	}
}

// deploymentFromFields converts the shared GraphQL fragment into the domain type.
func deploymentFromFields(f *schema.DeploymentFields) *Deployment {
	return &Deployment{
		IntID:     int64(f.DeploymentId),
		Name:      f.DeploymentName,
		Type:      string(f.DeploymentType),
		Status:    string(f.DeploymentStatus),
		AgentType: string(f.AgentType),
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
		result[i] = *deploymentFromFields(&d.DeploymentFields)
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
