package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// UserGrant represents a permission grant for a regular (non-service) user at a specific scope.
type UserGrant struct {
	UserID          string // the Dagster+ user id (stable)
	Email           string // the email used by the underlying GraphQL mutations
	DeploymentScope string // "organization" | "deployment" | "all_branch_deployments" | "branch_deployments"
	DeploymentID    int64  // 0 for org and global all_branch scope; parent deployment id for branch_deployments; target deployment id for deployment scope
	Grant           string
	CustomRoleID    string
	LocationGrants  []LocationGrant
}

func userGrantFromFields(userID, email string, grant schema.PermissionGrant, customRoleId string, deploymentScope schema.PermissionDeploymentScope, deploymentId int, locationGrants []schema.UserWithGrantsFieldsDeploymentPermissionGrantsDagsterCloudScopedPermissionGrantLocationGrantsLocationScopedGrant) *UserGrant {
	scope := scopeForGrant(deploymentScope, deploymentId)
	lgs := make([]LocationGrant, len(locationGrants))
	for i, lg := range locationGrants {
		lgs[i] = LocationGrant{
			LocationName: lg.LocationName,
			Grant:        string(lg.Grant),
			CustomRoleID: lg.CustomRoleId,
		}
	}
	return &UserGrant{
		UserID:          userID,
		Email:           email,
		DeploymentScope: scope,
		DeploymentID:    int64(deploymentId),
		Grant:           string(grant),
		CustomRoleID:    customRoleId,
		LocationGrants:  lgs,
	}
}

func findUserGrantInFields(f schema.UserWithGrantsFields, deploymentScope string, deploymentID int64) *UserGrant {
	userID := fmt.Sprintf("%d", f.User.UserId)
	email := f.User.Email
	switch deploymentScope {
	case "organization":
		g := f.OrganizationPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return userGrantFromFields(userID, email, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
	case "all_branch_deployments":
		g := f.AllBranchDeploymentsPermissionGrant
		if g.Grant == "" {
			return nil
		}
		return userGrantFromFields(userID, email, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
	case "branch_deployments":
		for _, g := range f.PerBaseBranchDeploymentsPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return userGrantFromFields(userID, email, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil)
			}
		}
	default:
		for _, g := range f.DeploymentPermissionGrants {
			if int64(g.DeploymentId) == deploymentID {
				return userGrantFromFields(userID, email, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, g.LocationGrants)
			}
		}
	}
	return nil
}

// SetUserGrant creates or updates a permission grant for a user. The user must
// already be a member of the organization (the API identifies them by email).
func (c *Client) SetUserGrant(ctx context.Context, grant UserGrant) (*UserGrant, error) {
	locationGrants := make([]schema.LocationScopedGrantInput, len(grant.LocationGrants))
	for i, lg := range grant.LocationGrants {
		locationGrants[i] = schema.LocationScopedGrantInput{
			LocationName: lg.LocationName,
			Grant:        schema.PermissionGrant(lg.Grant),
			CustomRoleId: lg.CustomRoleID,
		}
	}
	resp, err := schema.SetUserGrant(ctx, c.gqlClient(""),
		grant.Email,
		grantScopeToAPI[grant.DeploymentScope],
		schema.PermissionGrant(grant.Grant),
		grant.CustomRoleID,
		int(grant.DeploymentID),
		locationGrants,
	)
	if err != nil {
		return nil, fmt.Errorf("SetUserGrant: %w", err)
	}

	switch r := resp.CreateOrUpdateUserPermissions.(type) {
	case *schema.SetUserGrantCreateOrUpdateUserPermissionsDagsterCloudUserWithScopedPermissionGrants:
		g := findUserGrantInFields(r.UserWithGrantsFields, grant.DeploymentScope, grant.DeploymentID)
		if g == nil {
			return nil, fmt.Errorf("SetUserGrant: grant not found in response")
		}
		return g, nil
	case *schema.SetUserGrantCreateOrUpdateUserPermissionsUserNotFoundError:
		return nil, fmt.Errorf("SetUserGrant: %s", r.Message)
	case *schema.SetUserGrantCreateOrUpdateUserPermissionsCantRemoveAllAdminsError:
		return nil, fmt.Errorf("SetUserGrant: %s", r.Message)
	default:
		return nil, fmt.Errorf("SetUserGrant: unexpected result type %T", resp.CreateOrUpdateUserPermissions)
	}
}

