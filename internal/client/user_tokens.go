package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// UserToken represents a Dagster+ user token.
type UserToken struct {
	ID     string
	UserID string
	Name   string
	Token  string // only populated on creation
}

// CreateUserToken creates a new user token for the current user.
func (c *Client) CreateUserToken(ctx context.Context, name string) (*UserToken, error) {
	resp, err := schema.CreateUserToken(ctx, c.gqlClient(""), name)
	if err != nil {
		return nil, fmt.Errorf("CreateUserToken: %w", err)
	}

	switch r := resp.CreateUserToken.(type) {
	case *schema.CreateUserTokenCreateUserTokenDagsterCloudUserToken:
		if r.UserTokenFields.Token == "" {
			return nil, fmt.Errorf("CreateUserToken: no token in response")
		}
		return &UserToken{
			ID:     strconv.Itoa(r.UserTokenFields.Id),
			UserID: strconv.Itoa(r.UserTokenFields.User.UserId),
			Name:   name,
			Token:  r.UserTokenFields.Token,
		}, nil
	default:
		return nil, fmt.Errorf("CreateUserToken: unexpected result type %T", resp.CreateUserToken)
	}
}

// ListUserTokens returns all tokens for the given user.
func (c *Client) ListUserTokens(ctx context.Context, userID string) ([]UserToken, error) {
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return nil, fmt.Errorf("ListUserTokens: invalid userID %q: %w", userID, err)
	}

	resp, err := schema.ListUserTokens(ctx, c.gqlClient(""), intUserID)
	if err != nil {
		return nil, fmt.Errorf("ListUserTokens: %w", err)
	}

	switch r := resp.UserTokensOrError.(type) {
	case *schema.ListUserTokensUserTokensOrErrorDagsterCloudUserTokens:
		tokens := make([]UserToken, len(r.Tokens))
		for i, t := range r.Tokens {
			tokens[i] = UserToken{
				ID:     strconv.Itoa(t.UserTokenFields.Id),
				UserID: userID,
				Name:   t.UserTokenFields.Description,
			}
		}
		return tokens, nil
	default:
		return nil, fmt.Errorf("ListUserTokens: unexpected result type %T", resp.UserTokensOrError)
	}
}

// GetUserToken retrieves a user token by ID. userID is required to query the API.
func (c *Client) GetUserToken(ctx context.Context, id, userID string) (*UserToken, error) {
	tokens, err := c.ListUserTokens(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		if tokens[i].ID == id {
			return &tokens[i], nil
		}
	}
	return nil, fmt.Errorf("GetUserToken: token %q not found", id)
}

// DeleteUserToken revokes a user token by ID.
func (c *Client) DeleteUserToken(ctx context.Context, tokenID, userID string) error {
	intTokenID, err := strconv.Atoi(tokenID)
	if err != nil {
		return fmt.Errorf("DeleteUserToken: invalid tokenId %q: %w", tokenID, err)
	}
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return fmt.Errorf("DeleteUserToken: invalid userId %q: %w", userID, err)
	}
	_, err = schema.RevokeUserToken(ctx, c.gqlClient(""), intTokenID, intUserID)
	if err != nil {
		return fmt.Errorf("DeleteUserToken: %w", err)
	}
	return nil
}
