package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// TeamGrant represents a permission grant for a team at a specific scope.
type TeamGrant struct {
	TeamID          string
	DeploymentScope string // "organization" | "deployment"
	DeploymentID    int64  // 0 for org scope
	Grant           string // "VIEWER" | "EDITOR" | "LAUNCHER" | "ADMIN" | "CUSTOM"
	CustomRoleID    string // only when Grant == "CUSTOM"
}

var grantScopeToAPI = map[string]schema.PermissionDeploymentScope{
	"organization":           "ORGANIZATION",
	"deployment":             "DEPLOYMENT",
	"all_branch_deployments": "ALL_BRANCH_DEPLOYMENTS",
}

var apiToGrantScope = map[schema.PermissionDeploymentScope]string{
	"ORGANIZATION":           "organization",
	"DEPLOYMENT":             "deployment",
	"ALL_BRANCH_DEPLOYMENTS": "all_branch_deployments",
}

func teamGrantFromFields(teamID string, grant schema.PermissionGrant, customRoleId string, deploymentScope schema.PermissionDeploymentScope, deploymentId int) *TeamGrant {
	scope := apiToGrantScope[deploymentScope]
	if scope == "" {
		scope = string(deploymentScope)
	}
	return &TeamGrant{
		TeamID:          teamID,
		DeploymentScope: scope,
		DeploymentID:    int64(deploymentId),
		Grant:           string(grant),
		CustomRoleID:    customRoleId,
	}
}

func findGrantInFields(f schema.TeamPermissionsFields, deploymentScope string, deploymentID int64) *TeamGrant {
	teamID := f.Team.Id
	switch deploymentScope {
	case "organization":
		g := f.OrganizationPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId)
	case "all_branch_deployments":
		g := f.AllBranchDeploymentsPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId)
	default:
		for _, g := range f.DeploymentPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId)
			}
		}
	}
	return nil
}

// SetTeamGrant creates or updates a permission grant for a team.
func (c *Client) SetTeamGrant(ctx context.Context, grant TeamGrant) (*TeamGrant, error) {
	resp, err := schema.SetTeamGrant(ctx, c.gqlClient(""),
		grant.TeamID,
		grantScopeToAPI[grant.DeploymentScope],
		schema.PermissionGrant(grant.Grant),
		grant.CustomRoleID,
		int(grant.DeploymentID),
	)
	if err != nil {
		return nil, fmt.Errorf("SetTeamGrant: %w", err)
	}

	switch r := resp.CreateOrUpdateTeamPermission.(type) {
	case *schema.SetTeamGrantCreateOrUpdateTeamPermissionCreateOrUpdateTeamPermissionSuccess:
		g := findGrantInFields(r.TeamPermissions.TeamPermissionsFields, grant.DeploymentScope, grant.DeploymentID)
		if g == nil {
			return nil, fmt.Errorf("SetTeamGrant: grant not found in response")
		}
		return g, nil
	default:
		return nil, fmt.Errorf("SetTeamGrant: unexpected result type %T", resp.CreateOrUpdateTeamPermission)
	}
}

// GetTeamGrant retrieves a specific grant for a team by scope (and optional deployment ID).
func (c *Client) GetTeamGrant(ctx context.Context, teamID, deploymentScope string, deploymentID int64) (*TeamGrant, error) {
	resp, err := schema.ListTeamGrants(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("GetTeamGrant: %w", err)
	}

	for _, tp := range resp.TeamPermissions {
		if tp.TeamPermissionsFields.Team.Id == teamID {
			g := findGrantInFields(tp.TeamPermissionsFields, deploymentScope, deploymentID)
			if g != nil {
				return g, nil
			}
		}
	}
	return nil, fmt.Errorf("GetTeamGrant: grant for team %q scope %q not found", teamID, deploymentScope)
}

// DeleteTeamGrant removes a permission grant for a team at the given scope.
func (c *Client) DeleteTeamGrant(ctx context.Context, teamID, deploymentScope string, deploymentID int64) error {
	_, err := schema.DeleteTeamGrant(ctx, c.gqlClient(""),
		teamID,
		grantScopeToAPI[deploymentScope],
		int(deploymentID),
	)
	if err != nil {
		return fmt.Errorf("DeleteTeamGrant: %w", err)
	}
	return nil
}
