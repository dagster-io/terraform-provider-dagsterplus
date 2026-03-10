package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// AddUserToTeam adds a user to a team.
func (c *Client) AddUserToTeam(ctx context.Context, teamID, userID string) error {
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return fmt.Errorf("AddUserToTeam: invalid userId %q: %w", userID, err)
	}
	_, err = schema.AddMemberToTeam(ctx, c.gqlClient(""), teamID, intUserID)
	if err != nil {
		return fmt.Errorf("AddUserToTeam: %w", err)
	}
	return nil
}

// RemoveUserFromTeam removes a user from a team.
func (c *Client) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return fmt.Errorf("RemoveUserFromTeam: invalid userId %q: %w", userID, err)
	}
	_, err = schema.RemoveMemberFromTeam(ctx, c.gqlClient(""), teamID, intUserID)
	if err != nil {
		return fmt.Errorf("RemoveUserFromTeam: %w", err)
	}
	return nil
}

// IsUserInTeam checks whether a user is a member of a team.
func (c *Client) IsUserInTeam(ctx context.Context, teamID, userID string) (bool, error) {
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return false, fmt.Errorf("IsUserInTeam: invalid userId %q: %w", userID, err)
	}

	resp, err := schema.ListTeamMembers(ctx, c.gqlClient(""))
	if err != nil {
		return false, fmt.Errorf("IsUserInTeam: %w", err)
	}

	for _, tp := range resp.TeamPermissions {
		if tp.Team.Id == teamID {
			for _, m := range tp.Team.Members {
				if m.UserId == intUserID {
					return true, nil
				}
			}
			return false, nil
		}
	}
	return false, nil
}
