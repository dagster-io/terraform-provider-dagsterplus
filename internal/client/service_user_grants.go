package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// ServiceUserGrant represents a permission grant for a service user at a specific scope.
type ServiceUserGrant struct {
	ServiceUserID   string
	DeploymentScope string // "organization" | "deployment" | "all_branch_deployments"
	DeploymentID    int64  // 0 for org scope
	Grant           string // "VIEWER" | "EDITOR" | "LAUNCHER" | "ADMIN" | "CUSTOM"
	CustomRoleID    string // only when Grant == "CUSTOM"
	LocationGrants  []LocationGrant
}

func serviceUserGrantFromFields(serviceUserID string, grant schema.PermissionGrant, customRoleId string, deploymentScope schema.PermissionDeploymentScope, deploymentId int, locationGrants []schema.ServiceUserGrantsFieldsDeploymentPermissionGrantsDagsterCloudScopedPermissionGrantLocationGrantsLocationScopedGrant) *ServiceUserGrant {
	scope := apiToGrantScope[deploymentScope]
	if scope == "" {
		scope = string(deploymentScope)
	}
	lgs := make([]LocationGrant, len(locationGrants))
	for i, lg := range locationGrants {
		lgs[i] = LocationGrant{
			LocationName: lg.LocationName,
			Grant:        string(lg.Grant),
			CustomRoleID: lg.CustomRoleId,
		}
	}
	return &ServiceUserGrant{
		ServiceUserID:   serviceUserID,
		DeploymentScope: scope,
		DeploymentID:    int64(deploymentId),
		Grant:           string(grant),
		CustomRoleID:    customRoleId,
		LocationGrants:  lgs,
	}
}

func findServiceUserGrantInFields(f schema.ServiceUserGrantsFields, serviceUserID, deploymentScope string, deploymentID int64) *ServiceUserGrant {
	switch deploymentScope {
	case "organization":
		g := f.OrganizationPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return serviceUserGrantFromFields(serviceUserID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
	case "all_branch_deployments":
		g := f.AllBranchDeploymentsPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return serviceUserGrantFromFields(serviceUserID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
	default:
		for _, g := range f.DeploymentPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return serviceUserGrantFromFields(serviceUserID, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, g.LocationGrants)
			}
		}
	}
	return nil
}

// SetServiceUserGrant creates or updates a permission grant for a service user.
func (c *Client) SetServiceUserGrant(ctx context.Context, grant ServiceUserGrant) (*ServiceUserGrant, error) {
	locationGrants := make([]schema.LocationScopedGrantInput, len(grant.LocationGrants))
	for i, lg := range grant.LocationGrants {
		locationGrants[i] = schema.LocationScopedGrantInput{
			LocationName: lg.LocationName,
			Grant:        schema.PermissionGrant(lg.Grant),
			CustomRoleId: lg.CustomRoleID,
		}
	}
	resp, err := schema.SetServiceUserGrant(ctx, c.gqlClient(""),
		grant.ServiceUserID,
		grantScopeToAPI[grant.DeploymentScope],
		schema.PermissionGrant(grant.Grant),
		grant.CustomRoleID,
		int(grant.DeploymentID),
		locationGrants,
	)
	if err != nil {
		return nil, fmt.Errorf("SetServiceUserGrant: %w", err)
	}

	switch r := resp.CreateOrUpdateServiceUserPermissions.(type) {
	case *schema.SetServiceUserGrantCreateOrUpdateServiceUserPermissionsServiceUserWithScopedPermissionGrants:
		g := findServiceUserGrantInFields(r.ServiceUserGrantsFields, grant.ServiceUserID, grant.DeploymentScope, grant.DeploymentID)
		if g == nil {
			return nil, fmt.Errorf("SetServiceUserGrant: grant not found in response")
		}
		return g, nil
	default:
		return nil, fmt.Errorf("SetServiceUserGrant: unexpected result type %T", resp.CreateOrUpdateServiceUserPermissions)
	}
}

// GetServiceUserGrant retrieves a specific grant for a service user by scope (and optional deployment ID).
func (c *Client) GetServiceUserGrant(ctx context.Context, serviceUserID, deploymentScope string, deploymentID int64) (*ServiceUserGrant, error) {
	resp, err := schema.GetServiceUserGrants(ctx, c.gqlClient(""), serviceUserID)
	if err != nil {
		return nil, fmt.Errorf("GetServiceUserGrant: %w", err)
	}

	switch r := resp.ServiceUser.(type) {
	case *schema.GetServiceUserGrantsServiceUserServiceUserWithScopedPermissionGrants:
		g := findServiceUserGrantInFields(r.ServiceUserGrantsFields, serviceUserID, deploymentScope, deploymentID)
		if g == nil {
			return nil, fmt.Errorf("GetServiceUserGrant: grant for service user %q scope %q not found", serviceUserID, deploymentScope)
		}
		return g, nil
	default:
		return nil, fmt.Errorf("GetServiceUserGrant: service user %q not found", serviceUserID)
	}
}

// DeleteServiceUserGrant removes a permission grant for a service user at the given scope.
func (c *Client) DeleteServiceUserGrant(ctx context.Context, serviceUserID, deploymentScope string, deploymentID int64) error {
	_, err := schema.DeleteServiceUserGrant(ctx, c.gqlClient(""),
		serviceUserID,
		grantScopeToAPI[deploymentScope],
		int(deploymentID),
	)
	if err != nil {
		return fmt.Errorf("DeleteServiceUserGrant: %w", err)
	}
	return nil
}
