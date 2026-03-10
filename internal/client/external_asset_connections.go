package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// ExternalAssetConnection represents a Dagster+ external asset connection.
type ExternalAssetConnection struct {
	ID               string
	Name             string
	Description      string
	ConnectionType   string // SNOWFLAKE | BIGQUERY | POSTGRES | DATABRICKS
	SourceConfigYaml string
	CronString       string
	Timezone         string
	ScheduleStatus   string // RUNNING | STOPPED
}

// CreateExternalAssetConnection creates a new external asset connection.
func (c *Client) CreateExternalAssetConnection(ctx context.Context, conn ExternalAssetConnection) (*ExternalAssetConnection, error) {
	resp, err := schema.CreateExternalAssetConnection(ctx, c.gqlClient(""),
		schema.ExternalAssetConnectionType(conn.ConnectionType),
		conn.Name, conn.Description, conn.SourceConfigYaml, conn.CronString, conn.Timezone,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateExternalAssetConnection: %w", err)
	}

	switch r := resp.CreateExternalAssetConnection.(type) {
	case *schema.CreateExternalAssetConnectionCreateExternalAssetConnection:
		return externalAssetConnectionFromFields(r.ExternalAssetConnectionFields), nil
	default:
		return nil, fmt.Errorf("CreateExternalAssetConnection: unexpected result type %T", resp.CreateExternalAssetConnection)
	}
}

// UpdateExternalAssetConnection updates an existing external asset connection.
func (c *Client) UpdateExternalAssetConnection(ctx context.Context, conn ExternalAssetConnection) (*ExternalAssetConnection, error) {
	resp, err := schema.UpdateExternalAssetConnection(ctx, c.gqlClient(""),
		conn.ID, conn.Name, conn.Description, conn.SourceConfigYaml, conn.CronString, conn.Timezone,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateExternalAssetConnection: %w", err)
	}

	switch r := resp.UpdateExternalAssetConnection.(type) {
	case *schema.UpdateExternalAssetConnectionUpdateExternalAssetConnection:
		return externalAssetConnectionFromFields(r.ExternalAssetConnectionFields), nil
	default:
		return nil, fmt.Errorf("UpdateExternalAssetConnection: unexpected result type %T", resp.UpdateExternalAssetConnection)
	}
}

// DeleteExternalAssetConnection deletes an external asset connection by ID.
func (c *Client) DeleteExternalAssetConnection(ctx context.Context, id string) error {
	resp, err := schema.DeleteExternalAssetConnection(ctx, c.gqlClient(""), id)
	if err != nil {
		return fmt.Errorf("DeleteExternalAssetConnection: %w", err)
	}

	switch resp.DeleteExternalAssetConnection.(type) {
	case *schema.DeleteExternalAssetConnectionDeleteExternalAssetConnectionDeleteExternalAssetConnectionSuccess:
		return nil
	default:
		return fmt.Errorf("DeleteExternalAssetConnection: unexpected result type %T", resp.DeleteExternalAssetConnection)
	}
}

// GetExternalAssetConnection retrieves an external asset connection by ID.
func (c *Client) GetExternalAssetConnection(ctx context.Context, id string) (*ExternalAssetConnection, error) {
	resp, err := schema.GetExternalAssetConnection(ctx, c.gqlClient(""), id)
	if err != nil {
		return nil, fmt.Errorf("GetExternalAssetConnection: %w", err)
	}

	conn := resp.ExternalAssetConnection
	if conn.Id == "" {
		return nil, fmt.Errorf("GetExternalAssetConnection: connection %q not found", id)
	}
	return externalAssetConnectionFromFields(conn.ExternalAssetConnectionFields), nil
}

// ListExternalAssetConnections returns all external asset connections.
func (c *Client) ListExternalAssetConnections(ctx context.Context) ([]ExternalAssetConnection, error) {
	resp, err := schema.ListExternalAssetConnections(ctx, c.gqlClient(""))
	if err != nil {
		return nil, fmt.Errorf("ListExternalAssetConnections: %w", err)
	}

	result := make([]ExternalAssetConnection, len(resp.ExternalAssetConnections))
	for i, conn := range resp.ExternalAssetConnections {
		result[i] = *externalAssetConnectionFromFields(conn.ExternalAssetConnectionFields)
	}
	return result, nil
}

func externalAssetConnectionFromFields(f schema.ExternalAssetConnectionFields) *ExternalAssetConnection {
	return &ExternalAssetConnection{
		ID:               f.Id,
		Name:             f.Name,
		Description:      f.Description,
		ConnectionType:   string(f.ConnectionType),
		SourceConfigYaml: f.SourceConfigYaml,
		CronString:       f.CronString,
		Timezone:         f.Timezone,
		ScheduleStatus:   string(f.ScheduleStatus),
	}
}
