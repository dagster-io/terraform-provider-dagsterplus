package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// Role represents a Dagster+ custom role.
type Role struct {
	ID          string
	Name        string
	Description string
	Icon        string   // maps to iconName in the API
	RoleType    string   // "deployment" | "organization"
	Permissions []string // snake_case permission names
}

var roleTypeToAPI = map[string]schema.CustomRoleDeploymentScope{
	"deployment":   "DEPLOYMENT",
	"organization": "ORGANIZATION",
}

var apiToRoleType = map[schema.CustomRoleDeploymentScope]string{
	"DEPLOYMENT":   "deployment",
	"ORGANIZATION": "organization",
}

// permToAPI converts a snake_case permission name to the API enum value (e.g. "edit_alerts" → "EDIT_ALERTS").
func permToAPI(p string) schema.CustomRolePermission {
	return schema.CustomRolePermission(strings.ToUpper(p))
}

// apiToPerm converts an API enum value to snake_case (e.g. "EDIT_ALERTS" → "edit_alerts").
func apiToPerm(p schema.CustomRolePermission) string { return strings.ToLower(string(p)) }

func roleFromFields(f schema.CustomRoleFields) Role {
	rt := apiToRoleType[f.DeploymentScope]
	if rt == "" {
		rt = string(f.DeploymentScope)
	}
	perms := make([]string, len(f.Permissions))
	for i, p := range f.Permissions {
		perms[i] = apiToPerm(p)
	}
	return Role{
		ID:          f.Id,
		Name:        f.Name,
		Description: f.Description,
		Icon:        f.IconName,
		RoleType:    rt,
		Permissions: perms,
	}
}

// CreateRole creates a new custom role in the organization.
func (c *Client) CreateRole(ctx context.Context, role Role) (*Role, error) {
	perms := make([]schema.CustomRolePermission, len(role.Permissions))
	for i, p := range role.Permissions {
		perms[i] = permToAPI(p)
	}

	resp, err := schema.CreateRole(ctx, c.gqlClient(""),
		role.Name, role.Description, role.Icon,
		roleTypeToAPI[role.RoleType], perms,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateRole: %w", err)
	}

	switch r := resp.CreateCustomRole.(type) {
	case *schema.CreateRoleCreateCustomRoleCreateOrUpdateCustomRoleSuccess:
		result := roleFromFields(r.CustomRole.CustomRoleFields)
		return &result, nil
	default:
		return nil, fmt.Errorf("CreateRole: unexpected result type %T", resp.CreateCustomRole)
	}
}

// GetRole retrieves a custom role by ID.
func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	roles, err := c.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].ID == id {
			return &roles[i], nil
		}
	}
	return nil, fmt.Errorf("GetRole: role %q not found", id)
}

// GetRoleByName retrieves a custom role by name.
func (c *Client) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	roles, err := c.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].Name == name {
			return &roles[i], nil
		}
	}
	return nil, fmt.Errorf("GetRoleByName: role %q not found", name)
}

// ListRoles returns all custom roles in the organization.
func (c *Client) ListRoles(ctx context.Context) ([]Role, error) {
	resp, err := schema.ListRoles(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListRoles: %w", err)
	}

	roles := make([]Role, len(resp.CustomRoles))
	for i, r := range resp.CustomRoles {
		roles[i] = roleFromFields(r.CustomRoleFields)
	}
	return roles, nil
}

// UpdateRole updates the name, description, icon, and permissions of an existing custom role.
func (c *Client) UpdateRole(ctx context.Context, role Role) (*Role, error) {
	perms := make([]schema.CustomRolePermission, len(role.Permissions))
	for i, p := range role.Permissions {
		perms[i] = permToAPI(p)
	}

	resp, err := schema.UpdateRole(ctx, c.gqlClient(""),
		role.ID, role.Name, role.Description, role.Icon, perms,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateRole: %w", err)
	}

	switch r := resp.UpdateCustomRole.(type) {
	case *schema.UpdateRoleUpdateCustomRoleCreateOrUpdateCustomRoleSuccess:
		result := roleFromFields(r.CustomRole.CustomRoleFields)
		return &result, nil
	default:
		return nil, fmt.Errorf("UpdateRole: unexpected result type %T", resp.UpdateCustomRole)
	}
}

// DeleteRole removes a custom role from the organization.
func (c *Client) DeleteRole(ctx context.Context, id string) error {
	_, err := schema.DeleteRole(ctx, c.gqlClient(""), id)
	if err != nil {
		return fmt.Errorf("DeleteRole: %w", err)
	}
	return nil
}
