package client

import (
	"context"
	"fmt"
	"strings"
)

// AlertPolicy represents a Dagster+ alert policy.
type AlertPolicy struct {
	ID                  string
	Name                string
	Description         string
	PolicyType          string // "asset" | "run" | "code_location" | "automation" | "budget" | "insight_metric"
	Enabled             bool
	EventTypes          []string
	AlertTargetType     string // "all_assets" | "asset_selection" | "asset_key"
	AlertTargetValue    string // asset selection string or asset key (empty for all_assets)
	NotificationService AlertPolicyNotification
}

// AlertPolicyNotification holds the notification service config for an alert policy.
// Exactly one notification type should be populated, indicated by Type.
type AlertPolicyNotification struct {
	Type                  string   // "email" | "email_owners" | "slack" | "microsoft_teams" | "pagerduty"
	EmailAddresses        []string // email
	DefaultEmailAddresses []string // email_owners (optional fallback addresses)
	SlackWorkspaceName    string   // slack
	SlackChannelName      string   // slack
	WebhookURL            string   // microsoft_teams
	IntegrationKey        string   // pagerduty
}

// alertPolicyRaw is the GraphQL response shape for an alert policy.
type alertPolicyRaw struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	EventTypes  []string `json:"eventTypes"`
	AlertTargets []struct {
		Typename             string `json:"__typename"`
		AssetSelectionString string `json:"assetSelectionString"`
		AssetKey      struct {
			Path []string `json:"path"`
		} `json:"assetKey"`
	} `json:"alertTargets"`
	NotificationService struct {
		Typename              string   `json:"__typename"`
		EmailAddresses        []string `json:"emailAddresses"`
		DefaultEmailAddresses []string `json:"defaultEmailAddresses"`
		SlackWorkspaceName    string   `json:"slackWorkspaceName"`
		SlackChannelName      string   `json:"slackChannelName"`
		WebhookURL            string   `json:"webhookUrl"`
		IntegrationKey        string   `json:"integrationKey"`
	} `json:"notificationService"`
}

func (r *alertPolicyRaw) toAlertPolicy() AlertPolicy {
	ns := AlertPolicyNotification{}
	switch r.NotificationService.Typename {
	case "EmailAlertPolicyNotification":
		ns.Type = "email"
		ns.EmailAddresses = r.NotificationService.EmailAddresses
	case "EmailOwnersAlertPolicyNotification":
		ns.Type = "email_owners"
		ns.DefaultEmailAddresses = r.NotificationService.DefaultEmailAddresses
	case "SlackAlertPolicyNotification":
		ns.Type = "slack"
		ns.SlackWorkspaceName = r.NotificationService.SlackWorkspaceName
		ns.SlackChannelName = r.NotificationService.SlackChannelName
	case "MicrosoftTeamsAlertPolicyNotification":
		ns.Type = "microsoft_teams"
		ns.WebhookURL = r.NotificationService.WebhookURL
	case "PagerdutyAlertPolicyNotification":
		ns.Type = "pagerduty"
		ns.IntegrationKey = r.NotificationService.IntegrationKey
	}
	targetType := "all_assets"
	targetValue := ""
	if len(r.AlertTargets) > 0 {
		t := r.AlertTargets[0]
		switch t.Typename {
		case "AssetSelectionTarget":
			targetType = "asset_selection"
			targetValue = t.AssetSelectionString
		case "AssetKeyTarget":
			targetType = "asset_key"
			targetValue = strings.Join(t.AssetKey.Path, "/")
		}
	}

	return AlertPolicy{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Enabled:     r.Enabled,
		EventTypes:          r.EventTypes,
		AlertTargetType:     targetType,
		AlertTargetValue:    targetValue,
		NotificationService: ns,
	}
}

// CreateOrUpdateAlertPolicy creates or updates an alert policy in a deployment.
func (c *Client) CreateOrUpdateAlertPolicy(ctx context.Context, deployment string, policy AlertPolicy) (*AlertPolicy, error) {
	const mutation = `
mutation CreateOrUpdateAlertPolicy($document: GenericScalar!) {
  createOrUpdateAlertPolicyFromDocument(document: $document) {
    __typename
    ... on AlertPolicy {
      id
      name
    }
    ... on InvalidAlertPolicyError {
      message
      errors
    }
  }
}`

	var result struct {
		CreateOrUpdateAlertPolicyFromDocument struct {
			Typename string   `json:"__typename"`
			Message  string   `json:"message"`
			Errors   []string `json:"errors"`
		} `json:"createOrUpdateAlertPolicyFromDocument"`
	}

	if err := c.doGraphQL(ctx, deployment, mutation, map[string]any{"document": buildPolicyDoc(policy)}, &result); err != nil {
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: %w", err)
	}

	switch result.CreateOrUpdateAlertPolicyFromDocument.Typename {
	case "AlertPolicy":
		return c.GetAlertPolicy(ctx, deployment, policy.Name)
	case "InvalidAlertPolicyError":
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: invalid policy: %s %v",
			result.CreateOrUpdateAlertPolicyFromDocument.Message,
			result.CreateOrUpdateAlertPolicyFromDocument.Errors)
	default:
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: unexpected result type %q",
			result.CreateOrUpdateAlertPolicyFromDocument.Typename)
	}
}

