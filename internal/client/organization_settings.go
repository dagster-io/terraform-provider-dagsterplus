package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// GetOrganizationSettings retrieves the organization-level settings as a JSON string.
func (c *Client) GetOrganizationSettings(ctx context.Context) (string, error) {
	resp, err := schema.GetOrganizationSettings(ctx, c.gqlClient(""))
	if err != nil {
		return "", fmt.Errorf("GetOrganizationSettings: %w", err)
	}

	raw := resp.OrganizationSettings.Settings
	if len(raw) == 0 || string(raw) == "null" {
		return "{}", nil
	}
	return string(raw), nil
}

// SetOrganizationSettings updates the organization-level settings.
// settingsJSON must be a valid JSON object string.
func (c *Client) SetOrganizationSettings(ctx context.Context, settingsJSON string) error {
	var raw json.RawMessage
	if settingsJSON == "" || settingsJSON == "{}" {
		raw = json.RawMessage("{}")
	} else {
		raw = json.RawMessage(settingsJSON)
	}

	resp, err := schema.SetOrganizationSettings(ctx, c.gqlClient(""), raw)
	if err != nil {
		return fmt.Errorf("SetOrganizationSettings: %w", err)
	}

	switch resp.SetOrganizationSettings.(type) {
	case *schema.SetOrganizationSettingsSetOrganizationSettings:
		return nil
	default:
		return fmt.Errorf("SetOrganizationSettings: unexpected result type %T", resp.SetOrganizationSettings)
	}
}
