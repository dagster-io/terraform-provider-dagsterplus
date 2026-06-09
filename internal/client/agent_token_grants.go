package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// AgentTokenGrant represents a permission grant for an agent token at a specific scope.
//
// Agent tokens only ever carry the AGENT grant, so (unlike user / service user
// grants) there is no grant level, custom role, or location grant to track here.
type AgentTokenGrant struct {
	AgentTokenID    string
	DeploymentScope string // "organization" | "deployment" | "all_branch_deployments" | "branch_deployments"
	DeploymentID    int64  // 0 for organization / all_branch_deployments scopes
}

func agentTokenGrantFromFields(agentTokenID string, deploymentScope schema.PermissionDeploymentScope, deploymentId int) *AgentTokenGrant {
	return &AgentTokenGrant{
		AgentTokenID:    agentTokenID,
		DeploymentScope: scopeForGrant(deploymentScope, deploymentId),
		DeploymentID:    int64(deploymentId),
	}
}

func findAgentTokenGrantInFields(f schema.AgentTokenGrantsFields, agentTokenID, deploymentScope string, deploymentID int64) *AgentTokenGrant {
	switch deploymentScope {
	case "organization":
		g := f.OrganizationPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return agentTokenGrantFromFields(agentTokenID, g.DeploymentScope, g.DeploymentId)
	case "all_branch_deployments":
		g := f.AllBranchDeploymentsPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return agentTokenGrantFromFields(agentTokenID, g.DeploymentScope, g.DeploymentId)
	case "branch_deployments":
		for _, g := range f.PerBaseBranchDeploymentsPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return agentTokenGrantFromFields(agentTokenID, g.DeploymentScope, g.DeploymentId)
			}
		}
	default:
		for _, g := range f.DeploymentPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return agentTokenGrantFromFields(agentTokenID, g.DeploymentScope, g.DeploymentId)
			}
		}
	}
	return nil
}

// SetAgentTokenGrant creates or updates a permission grant for an agent token.
func (c *Client) SetAgentTokenGrant(ctx context.Context, grant AgentTokenGrant) (*AgentTokenGrant, error) {
	intID, err := strconv.Atoi(grant.AgentTokenID)
	if err != nil {
		return nil, fmt.Errorf("SetAgentTokenGrant: invalid agent token id %q: %w", grant.AgentTokenID, err)
	}
	resp, err := schema.SetAgentGrant(ctx, c.gqlClient(""),
		intID,
		grantScopeToAPI[grant.DeploymentScope],
		int(grant.DeploymentID),
	)
	if err != nil {
		return nil, fmt.Errorf("SetAgentTokenGrant: %w", err)
	}

	switch r := resp.CreateOrUpdateAgentPermissions.(type) {
	case *schema.SetAgentGrantCreateOrUpdateAgentPermissionsDagsterCloudAgentToken:
		g := findAgentTokenGrantInFields(r.Permissions.AgentTokenGrantsFields, grant.AgentTokenID, grant.DeploymentScope, grant.DeploymentID)
		if g == nil {
			return nil, fmt.Errorf("SetAgentTokenGrant: grant not found in response")
		}
		return g, nil
	case *schema.SetAgentGrantCreateOrUpdateAgentPermissionsUnauthorizedError:
		return nil, fmt.Errorf("SetAgentTokenGrant: unauthorized; verify your API token has permission to edit agent tokens")
	default:
		return nil, fmt.Errorf("SetAgentTokenGrant: unexpected result type %T", resp.CreateOrUpdateAgentPermissions)
	}
}

