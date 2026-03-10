package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// Organization represents Dagster+ org-level metadata.
type Organization struct {
	IntID    int
	PublicID string
	Name     string
	Status   string
	Slack    *SlackInstallation
	GitHub   *GitHubInstallation
}

// SlackInstallation holds Slack App installation info.
type SlackInstallation struct {
	TeamID   string
	TeamName string
}

// GitHubInstallation holds GitHub App installation info.
type GitHubInstallation struct {
	AccountName string
	AppID       string
	SettingsURL string
	Repos       []string
}

// AtlanIntegration holds Atlan integration settings.
type AtlanIntegration struct {
	Token  string
	Domain string
}

// GetVersion retrieves the Dagster version running in the given deployment.
func (c *Client) GetVersion(ctx context.Context, deployment string) (string, error) {
	resp, err := schema.GetVersion(ctx, c.gqlClient(deployment))
	if err != nil {
		return "", fmt.Errorf("GetVersion: %w", err)
	}
	return resp.Version, nil
}

// GetOrganization retrieves org-level metadata.
func (c *Client) GetOrganization(ctx context.Context) (*Organization, error) {
	resp, err := schema.GetOrganization(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: %w", err)
	}
	org := &Organization{
		IntID:    resp.Organization.Id,
		PublicID: resp.Organization.PublicId,
		Name:     resp.Organization.Name,
		Status:   string(resp.Organization.Status),
	}
	if s := resp.Organization.Metadata.SlackAppInstallation; s.TeamId != "" {
		org.Slack = &SlackInstallation{
			TeamID:   s.TeamId,
			TeamName: s.TeamName,
		}
	}
	if g := resp.Organization.Metadata.GithubAppInstallation; g.AppId != "" {
		org.GitHub = &GitHubInstallation{
			AccountName: g.AccountName,
			AppID:       g.AppId,
			SettingsURL: g.SettingsUrl,
			Repos:       g.Repos,
		}
	}
	return org, nil
}

// GetScimSyncEnabled returns whether SCIM sync is currently enabled.
func (c *Client) GetScimSyncEnabled(ctx context.Context) (bool, error) {
	resp, err := schema.GetScimSyncEnabled(ctx, c.gqlClient(""))
	if err != nil {
		return false, fmt.Errorf("GetScimSyncEnabled: %w", err)
	}
	return resp.ScimSyncEnabled, nil
}

// SetScimSyncEnabled enables or disables SCIM sync for the organization.
func (c *Client) SetScimSyncEnabled(ctx context.Context, enabled bool) error {
	resp, err := schema.SetScimSyncEnabled(ctx, c.gqlClient(""), enabled)
	if err != nil {
		return fmt.Errorf("SetScimSyncEnabled: %w", err)
	}
	switch resp.SetScimSyncEnabled.(type) {
	case *schema.SetScimSyncEnabledSetScimSyncEnabledSetScimSyncEnabledSuccess:
		return nil
	default:
		return fmt.Errorf("SetScimSyncEnabled: unexpected result type %T", resp.SetScimSyncEnabled)
	}
}

// GetAtlanIntegration retrieves the current Atlan integration settings.
// Returns nil if no integration is configured.
func (c *Client) GetAtlanIntegration(ctx context.Context) (*AtlanIntegration, error) {
	resp, err := schema.GetAtlanIntegrationSettings(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("GetAtlanIntegration: %w", err)
	}
	switch s := resp.AtlanIntegration.AtlanIntegrationSettingsOrError.(type) {
	case *schema.GetAtlanIntegrationSettingsAtlanIntegrationAtlanIntegrationQueryAtlanIntegrationSettingsOrErrorAtlanIntegrationSettings:
		return &AtlanIntegration{Token: s.Token, Domain: s.Domain}, nil
	case *schema.GetAtlanIntegrationSettingsAtlanIntegrationAtlanIntegrationQueryAtlanIntegrationSettingsOrErrorAtlanIntegrationSettingsUnset:
		return nil, nil
	default:
		return nil, fmt.Errorf("GetAtlanIntegration: unexpected result type %T", resp.AtlanIntegration.AtlanIntegrationSettingsOrError)
	}
}

// SetAtlanIntegration creates or replaces the Atlan integration settings.
func (c *Client) SetAtlanIntegration(ctx context.Context, token, domain string) error {
	resp, err := schema.SetAtlanIntegrationSettings(ctx, c.gqlClient(""), token, domain)
	if err != nil {
		return fmt.Errorf("SetAtlanIntegration: %w", err)
	}
	switch resp.SetAtlanIntegrationSettings.(type) {
	case *schema.SetAtlanIntegrationSettingsSetAtlanIntegrationSettingsSetAtlanIntegrationSettingsSuccess:
		return nil
	default:
		return fmt.Errorf("SetAtlanIntegration: unexpected result type %T", resp.SetAtlanIntegrationSettings)
	}
}

// DeleteAtlanIntegration removes the Atlan integration.
func (c *Client) DeleteAtlanIntegration(ctx context.Context) error {
	resp, err := schema.DeleteAtlanIntegrationSettings(ctx, c.gqlClient(""))
	if err != nil {
		return fmt.Errorf("DeleteAtlanIntegration: %w", err)
	}
	switch resp.DeleteAtlanIntegrationSettings.(type) {
	case *schema.DeleteAtlanIntegrationSettingsDeleteAtlanIntegrationSettingsDeleteAtlanIntegrationSuccess:
		return nil
	default:
		return fmt.Errorf("DeleteAtlanIntegration: unexpected result type %T", resp.DeleteAtlanIntegrationSettings)
	}
}

// SelectGithubInstallation selects the GitHub App installation to use.
func (c *Client) SelectGithubInstallation(ctx context.Context, accountName string) (*GitHubInstallation, error) {
	resp, err := schema.SelectGithubInstallation(ctx, c.gqlClient(""), accountName)
	if err != nil {
		return nil, fmt.Errorf("SelectGithubInstallation: %w", err)
	}
	switch r := resp.SelectInstallation.(type) {
	case *schema.SelectGithubInstallationSelectInstallationGithubAppInstallation:
		return &GitHubInstallation{
			AccountName: r.AccountName,
			AppID:       r.AppId,
			SettingsURL: r.SettingsUrl,
			Repos:       r.Repos,
		}, nil
	default:
		return nil, fmt.Errorf("SelectGithubInstallation: unexpected result type %T", resp.SelectInstallation)
	}
}

// DeselectGithubInstallation removes the active GitHub App installation.
func (c *Client) DeselectGithubInstallation(ctx context.Context) error {
	resp, err := schema.DeselectGithubInstallation(ctx, c.gqlClient(""))
	if err != nil {
		return fmt.Errorf("DeselectGithubInstallation: %w", err)
	}
	switch resp.DeselectInstallation.(type) {
	case *schema.DeselectGithubInstallationDeselectInstallationDeselectInstallationSuccess:
		return nil
	default:
		return fmt.Errorf("DeselectGithubInstallation: unexpected result type %T", resp.DeselectInstallation)
	}
}
