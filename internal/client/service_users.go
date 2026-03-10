package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// ServiceUser represents a Dagster+ service user.
type ServiceUser struct {
	ID          string
	Name        string
	Description string
}

// ServiceToken represents a Dagster+ service token.
type ServiceToken struct {
	ID            string
	Token         string // only populated on creation
	Description   string
	ServiceUserID string
	Revoked       bool
}

// CreateServiceUser creates a new service user.
func (c *Client) CreateServiceUser(ctx context.Context, name, description string) (*ServiceUser, error) {
	resp, err := schema.CreateServiceUser(ctx, c.gqlClient(""), name, description)
	if err != nil {
		return nil, fmt.Errorf("CreateServiceUser: %w", err)
	}

	switch r := resp.CreateServiceUser.(type) {
	case *schema.CreateServiceUserCreateServiceUserServiceUserWithScopedPermissionGrants:
		return &ServiceUser{
			ID:          r.Id,
			Name:        r.ServiceUser.Name,
			Description: r.ServiceUser.Description,
		}, nil
	case *schema.CreateServiceUserCreateServiceUserServiceUserNameConflictError:
		return nil, fmt.Errorf("CreateServiceUser: %s", r.Message)
	default:
		return nil, fmt.Errorf("CreateServiceUser: unexpected result type %T", resp.CreateServiceUser)
	}
}

// UpdateServiceUser updates a service user's name and/or description.
func (c *Client) UpdateServiceUser(ctx context.Context, id, name, description string) (*ServiceUser, error) {
	resp, err := schema.UpdateServiceUser(ctx, c.gqlClient(""), id, name, description)
	if err != nil {
		return nil, fmt.Errorf("UpdateServiceUser: %w", err)
	}

	switch r := resp.UpdateServiceUser.(type) {
	case *schema.UpdateServiceUserUpdateServiceUserServiceUserWithScopedPermissionGrants:
		return &ServiceUser{
			ID:          r.Id,
			Name:        r.ServiceUser.Name,
			Description: r.ServiceUser.Description,
		}, nil
	case *schema.UpdateServiceUserUpdateServiceUserServiceUserNameConflictError:
		return nil, fmt.Errorf("UpdateServiceUser: %s", r.Message)
	case *schema.UpdateServiceUserUpdateServiceUserServiceUserNotFoundError:
		return nil, fmt.Errorf("UpdateServiceUser: %s", r.Message)
	default:
		return nil, fmt.Errorf("UpdateServiceUser: unexpected result type %T", resp.UpdateServiceUser)
	}
}

// DeleteServiceUser deletes a service user by ID.
func (c *Client) DeleteServiceUser(ctx context.Context, serviceUserID string) error {
	resp, err := schema.DeleteServiceUser(ctx, c.gqlClient(""), serviceUserID)
	if err != nil {
		return fmt.Errorf("DeleteServiceUser: %w", err)
	}

	switch r := resp.DeleteServiceUser.(type) {
	case *schema.DeleteServiceUserDeleteServiceUserDeleteServiceUserSuccess:
		if !r.Success {
			return fmt.Errorf("DeleteServiceUser: API returned success=false")
		}
		return nil
	case *schema.DeleteServiceUserDeleteServiceUserServiceUserNotFoundError:
		return fmt.Errorf("DeleteServiceUser: %s", r.Message)
	default:
		return fmt.Errorf("DeleteServiceUser: unexpected result type %T", resp.DeleteServiceUser)
	}
}

// GetServiceUser retrieves a service user by ID.
func (c *Client) GetServiceUser(ctx context.Context, serviceUserID string) (*ServiceUser, error) {
	resp, err := schema.GetServiceUser(ctx, c.gqlClient(""), serviceUserID)
	if err != nil {
		return nil, fmt.Errorf("GetServiceUser: %w", err)
	}

	switch r := resp.ServiceUser.(type) {
	case *schema.GetServiceUserServiceUserServiceUserWithScopedPermissionGrants:
		return &ServiceUser{
			ID:          r.Id,
			Name:        r.ServiceUser.Name,
			Description: r.ServiceUser.Description,
		}, nil
	case *schema.GetServiceUserServiceUserServiceUserNotFoundError:
		return nil, fmt.Errorf("GetServiceUser: service user %q not found", serviceUserID)
	default:
		return nil, fmt.Errorf("GetServiceUser: unexpected result type %T", resp.ServiceUser)
	}
}