// GetAlertPolicy retrieves an alert policy by name from a deployment.
func (c *Client) GetAlertPolicy(ctx context.Context, deployment, name string) (*AlertPolicy, error) {
	policies, err := c.ListAlertPolicies(ctx, deployment)
	if err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i].Name == name {
			return &policies[i], nil
		}
	}
	return nil, fmt.Errorf("GetAlertPolicy: policy %q not found in deployment %q", name, deployment)
}

// ListAlertPolicies returns all alert policies for a deployment.
func (c *Client) ListAlertPolicies(ctx context.Context, deployment string) ([]AlertPolicy, error) {
	const query = `
query ListAlertPolicies {
  alertPolicies {
    id
    name
    description
    enabled
    eventTypes
    alertTargets {
      __typename
      ... on AssetSelectionTarget {
        assetSelectionString
      }
      ... on AssetKeyTarget {
        assetKey {
          path
        }
      }
    }
    notificationService {
      __typename
      ... on EmailAlertPolicyNotification {
        emailAddresses
      }
      ... on EmailOwnersAlertPolicyNotification {
        defaultEmailAddresses
      }
      ... on SlackAlertPolicyNotification {
        slackWorkspaceName
        slackChannelName
      }
      ... on MicrosoftTeamsAlertPolicyNotification {
        webhookUrl
      }
      ... on PagerdutyAlertPolicyNotification {
        integrationKey
      }
    }
  }
}`

	var result struct {
		AlertPolicies []alertPolicyRaw `json:"alertPolicies"`
	}

	if err := c.doGraphQL(ctx, deployment, query, nil, &result); err != nil {
		return nil, fmt.Errorf("ListAlertPolicies: %w", err)
	}

	policies := make([]AlertPolicy, len(result.AlertPolicies))
	for i, p := range result.AlertPolicies {
		policies[i] = p.toAlertPolicy()
	}
	return policies, nil
}

// DeleteAlertPolicy deletes an alert policy by name from a deployment.
func (c *Client) DeleteAlertPolicy(ctx context.Context, deployment, name string) error {
	const mutation = `
mutation DeleteAlertPolicy($alertPolicyName: String!) {
  deleteAlertPolicy(alertPolicyName: $alertPolicyName) {
    __typename
    ... on DeleteAlertPolicySuccess {
      alertPolicyName
    }
  }
}`

	var result struct {
		DeleteAlertPolicy struct {
			Typename string `json:"__typename"`
		} `json:"deleteAlertPolicy"`
	}

	if err := c.doGraphQL(ctx, deployment, mutation, map[string]any{"alertPolicyName": name}, &result); err != nil {
		return fmt.Errorf("DeleteAlertPolicy: %w", err)
	}

	if result.DeleteAlertPolicy.Typename != "DeleteAlertPolicySuccess" {
		return fmt.Errorf("DeleteAlertPolicy: unexpected result type %q", result.DeleteAlertPolicy.Typename)
	}
	return nil
}

// buildPolicyDoc constructs the GenericScalar document for create/update mutations.
func buildPolicyDoc(policy AlertPolicy) map[string]any {
	doc := map[string]any{
		"name":        policy.Name,
		"description": policy.Description,
		"enabled":     policy.Enabled,
		"event_types": policy.EventTypes,
	}

	switch policy.AlertTargetType {
	case "asset_selection":
		doc["alert_targets"] = []any{
			map[string]any{"asset_selection_target": map[string]any{"asset_selection": policy.AlertTargetValue}},
		}
	case "asset_key":
		doc["alert_targets"] = []any{
			map[string]any{"asset_key_target": map[string]any{"asset_key": strings.Split(policy.AlertTargetValue, "/")}},
		}
	// "all_assets" and default: omit alert_targets (empty = all assets)
	}

	ns := policy.NotificationService
	switch ns.Type {
	case "email":
		doc["notification_service"] = map[string]any{
			"email": map[string]any{"email_addresses": ns.EmailAddresses},
		}
	case "email_owners":
		inner := map[string]any{}
		if len(ns.DefaultEmailAddresses) > 0 {
			inner["default_email_addresses"] = ns.DefaultEmailAddresses
		}
		doc["notification_service"] = map[string]any{"email_owners": inner}
	case "slack":
		doc["notification_service"] = map[string]any{
			"slack": map[string]any{
				"slack_workspace_name": ns.SlackWorkspaceName,
				"slack_channel_name":   ns.SlackChannelName,
			},
		}
	case "microsoft_teams":
		doc["notification_service"] = map[string]any{
			"microsoft_teams": map[string]any{"webhook_url": ns.WebhookURL},
		}
	case "pagerduty":
		doc["notification_service"] = map[string]any{
			"pagerduty": map[string]any{"integration_key": ns.IntegrationKey},
		}
	}

	return doc
}
