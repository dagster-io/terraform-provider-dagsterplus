package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// LocationGrant represents a permission grant scoped to a specific code location.
type LocationGrant struct {
	LocationName string
	Grant        string
	CustomRoleID string
}

// TeamGrant represents a permission grant for a team at a specific scope.
type TeamGrant struct {
	TeamID          string
	DeploymentScope string // "organization" | "deployment"
	DeploymentID    int64  // 0 for org scope
	Grant           string // "VIEWER" | "EDITOR" | "LAUNCHER" | "ADMIN" | "CUSTOM"
	CustomRoleID    string // only when Grant == "CUSTOM"
	LocationGrants  []LocationGrant
}

var grantScopeToAPI = map[string]schema.PermissionDeploymentScope{
	"organization":           "ORGANIZATION",
	"deployment":             "DEPLOYMENT",
	"all_branch_deployments": "ALL_BRANCH_DEPLOYMENTS",
	"branch_deployments":     "ALL_BRANCH_DEPLOYMENTS",
}

// scopeForGrant maps an API enum + deploymentId back to an internal scope string.
// ALL_BRANCH_DEPLOYMENTS is ambiguous on the wire: when deploymentId is non-zero,
// it represents a per-base-branch grant under that parent deployment.
func scopeForGrant(deploymentScope schema.PermissionDeploymentScope, deploymentId int) string {
	switch deploymentScope {
	case "ORGANIZATION":
		return "organization"
	case "DEPLOYMENT":
		return "deployment"
	case "ALL_BRANCH_DEPLOYMENTS":
		if deploymentId != 0 {
			return "branch_deployments"
		}
		return "all_branch_deployments"
	}
	return string(deploymentScope)
}

func teamGrantFromFields(teamID string, grant schema.PermissionGrant, customRoleId string, deploymentScope schema.PermissionDeploymentScope, deploymentId int, locationGrants []schema.TeamPermissionsFieldsDeploymentPermissionGrantsDagsterCloudScopedPermissionGrantLocationGrantsLocationScopedGrant) *TeamGrant {
	scope := scopeForGrant(deploymentScope, deploymentId)
	lgs := make([]LocationGrant, len(locationGrants))
	for i, lg := range locationGrants {
		lgs[i] = LocationGrant{
			LocationName: lg.LocationName,
			Grant:        string(lg.Grant),
			CustomRoleID: lg.CustomRoleId,
		}
	}
	return &TeamGrant{
		TeamID:          teamID,
		DeploymentScope: scope,
		DeploymentID:    int64(deploymentId),
		Grant:           string(grant),
		CustomRoleID:    customRoleId,
		LocationGrants:  lgs,
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
		return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
	case "all_branch_deployments":
		g := f.AllBranchDeploymentsPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
	case "branch_deployments":
		for _, g := range f.PerBaseBranchDeploymentsPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
			}
		}
	default:
		for _, g := range f.DeploymentPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, g.LocationGrants)
			}
		}
	}
	return nil
}

// SetTeamGrant creates or updates a permission grant for a team.
func (c *Client) SetTeamGrant(ctx context.Context, grant TeamGrant) (*TeamGrant, error) {
	locationGrants := make([]schema.LocationScopedGrantInput, len(grant.LocationGrants))
	for i, lg := range grant.LocationGrants {
		locationGrants[i] = schema.LocationScopedGrantInput{
			LocationName: lg.LocationName,
			Grant:        schema.PermissionGrant(lg.Grant),
			CustomRoleId: lg.CustomRoleID,
		}
	}
	resp, err := schema.SetTeamGrant(ctx, c.gqlClient(""),
		grant.TeamID,
		grantScopeToAPI[grant.DeploymentScope],
		schema.PermissionGrant(grant.Grant),
		grant.CustomRoleID,
		int(grant.DeploymentID),
		locationGrants,
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

// ListTeamDeploymentGrants returns all deployment-scoped grants for a given team ID.
func (c *Client) ListTeamDeploymentGrants(ctx context.Context, teamID string) ([]TeamGrant, error) {
	resp, err := schema.ListTeamGrants(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListTeamDeploymentGrants: %w", err)
	}

	for _, tp := range resp.TeamPermissions {
		if tp.TeamPermissionsFields.Team.Id != teamID {
			continue
		}
		var grants []TeamGrant
		for _, g := range tp.TeamPermissionsFields.DeploymentPermissionGrants {
			grants = append(grants, *teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, g.LocationGrants))
		}
		return grants, nil
	}
	return nil, nil
}

// ListTeamBranchDeploymentsGrants returns all per-base-branch grants for a given team ID.
func (c *Client) ListTeamBranchDeploymentsGrants(ctx context.Context, teamID string) ([]TeamGrant, error) {
	resp, err := schema.ListTeamGrants(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListTeamBranchDeploymentsGrants: %w", err)
	}

	for _, tp := range resp.TeamPermissions {
		if tp.TeamPermissionsFields.Team.Id != teamID {
			continue
		}
		var grants []TeamGrant
		for _, g := range tp.TeamPermissionsFields.PerBaseBranchDeploymentsPermissionGrants {
			grants = append(grants, *teamGrantFromFields(teamID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil))
		}
		return grants, nil
	}
	return nil, nil
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
