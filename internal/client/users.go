package client

import (
	"context"
	"fmt"
)

// User represents a Dagster+ organization member.
type User struct {
	ID    string `json:"userId"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string // populated from organizationPermissionGrant.grant
}

// userWithGrants is the raw shape returned by the API.
type userWithGrants struct {
	User struct {
		ID    int64  `json:"userId"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
	OrgGrant *struct {
		Grant string `json:"grant"`
	} `json:"organizationPermissionGrant"`
}

func (u *userWithGrants) toUser() User {
	role := ""
	if u.OrgGrant != nil {
		role = u.OrgGrant.Grant
	}
	return User{
		ID:    fmt.Sprintf("%d", u.User.ID),
		Email: u.User.Email,
		Name:  u.User.Name,
		Role:  role,
	}
}

// AddUser adds a user to the organization by email.
func (c *Client) AddUser(ctx context.Context, email string) (*User, error) {
	const mutation = `
mutation AddUser($email: String!) {
  addUserToOrganization(email: $email) {
    __typename
    ... on AddUserToOrganizationSuccess {
      userWithGrants {
        user {
          userId
          email
          name
        }
        organizationPermissionGrant {
          grant
        }
      }
    }
  }
}`

	var result struct {
		AddUserToOrganization struct {
			Typename       string          `json:"__typename"`
			UserWithGrants *userWithGrants `json:"userWithGrants"`
		} `json:"addUserToOrganization"`
	}

	if err := c.doGraphQL(ctx, "", mutation, map[string]any{"email": email}, &result); err != nil {
		return nil, fmt.Errorf("AddUser: %w", err)
	}

	if result.AddUserToOrganization.UserWithGrants == nil {
		return nil, fmt.Errorf("AddUser: unexpected result type %q", result.AddUserToOrganization.Typename)
	}

	u := result.AddUserToOrganization.UserWithGrants.toUser()
	return &u, nil
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
	ids := make([]string, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return nil, fmt.Errorf("GetUser: user %q not found (known IDs: %v)", id, ids)
}

// ListUsers returns all members of the organization.
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	const query = `
query GetUsers {
  usersOrError {
    __typename
    ... on DagsterCloudUsersWithScopedPermissionGrants {
      users {
        user {
          userId
          email
          name
        }
        organizationPermissionGrant {
          grant
        }
      }
    }
  }
}`

	var result struct {
		UsersOrError struct {
			Typename string           `json:"__typename"`
			Users    []userWithGrants `json:"users"`
		} `json:"usersOrError"`
	}

	if err := c.doGraphQL(ctx, "", query, nil, &result); err != nil {
		return nil, fmt.Errorf("ListUsers: %w", err)
	}

	if result.UsersOrError.Typename != "DagsterCloudUsersWithScopedPermissionGrants" {
		return nil, fmt.Errorf("ListUsers: unexpected result type %q", result.UsersOrError.Typename)
	}

	users := make([]User, len(result.UsersOrError.Users))
	for i, u := range result.UsersOrError.Users {
		users[i] = u.toUser()
	}
	return users, nil
}

// RemoveUser removes a member from the organization by email.
func (c *Client) RemoveUser(ctx context.Context, email string) error {
	const mutation = `
mutation RemoveUser($email: String!) {
  removeUserFromOrganization(email: $email) {
    ... on RemoveUserFromOrganizationSuccess {
      email
    }
  }
}`

	if err := c.doGraphQL(ctx, "", mutation, map[string]any{"email": email}, nil); err != nil {
		return fmt.Errorf("RemoveUser: %w", err)
	}

	return nil
}
