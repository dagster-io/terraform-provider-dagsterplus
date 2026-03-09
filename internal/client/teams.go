package client

import (
	"context"
	"fmt"
)

// Team represents a Dagster+ team.
type Team struct {
	ID   string `json:"teamId"`
	Name string `json:"teamName"`
}

// Permission represents a deployment-level permission for a team.
type Permission struct {
	DeploymentName string `json:"deploymentName"`
	Role           string `json:"deploymentRole"`
}

// CreateTeam creates a new team in the organization.
func (c *Client) CreateTeam(ctx context.Context, name string) (*Team, error) {
	const mutation = `
mutation CreateTeam($teamName: String!) {
  createOrganizationMemberTeam(teamName: $teamName) {
    team {
      teamId
      teamName
    }
  }
}`

	var result struct {
		CreateOrganizationMemberTeam struct {
			Team *Team `json:"team"`
		} `json:"createOrganizationMemberTeam"`
	}

	err := c.doGraphQL(ctx, "", mutation, map[string]any{
		"teamName": name,
	}, &result)
	if err != nil {
		return nil, fmt.Errorf("CreateTeam: %w", err)
	}

	if result.CreateOrganizationMemberTeam.Team == nil {
		return nil, fmt.Errorf("CreateTeam: API returned nil team")
	}

	return result.CreateOrganizationMemberTeam.Team, nil
}

// GetTeam retrieves a team by ID.
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	teams, err := c.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	for i := range teams {
		if teams[i].ID == id {
			return &teams[i], nil
		}
	}
	return nil, fmt.Errorf("GetTeam: team %q not found", id)
}

// ListTeams returns all teams in the organization.
func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	const query = `
query ListTeams {
  organizationMemberTeams {
    teamId
    teamName
  }
}`

	var result struct {
		OrganizationMemberTeams []Team `json:"organizationMemberTeams"`
	}

	if err := c.doGraphQL(ctx, "", query, nil, &result); err != nil {
		return nil, fmt.Errorf("ListTeams: %w", err)
	}

	return result.OrganizationMemberTeams, nil
}

// UpdateTeamPermissions sets deployment-level permissions for a team.
func (c *Client) UpdateTeamPermissions(ctx context.Context, teamID string, perms []Permission) error {
	const mutation = `
mutation UpdateTeamPermissions($teamId: String!, $deploymentPermissions: [DeploymentPermissionInput!]!) {
  updateOrganizationMemberTeamPermissions(
    teamId: $teamId
    deploymentPermissions: $deploymentPermissions
  ) {
    success
  }
}`

	type deploymentPermInput struct {
		DeploymentName string `json:"deploymentName"`
		DeploymentRole string `json:"deploymentRole"`
	}

	inputs := make([]deploymentPermInput, len(perms))
	for i, p := range perms {
		inputs[i] = deploymentPermInput{
			DeploymentName: p.DeploymentName,
			DeploymentRole: p.Role,
		}
	}

	err := c.doGraphQL(ctx, "", mutation, map[string]any{
		"teamId":                teamID,
		"deploymentPermissions": inputs,
	}, nil)
	if err != nil {
		return fmt.Errorf("UpdateTeamPermissions: %w", err)
	}

	return nil
}

// GetTeamPermissions retrieves the deployment permissions for a team.
func (c *Client) GetTeamPermissions(ctx context.Context, teamID string) ([]Permission, error) {
	const query = `
query GetTeamPermissions($teamId: String!) {
  organizationMemberTeam(teamId: $teamId) {
    deploymentPermissions {
      deploymentName
      deploymentRole
    }
  }
}`

	var result struct {
		OrganizationMemberTeam struct {
			DeploymentPermissions []Permission `json:"deploymentPermissions"`
		} `json:"organizationMemberTeam"`
	}

	if err := c.doGraphQL(ctx, "", query, map[string]any{"teamId": teamID}, &result); err != nil {
		return nil, fmt.Errorf("GetTeamPermissions: %w", err)
	}

	return result.OrganizationMemberTeam.DeploymentPermissions, nil
}

// DeleteTeam removes a team from the organization.
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	const mutation = `
mutation DeleteTeam($teamId: String!) {
  deleteOrganizationMemberTeam(teamId: $teamId) {
    success
  }
}`

	err := c.doGraphQL(ctx, "", mutation, map[string]any{
		"teamId": id,
	}, nil)
	if err != nil {
		return fmt.Errorf("DeleteTeam: %w", err)
	}

	return nil
}
