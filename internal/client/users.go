package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// User represents a Dagster+ organization member.
type User struct {
	ID                string
	Email             string
	Name              string
	Role              string // populated from organizationPermissionGrant.grant
	Picture           string
	IsScimProvisioned bool
}

func userFromFields(f schema.UserWithGrantsFields) User {
	role := ""
	if f.OrganizationPermissionGrant.ScopedGrantFields.Grant != "" {
		role = string(f.OrganizationPermissionGrant.ScopedGrantFields.Grant)
	}
	return User{
		ID:                fmt.Sprintf("%d", f.User.UserId),
		Email:             f.User.Email,
		Name:              f.User.Name,
		Role:              role,
		Picture:           f.User.Picture,
		IsScimProvisioned: f.User.IsScimProvisioned,
	}
}

// AddUser adds a user to the organization by email.
func (c *Client) AddUser(ctx context.Context, email string) (*User, error) {
	resp, err := schema.AddUser(ctx, c.gqlClient(""), email)
	if err != nil {
		return nil, fmt.Errorf("AddUser: %w", err)
	}

	switch r := resp.AddUserToOrganization.(type) {
	case *schema.AddUserAddUserToOrganizationAddUserToOrganizationSuccess:
		u := userFromFields(r.UserWithGrants.UserWithGrantsFields)
		return &u, nil
	default:
		return nil, fmt.Errorf("AddUser: unexpected result type %T", resp.AddUserToOrganization)
	}
}

// GetUser retrieves a user by ID.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].ID == id {
			return &users[i], nil
		}
	}
	return nil, fmt.Errorf("GetUser: user %q not found", id)
}

// GetUserByEmail retrieves a user by email address.
// It first searches the results of ListUsers; if not found there (e.g. for the
// org owner who was not added via the invite flow), it falls back to the
// organization-level user list which includes all members.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].Email == email {
			return &users[i], nil
		}
	}

	// Fallback: the org owner (and any other user not returned by the root
	// usersOrError query) may appear in organization.usersOrError.
	resp, err := schema.ListAllOrgUsers(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail fallback: %w", err)
	}
	switch r := resp.Organization.UsersOrError.(type) {
	case *schema.ListAllOrgUsersOrganizationDagsterCloudOrganizationUsersOrErrorDagsterCloudUserList:
		for _, u := range r.Users {
			if u.Email == email {
				return &User{
					ID:                fmt.Sprintf("%d", u.UserId),
					Email:             u.Email,
					Name:              u.Name,
					Picture:           u.Picture,
					IsScimProvisioned: u.IsScimProvisioned,
				}, nil
			}
		}
		return nil, fmt.Errorf("GetUserByEmail: user %q not found", email)
	default:
		return nil, fmt.Errorf("GetUserByEmail: unexpected result type %T from organization.usersOrError", r)
	}
}

// ListUsers returns all members of the organization.
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	resp, err := schema.ListUsers(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListUsers: %w", err)
	}

	switch r := resp.UsersOrError.(type) {
	case *schema.ListUsersUsersOrErrorDagsterCloudUsersWithScopedPermissionGrants:
		users := make([]User, len(r.Users))
		for i, u := range r.Users {
			users[i] = userFromFields(u.UserWithGrantsFields)
		}
		return users, nil
	default:
		return nil, fmt.Errorf("ListUsers: unexpected result type %T", resp.UsersOrError)
	}
}

// RemoveUser removes a member from the organization by email.
func (c *Client) RemoveUser(ctx context.Context, email string) error {
	_, err := schema.RemoveUser(ctx, c.gqlClient(""), email)
	if err != nil {
		return fmt.Errorf("RemoveUser: %w", err)
	}
	return nil
}
