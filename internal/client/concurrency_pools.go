package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// ConcurrencyPool represents an explicit concurrency limit set on a pool
// (concurrency key) within a deployment.
type ConcurrencyPool struct {
	Name  string
	Limit int
	// IsSet is false when the pool has no explicit limit configured and is
	// instead falling back to the deployment's default pool limit. The API
	// always returns a ConcurrencyKeyInfo for any key, so this distinguishes a
	// pool we manage from one that only exists implicitly via the default.
	IsSet bool
}

// GetConcurrencyPool retrieves the explicit concurrency limit for a pool.
func (c *Client) GetConcurrencyPool(ctx context.Context, deployment, name string) (*ConcurrencyPool, error) {
	resp, err := schema.GetConcurrencyLimit(ctx, c.gqlClient(deployment), name)
	if err != nil {
		return nil, fmt.Errorf("GetConcurrencyPool: %w", err)
	}

	info := resp.Instance.ConcurrencyLimit
	return &ConcurrencyPool{
		Name:  name,
		Limit: info.Limit,
		IsSet: !info.UsingDefaultLimit,
	}, nil
}

// SetConcurrencyPool sets the explicit concurrency limit for a pool.
func (c *Client) SetConcurrencyPool(ctx context.Context, deployment, name string, limit int) error {
	resp, err := schema.SetConcurrencyLimit(ctx, c.gqlClient(deployment), name, limit)
	if err != nil {
		return fmt.Errorf("SetConcurrencyPool: %w", err)
	}
	if !resp.SetConcurrencyLimit {
		return fmt.Errorf("SetConcurrencyPool: API reported failure setting limit for pool %q", name)
	}
	return nil
}

// DeleteConcurrencyPool removes the explicit concurrency limit for a pool,
// reverting it to the deployment's default pool limit.
func (c *Client) DeleteConcurrencyPool(ctx context.Context, deployment, name string) error {
	resp, err := schema.DeleteConcurrencyLimit(ctx, c.gqlClient(deployment), name)
	if err != nil {
		return fmt.Errorf("DeleteConcurrencyPool: %w", err)
	}
	if !resp.DeleteConcurrencyLimit {
		return fmt.Errorf("DeleteConcurrencyPool: API reported failure deleting limit for pool %q", name)
	}
	return nil
}
