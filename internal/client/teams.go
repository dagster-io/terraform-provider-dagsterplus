package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// Team represents a Dagster+ team.
type Team struct {
	ID   string
	Name string
}

// CreateTeam creates a new team in the organization.
func (c *Client) CreateTeam(ctx context.Context, name string) (*Team, error) {
	resp, err := schema.CreateOrUpdateTeam(ctx, c.gqlClient(""), name)
	if err != nil {
		return nil, fmt.Errorf("CreateTeam: %w", err)
	}

	switch r := resp.CreateOrUpdateTeam.(type) {
	case *schema.CreateOrUpdateTeamCreateOrUpdateTeamCreateOrUpdateTeamSuccess:
		return &Team{ID: r.Team.TeamFields.Id, Name: r.Team.TeamFields.Name}, nil
	default:
		return nil, fmt.Errorf("CreateTeam: unexpected result type %T", resp.CreateOrUpdateTeam)
	}
}

// GetTeam retrieves a team by ID.
func (c *Client) GetTeam(ctx context.Context, id string) (*Team, error) {
	teams, err := c.ListTeams(ctx)
	if err != nil {
		return nil, err
	}
	for i := range teams {
		if teams[i].ID == id {
			return &teams[i], nil
		}
	}
	return nil, fmt.Errorf("GetTeam: team %q not found", id)
}

// ListTeams returns all teams in the organization.
func (c *Client) ListTeams(ctx context.Context) ([]Team, error) {
	resp, err := schema.ListTeams(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListTeams: %w", err)
	}

	teams := make([]Team, len(resp.TeamPermissions))
	for i, tp := range resp.TeamPermissions {
		teams[i] = Team{
			ID:   tp.Team.TeamFields.Id,
			Name: tp.Team.TeamFields.Name,
		}
	}
	return teams, nil
}

// RenameTeam updates the name of an existing team.
func (c *Client) RenameTeam(ctx context.Context, id, name string) (*Team, error) {
	resp, err := schema.RenameTeam(ctx, c.gqlClient(""), id, name)
	if err != nil {
		return nil, fmt.Errorf("RenameTeam: %w", err)
	}

	switch r := resp.RenameTeam.(type) {
	case *schema.RenameTeamRenameTeamDagsterCloudTeam:
		return &Team{ID: r.TeamFields.Id, Name: r.TeamFields.Name}, nil
	default:
		return nil, fmt.Errorf("RenameTeam: unexpected result type %T", resp.RenameTeam)
	}
}

// DeleteTeam removes a team from the organization.
func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	_, err := schema.DeleteTeam(ctx, c.gqlClient(""), id)
	if err != nil {
		return fmt.Errorf("DeleteTeam: %w", err)
	}
	return nil
}