// GetAgentTokenGrant retrieves a specific grant for an agent token by scope (and optional deployment ID).
func (c *Client) GetAgentTokenGrant(ctx context.Context, agentTokenID, deploymentScope string, deploymentID int64) (*AgentTokenGrant, error) {
	intID, err := strconv.Atoi(agentTokenID)
	if err != nil {
		return nil, fmt.Errorf("GetAgentTokenGrant: invalid agent token id %q: %w", agentTokenID, err)
	}
	resp, err := schema.GetAgentTokenGrants(ctx, c.gqlClient(""), intID)
	if err != nil {
		return nil, fmt.Errorf("GetAgentTokenGrant: %w", err)
	}

	switch r := resp.AgentTokenOrError.(type) {
	case *schema.GetAgentTokenGrantsAgentTokenOrErrorDagsterCloudAgentToken:
		g := findAgentTokenGrantInFields(r.Permissions.AgentTokenGrantsFields, agentTokenID, deploymentScope, deploymentID)
		if g == nil {
			return nil, fmt.Errorf("GetAgentTokenGrant: grant for agent token %q scope %q not found", agentTokenID, deploymentScope)
		}
		return g, nil
	default:
		return nil, fmt.Errorf("GetAgentTokenGrant: agent token %q not found", agentTokenID)
	}
}

// ListAgentTokenDeploymentGrants returns all deployment-scoped grants for an agent token.
func (c *Client) ListAgentTokenDeploymentGrants(ctx context.Context, agentTokenID string) ([]AgentTokenGrant, error) {
	intID, err := strconv.Atoi(agentTokenID)
	if err != nil {
		return nil, fmt.Errorf("ListAgentTokenDeploymentGrants: invalid agent token id %q: %w", agentTokenID, err)
	}
	resp, err := schema.GetAgentTokenGrants(ctx, c.gqlClient(""), intID)
	if err != nil {
		return nil, fmt.Errorf("ListAgentTokenDeploymentGrants: %w", err)
	}
	switch r := resp.AgentTokenOrError.(type) {
	case *schema.GetAgentTokenGrantsAgentTokenOrErrorDagsterCloudAgentToken:
		var grants []AgentTokenGrant
		for _, g := range r.Permissions.AgentTokenGrantsFields.DeploymentPermissionGrants {
			grants = append(grants, *agentTokenGrantFromFields(agentTokenID, g.DeploymentScope, g.DeploymentId))
		}
		return grants, nil
	default:
		return nil, fmt.Errorf("ListAgentTokenDeploymentGrants: agent token %q not found", agentTokenID)
	}
}

// ListAgentTokenBranchDeploymentsGrants returns all per-base-branch grants for an agent token.
func (c *Client) ListAgentTokenBranchDeploymentsGrants(ctx context.Context, agentTokenID string) ([]AgentTokenGrant, error) {
	intID, err := strconv.Atoi(agentTokenID)
	if err != nil {
		return nil, fmt.Errorf("ListAgentTokenBranchDeploymentsGrants: invalid agent token id %q: %w", agentTokenID, err)
	}
	resp, err := schema.GetAgentTokenGrants(ctx, c.gqlClient(""), intID)
	if err != nil {
		return nil, fmt.Errorf("ListAgentTokenBranchDeploymentsGrants: %w", err)
	}
	switch r := resp.AgentTokenOrError.(type) {
	case *schema.GetAgentTokenGrantsAgentTokenOrErrorDagsterCloudAgentToken:
		var grants []AgentTokenGrant
		for _, g := range r.Permissions.AgentTokenGrantsFields.PerBaseBranchDeploymentsPermissionGrants {
			grants = append(grants, *agentTokenGrantFromFields(agentTokenID, g.DeploymentScope, g.DeploymentId))
		}
		return grants, nil
	default:
		return nil, fmt.Errorf("ListAgentTokenBranchDeploymentsGrants: agent token %q not found", agentTokenID)
	}
}

// DeleteAgentTokenGrant removes a permission grant for an agent token at the given scope.
func (c *Client) DeleteAgentTokenGrant(ctx context.Context, agentTokenID, deploymentScope string, deploymentID int64) error {
	intID, err := strconv.Atoi(agentTokenID)
	if err != nil {
		return fmt.Errorf("DeleteAgentTokenGrant: invalid agent token id %q: %w", agentTokenID, err)
	}
	_, err = schema.DeleteAgentGrant(ctx, c.gqlClient(""),
		intID,
		grantScopeToAPI[deploymentScope],
		int(deploymentID),
	)
	if err != nil {
		return fmt.Errorf("DeleteAgentTokenGrant: %w", err)
	}
	return nil
}
