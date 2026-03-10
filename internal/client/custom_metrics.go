package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// CustomMetric represents a Dagster+ custom metric.
type CustomMetric struct {
	ID              string
	MetadataKey     string
	DisplayName     string
	Description     string
	CreateTimestamp float64
	UpdateTimestamp float64
}

// CreateCustomMetric creates a new custom metric.
func (c *Client) CreateCustomMetric(ctx context.Context, metadataKey, displayName, description string) (*CustomMetric, error) {
	resp, err := schema.CreateCustomMetric(ctx, c.gqlClient(""), metadataKey, displayName, description)
	if err != nil {
		return nil, fmt.Errorf("CreateCustomMetric: %w", err)
	}

	switch r := resp.CreateCustomMetric.(type) {
	case *schema.CreateCustomMetricCreateCustomMetricCreateCustomMetricSuccess:
		return &CustomMetric{
			ID:              r.CustomMetric.Id,
			MetadataKey:     r.CustomMetric.MetadataKey,
			DisplayName:     r.CustomMetric.DisplayName,
			Description:     r.CustomMetric.Description,
			CreateTimestamp: r.CustomMetric.CreateTimestamp,
			UpdateTimestamp: r.CustomMetric.UpdateTimestamp,
		}, nil
	case *schema.CreateCustomMetricCreateCustomMetricCustomMetricError:
		return nil, fmt.Errorf("CreateCustomMetric: %s", r.Message)
	default:
		return nil, fmt.Errorf("CreateCustomMetric: unexpected result type %T", resp.CreateCustomMetric)
	}
}

// UpdateCustomMetric updates a custom metric's display name and/or description.
func (c *Client) UpdateCustomMetric(ctx context.Context, customMetricID, displayName, description string) (*CustomMetric, error) {
	resp, err := schema.UpdateCustomMetric(ctx, c.gqlClient(""), customMetricID, displayName, description)
	if err != nil {
		return nil, fmt.Errorf("UpdateCustomMetric: %w", err)
	}

	switch r := resp.UpdateCustomMetric.(type) {
	case *schema.UpdateCustomMetricUpdateCustomMetricUpdateCustomMetricSuccess:
		return &CustomMetric{
			ID:              r.CustomMetric.Id,
			MetadataKey:     r.CustomMetric.MetadataKey,
			DisplayName:     r.CustomMetric.DisplayName,
			Description:     r.CustomMetric.Description,
			CreateTimestamp: r.CustomMetric.CreateTimestamp,
			UpdateTimestamp: r.CustomMetric.UpdateTimestamp,
		}, nil
	case *schema.UpdateCustomMetricUpdateCustomMetricCustomMetricError:
		return nil, fmt.Errorf("UpdateCustomMetric: %s", r.Message)
	default:
		return nil, fmt.Errorf("UpdateCustomMetric: unexpected result type %T", resp.UpdateCustomMetric)
	}
}

// DeleteCustomMetric deletes a custom metric by ID.
func (c *Client) DeleteCustomMetric(ctx context.Context, customMetricID string) error {
	resp, err := schema.DeleteCustomMetric(ctx, c.gqlClient(""), customMetricID)
	if err != nil {
		return fmt.Errorf("DeleteCustomMetric: %w", err)
	}

	switch r := resp.DeleteCustomMetric.(type) {
	case *schema.DeleteCustomMetricDeleteCustomMetricDeleteCustomMetricSuccess:
		_ = r
		return nil
	case *schema.DeleteCustomMetricDeleteCustomMetricCustomMetricError:
		return fmt.Errorf("DeleteCustomMetric: %s", r.Message)
	default:
		return fmt.Errorf("DeleteCustomMetric: unexpected result type %T", resp.DeleteCustomMetric)
	}
}

// GetCustomMetric retrieves a custom metric by ID.
func (c *Client) GetCustomMetric(ctx context.Context, customMetricID string) (*CustomMetric, error) {
	metrics, err := c.ListCustomMetrics(ctx)
	if err != nil {
		return nil, err
	}
	for i := range metrics {
		if metrics[i].ID == customMetricID {
			return &metrics[i], nil
		}
	}
	return nil, fmt.Errorf("GetCustomMetric: custom metric %q not found", customMetricID)
}

// ListCustomMetrics returns all custom metrics in the organization.
func (c *Client) ListCustomMetrics(ctx context.Context) ([]CustomMetric, error) {
	resp, err := schema.ListCustomMetrics(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListCustomMetrics: %w", err)
	}

	result := make([]CustomMetric, len(resp.CustomMetrics))
	for i, m := range resp.CustomMetrics {
		result[i] = CustomMetric{
			ID:              m.Id,
			MetadataKey:     m.MetadataKey,
			DisplayName:     m.DisplayName,
			Description:     m.Description,
			CreateTimestamp: m.CreateTimestamp,
			UpdateTimestamp: m.UpdateTimestamp,
		}
	}
	return result, nil
}