// GetUserGrant retrieves a specific grant for a user by scope.
// Identifies the user by their Dagster+ user id.
func (c *Client) GetUserGrant(ctx context.Context, userID, deploymentScope string, deploymentID int64) (*UserGrant, error) {
	resp, err := schema.ListUsers(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("GetUserGrant: %w", err)
	}

	switch r := resp.UsersOrError.(type) {
	case *schema.ListUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants:
		for _, u := range r.Users {
			if fmt.Sprintf("%d", u.UserWithGrantsFields.User.UserId) != userID {
				continue
			}
			g := findUserGrantInFields(u.UserWithGrantsFields, deploymentScope, deploymentID)
			if g == nil {
				return nil, fmt.Errorf("GetUserGrant: grant for user %q scope %q not found", userID, deploymentScope)
			}
			return g, nil
		}
		return nil, fmt.Errorf("GetUserGrant: user %q not found", userID)
	default:
		return nil, fmt.Errorf("GetUserGrant: unexpected result type %T", resp.UsersOrError)
	}
}

// ListUserDeploymentGrants returns all deployment-scoped grants for a user.
func (c *Client) ListUserDeploymentGrants(ctx context.Context, userID string) ([]UserGrant, error) {
	resp, err := schema.ListUsers(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListUserDeploymentGrants: %w", err)
	}
	switch r := resp.UsersOrError.(type) {
	case *schema.ListUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants:
		for _, u := range r.Users {
			if fmt.Sprintf("%d", u.UserWithGrantsFields.User.UserId) != userID {
				continue
			}
			var grants []UserGrant
			for _, g := range u.UserWithGrantsFields.DeploymentPermissionGrants {
				grants = append(grants, *userGrantFromFields(userID, u.UserWithGrantsFields.User.Email, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, g.LocationGrants))
			}
			return grants, nil
		}
		return nil, fmt.Errorf("ListUserDeploymentGrants: user %q not found", userID)
	default:
		return nil, fmt.Errorf("ListUserDeploymentGrants: unexpected result type %T", resp.UsersOrError)
	}
}

// ListUserBranchDeploymentsGrants returns all per-base-branch grants for a user.
func (c *Client) ListUserBranchDeploymentsGrants(ctx context.Context, userID string) ([]UserGrant, error) {
	resp, err := schema.ListUsers(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListUserBranchDeploymentsGrants: %w", err)
	}
	switch r := resp.UsersOrError.(type) {
	case *schema.ListUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants:
		for _, u := range r.Users {
			if fmt.Sprintf("%d", u.UserWithGrantsFields.User.UserId) != userID {
				continue
			}
			var grants []UserGrant
			for _, g := range u.UserWithGrantsFields.PerBaseBranchDeploymentsPermissionGrants {
				grants = append(grants, *userGrantFromFields(userID, u.UserWithGrantsFields.User.Email, g.Grant, g.CustomRoleId, g.DeploymentScope, g.DeploymentId, nil))
			}
			return grants, nil
		}
		return nil, fmt.Errorf("ListUserBranchDeploymentsGrants: user %q not found", userID)
	default:
		return nil, fmt.Errorf("ListUserBranchDeploymentsGrants: unexpected result type %T", resp.UsersOrError)
	}
}

// DeleteUserGrant removes a permission grant for a user at the given scope.
func (c *Client) DeleteUserGrant(ctx context.Context, email, deploymentScope string, deploymentID int64) error {
	_, err := schema.DeleteUserGrant(ctx, c.gqlClient(""),
		email,
		grantScopeToAPI[deploymentScope],
		int(deploymentID),
	)
	if err != nil {
		return fmt.Errorf("DeleteUserGrant: %w", err)
	}
	return nil
}
