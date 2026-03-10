package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// GetDeploymentSettings retrieves the settings for a deployment as a JSON string.
func (c *Client) GetDeploymentSettings(ctx context.Context, deployment string) (string, error) {
	resp, err := schema.GetDeploymentSettings(ctx, c.gqlClient(deployment))
	if err != nil {
		return "", fmt.Errorf("GetDeploymentSettings: %w", err)
	}

	raw := resp.DeploymentSettings.Settings
	if len(raw) == 0 || string(raw) == "null" {
		return "{}", nil
	}
	return string(raw), nil
}

// SetDeploymentSettings updates the settings for a deployment.
// settingsJSON must be a valid JSON object string.
func (c *Client) SetDeploymentSettings(ctx context.Context, deployment string, settingsJSON string) error {
	var raw json.RawMessage
	if settingsJSON == "" || settingsJSON == "{}" {
		raw = json.RawMessage("{}")
	} else {
		raw = json.RawMessage(settingsJSON)
	}

	resp, err := schema.SetDeploymentSettings(ctx, c.gqlClient(deployment), raw)
	if err != nil {
		return fmt.Errorf("SetDeploymentSettings: %w", err)
	}

	switch resp.SetDeploymentSettings.(type) {
	case *schema.SetDeploymentSettingsSetDeploymentSettings:
		return nil
	default:
		return fmt.Errorf("SetDeploymentSettings: unexpected result type %T", resp.SetDeploymentSettings)
	}
}
