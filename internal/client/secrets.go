package client

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// Secret represents a Dagster+ environment secret.
type Secret struct {
	ID                            string
	SecretName                    string
	SecretValue                   string
	FullDeploymentScope           bool
	AllBranchDeploymentsScope     bool
	SpecificBranchDeploymentScope string
	LocalDeploymentScope          bool
	LocationNames                 []string
}

// CreateSecret creates a new secret in the organization.
func (c *Client) CreateSecret(ctx context.Context, s Secret) (*Secret, error) {
	locationNames := make([]string, len(s.LocationNames))
	copy(locationNames, s.LocationNames)

	resp, err := schema.CreateSecret(ctx, c.gqlClient(""),
		s.SecretName, s.SecretValue,
		s.FullDeploymentScope, s.AllBranchDeploymentsScope,
		s.SpecificBranchDeploymentScope, s.LocalDeploymentScope,
		locationNames,
	)
	if err != nil {
		return nil, fmt.Errorf("CreateSecret: %w", err)
	}

	switch r := resp.CreateSecret.(type) {
	case *schema.CreateSecretCreateSecretCreateOrUpdateSecretSuccess:
		return secretFromSecret(r.Secret.Id, r.Secret.SecretName, r.Secret.SecretValue,
			r.Secret.FullDeploymentScope, r.Secret.AllBranchDeploymentsScope,
			r.Secret.SpecificBranchDeploymentScope, r.Secret.LocalDeploymentScope,
			r.Secret.LocationNames), nil
	case *schema.CreateSecretCreateSecretTooManySecretsError:
		return nil, fmt.Errorf("CreateSecret: %s", r.Message)
	case *schema.CreateSecretCreateSecretInvalidSecretInputError:
		return nil, fmt.Errorf("CreateSecret: %s", r.Message)
	case *schema.CreateSecretCreateSecretSecretAlreadyExistsError:
		return nil, fmt.Errorf("CreateSecret: %s", r.Message)
	default:
		return nil, fmt.Errorf("CreateSecret: unexpected result type %T", resp.CreateSecret)
	}
}

// UpdateSecret updates an existing secret.
func (c *Client) UpdateSecret(ctx context.Context, s Secret) (*Secret, error) {
	locationNames := make([]string, len(s.LocationNames))
	copy(locationNames, s.LocationNames)

	resp, err := schema.UpdateSecret(ctx, c.gqlClient(""),
		s.ID, s.SecretName, s.SecretValue,
		s.FullDeploymentScope, s.AllBranchDeploymentsScope,
		s.SpecificBranchDeploymentScope, s.LocalDeploymentScope,
		locationNames,
	)
	if err != nil {
		return nil, fmt.Errorf("UpdateSecret: %w", err)
	}

	switch r := resp.UpdateSecret.(type) {
	case *schema.UpdateSecretUpdateSecretCreateOrUpdateSecretSuccess:
		return secretFromSecret(r.Secret.Id, r.Secret.SecretName, r.Secret.SecretValue,
			r.Secret.FullDeploymentScope, r.Secret.AllBranchDeploymentsScope,
			r.Secret.SpecificBranchDeploymentScope, r.Secret.LocalDeploymentScope,
			r.Secret.LocationNames), nil
	case *schema.UpdateSecretUpdateSecretTooManySecretsError:
		return nil, fmt.Errorf("UpdateSecret: %s", r.Message)
	case *schema.UpdateSecretUpdateSecretInvalidSecretInputError:
		return nil, fmt.Errorf("UpdateSecret: %s", r.Message)
	default:
		return nil, fmt.Errorf("UpdateSecret: unexpected result type %T", resp.UpdateSecret)
	}
}

// DeleteSecret deletes a secret by ID.
func (c *Client) DeleteSecret(ctx context.Context, secretID string) error {
	resp, err := schema.DeleteSecret(ctx, c.gqlClient(""), secretID)
	if err != nil {
		return fmt.Errorf("DeleteSecret: %w", err)
	}

	switch resp.DeleteSecret.(type) {
	case *schema.DeleteSecretDeleteSecretDeleteSecretSuccess:
		return nil
	default:
		return fmt.Errorf("DeleteSecret: unexpected result type %T", resp.DeleteSecret)
	}
}

// GetSecretByID retrieves a secret by its ID.
func (c *Client) GetSecretByID(ctx context.Context, secretID string) (*Secret, error) {
	secrets, err := c.ListSecrets(ctx)
	if err != nil {
		return nil, err
	}
	for i := range secrets {
		if secrets[i].ID == secretID {
			return &secrets[i], nil
		}
	}
	return nil, fmt.Errorf("GetSecretByID: secret %q not found", secretID)
}

// GetSecretByName retrieves a secret by its name.
func (c *Client) GetSecretByName(ctx context.Context, name string) (*Secret, error) {
	resp, err := schema.ListSecrets(ctx, c.gqlClient(""), name)
	if err != nil {
		return nil, fmt.Errorf("GetSecretByName: %w", err)
	}

	switch r := resp.SecretsOrError.(type) {
	case *schema.ListSecretsSecretsOrErrorSecrets:
		for _, s := range r.Secrets {
			if s.SecretName == name {
				return secretFromSecret(s.Id, s.SecretName, s.SecretValue,
					s.FullDeploymentScope, s.AllBranchDeploymentsScope,
					s.SpecificBranchDeploymentScope, s.LocalDeploymentScope,
					s.LocationNames), nil
			}
		}
		return nil, fmt.Errorf("GetSecretByName: secret %q not found", name)
	default:
		return nil, fmt.Errorf("GetSecretByName: unexpected result type %T", resp.SecretsOrError)
	}
}

// ListSecrets returns all secrets in the organization.
func (c *Client) ListSecrets(ctx context.Context) ([]Secret, error) {
	resp, err := schema.ListSecrets(ctx, c.gqlClient(""), "")
	if err != nil {
		return nil, fmt.Errorf("ListSecrets: %w", err)
	}

	switch r := resp.SecretsOrError.(type) {
	case *schema.ListSecretsSecretsOrErrorSecrets:
		result := make([]Secret, len(r.Secrets))
		for i, s := range r.Secrets {
			result[i] = *secretFromSecret(s.Id, s.SecretName, s.SecretValue,
				s.FullDeploymentScope, s.AllBranchDeploymentsScope,
				s.SpecificBranchDeploymentScope, s.LocalDeploymentScope,
				s.LocationNames)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("ListSecrets: unexpected result type %T", resp.SecretsOrError)
	}
}

func secretFromSecret(id, name, value string, fullDeploy, allBranch bool, specificBranch string, local bool, locationNames []string) *Secret {
	locs := make([]string, len(locationNames))
	copy(locs, locationNames)
	return &Secret{
		ID:                            id,
		SecretName:                    name,
		SecretValue:                   value,
		FullDeploymentScope:           fullDeploy,
		AllBranchDeploymentsScope:     allBranch,
		SpecificBranchDeploymentScope: specificBranch,
		LocalDeploymentScope:          local,
		LocationNames:                 locs,
	}
}