// GetServiceUserByName retrieves a service user by name.
func (c *Client) GetServiceUserByName(ctx context.Context, name string) (*ServiceUser, error) {
	users, err := c.ListServiceUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i].Name == name {
			return &users[i], nil
		}
	}
	return nil, fmt.Errorf("GetServiceUserByName: service user %q not found", name)
}

// ListServiceUsers returns all service users in the organization.
func (c *Client) ListServiceUsers(ctx context.Context) ([]ServiceUser, error) {
	resp, err := schema.ListServiceUsers(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListServiceUsers: %w", err)
	}

	switch r := resp.ServiceUsersOrError.(type) {
	case *schema.ListServiceUsersServiceUsersOrErrorServiceUsersWithScopedPermissionGrants:
		result := make([]ServiceUser, len(r.ServiceUsers))
		for i, u := range r.ServiceUsers {
			result[i] = ServiceUser{
				ID:          u.Id,
				Name:        u.ServiceUser.Name,
				Description: u.ServiceUser.Description,
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("ListServiceUsers: unexpected result type %T", resp.ServiceUsersOrError)
	}
}

// CreateServiceToken creates a new service token for a service user.
func (c *Client) CreateServiceToken(ctx context.Context, serviceUserID, description string) (*ServiceToken, error) {
	resp, err := schema.CreateServiceToken(ctx, c.gqlClient(""), serviceUserID, description)
	if err != nil {
		return nil, fmt.Errorf("CreateServiceToken: %w", err)
	}

	switch r := resp.CreateServiceToken.(type) {
	case *schema.CreateServiceTokenCreateServiceToken:
		return &ServiceToken{
			ID:            r.Id,
			Token:         r.Token,
			Description:   r.Description,
			ServiceUserID: serviceUserID,
			Revoked:       r.Revoked,
		}, nil
	default:
		return nil, fmt.Errorf("CreateServiceToken: unexpected result type %T", resp.CreateServiceToken)
	}
}

// RevokeServiceToken revokes a service token.
func (c *Client) RevokeServiceToken(ctx context.Context, tokenID string) error {
	_, err := schema.RevokeServiceToken(ctx, c.gqlClient(""), tokenID)
	if err != nil {
		return fmt.Errorf("RevokeServiceToken: %w", err)
	}
	return nil
}

// UpdateServiceToken updates the description of a service token.
func (c *Client) UpdateServiceToken(ctx context.Context, tokenID, serviceUserID, description string) (*ServiceToken, error) {
	resp, err := schema.UpdateServiceToken(ctx, c.gqlClient(""), tokenID, serviceUserID, description)
	if err != nil {
		return nil, fmt.Errorf("UpdateServiceToken: %w", err)
	}

	switch r := resp.UpdateServiceToken.(type) {
	case *schema.UpdateServiceTokenUpdateServiceToken:
		return &ServiceToken{
			ID:            r.Id,
			Token:         r.Token,
			Description:   r.Description,
			ServiceUserID: serviceUserID,
			Revoked:       r.Revoked,
		}, nil
	default:
		return nil, fmt.Errorf("UpdateServiceToken: unexpected result type %T", resp.UpdateServiceToken)
	}
}

// GetServiceToken retrieves a service token by its ID within a service user's tokens.
func (c *Client) GetServiceToken(ctx context.Context, serviceUserID, tokenID string) (*ServiceToken, error) {
	tokens, err := c.ListServiceTokens(ctx, serviceUserID)
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		if tokens[i].ID == tokenID {
			return &tokens[i], nil
		}
	}
	return nil, fmt.Errorf("GetServiceToken: token %q not found for service user %q", tokenID, serviceUserID)
}

// ListServiceTokens returns all service tokens for a service user.
func (c *Client) ListServiceTokens(ctx context.Context, serviceUserID string) ([]ServiceToken, error) {
	resp, err := schema.ListServiceTokens(ctx, c.gqlClient(""), serviceUserID)
	if err != nil {
		return nil, fmt.Errorf("ListServiceTokens: %w", err)
	}

	switch r := resp.ServiceTokensOrError.(type) {
	case *schema.ListServiceTokensServiceTokensOrErrorServiceTokens:
		result := make([]ServiceToken, len(r.Tokens))
		for i, t := range r.Tokens {
			result[i] = ServiceToken{
				ID:            t.Id,
				Description:   t.Description,
				ServiceUserID: t.ServiceUser.Id,
				Revoked:       t.Revoked,
				// Token is intentionally empty — not returned by list
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("ListServiceTokens: unexpected result type %T", resp.ServiceTokensOrError)
	}
}
