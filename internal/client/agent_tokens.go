package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// AgentToken represents a Dagster+ agent token.
type AgentToken struct {
	ID    string
	Name  string
	Token string // only populated on creation
}

// CreateAgentToken creates a new agent token with the given label.
func (c *Client) CreateAgentToken(ctx context.Context, name string) (*AgentToken, error) {
	resp, err := schema.CreateAgentToken(ctx, c.gqlClient(""), name)
	if err != nil {
		return nil, fmt.Errorf("CreateAgentToken: %w", err)
	}

	switch r := resp.CreateAgentToken.(type) {
	case *schema.CreateAgentTokenCreateAgentTokenDagsterCloudAgentToken:
		if r.AgentTokenFields.Token == "" {
			return nil, fmt.Errorf("CreateAgentToken: no token in response")
		}
		return &AgentToken{
			ID:    strconv.Itoa(r.AgentTokenFields.Id),
			Name:  name,
			Token: r.AgentTokenFields.Token,
		}, nil
	default:
		return nil, fmt.Errorf("CreateAgentToken: unexpected result type %T", resp.CreateAgentToken)
	}
}

// ListAgentTokens returns all agent tokens for the organization.
func (c *Client) ListAgentTokens(ctx context.Context) ([]AgentToken, error) {
	resp, err := schema.ListAgentTokens(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListAgentTokens: %w", err)
	}

	switch r := resp.AgentTokensOrError.(type) {
	case *schema.ListAgentTokensAgentTokensOrErrorDagsterCloudAgentTokens:
		tokens := make([]AgentToken, len(r.Tokens))
		for i, t := range r.Tokens {
			tokens[i] = AgentToken{
				ID:   strconv.Itoa(t.AgentTokenFields.Id),
				Name: t.AgentTokenFields.Description,
			}
		}
		return tokens, nil
	default:
		return nil, fmt.Errorf("ListAgentTokens: unexpected result type %T", resp.AgentTokensOrError)
	}
}

// GetAgentToken retrieves an agent token by ID.
func (c *Client) GetAgentToken(ctx context.Context, id string) (*AgentToken, error) {
	tokens, err := c.ListAgentTokens(ctx)
	if err != nil {
		return nil, err
	}
	for i := range tokens {
		if tokens[i].ID == id {
			return &tokens[i], nil
		}
	}
	return nil, fmt.Errorf("GetAgentToken: token %q not found", id)
}

// DeleteAgentToken revokes an agent token by ID.
func (c *Client) DeleteAgentToken(ctx context.Context, id string) error {
	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("DeleteAgentToken: invalid id %q: %w", id, err)
	}
	_, err = schema.RevokeAgentToken(ctx, c.gqlClient(""), intID)
	if err != nil {
		return fmt.Errorf("DeleteAgentToken: %w", err)
	}
	return nil
}
