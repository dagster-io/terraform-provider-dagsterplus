# Terraform Provider for Dagster+

Manage [Dagster+](https://dagster.cloud) (Dagster Cloud) resources declaratively with Terraform.

## Status

| Entity                     | Resource                                  | Data Source                              | Imports | Status       | [datarootsio](https://github.com/datarootsio/terraform-provider-dagster) |
|----------------------------|-------------------------------------------|------------------------------------------|---------|--------------|-------------------------------------------------------------------------|
| Agent Token                | `dagsterplus_agent_token`                 | `dagsterplus_agent_token`                | Yes     | **Ready**    | - |
| Alert Policies (list)      | —                                         | `dagsterplus_alert_policies`             | —       | **Ready**    | - |
| Alert Policy               | `dagsterplus_alert_policy`                | `dagsterplus_alert_policy`               | Yes     | **Ready**    | - |
| Atlan Integration          | `dagsterplus_atlan_integration`           | —                                        | Yes     | **Ready**    | - |
| Code Location              | `dagsterplus_code_location`               | `dagsterplus_code_location`              | Yes     | **Ready**    | Yes |
| Code Location (document)   | `dagsterplus_code_location_from_document` | —                                        | Yes     | Experimental | Yes |
| Code Locations (list)      | —                                         | `dagsterplus_code_locations`             | —       | **Ready**    | - |
| Configuration Document     | —                                         | `dagsterplus_configuration_document`     | —       | Experimental | Yes |
| Custom Metric              | `dagsterplus_custom_metric`               | `dagsterplus_custom_metric`              | Yes     | **Ready**    | - |
| Deployment                 | `dagsterplus_deployment`                  | `dagsterplus_deployment`                 | Yes     | **Ready**    | Yes |
| Deployment Settings        | `dagsterplus_deployment_settings`         | —                                        | Yes     | **Ready**    | - |
| Deployments (list)         | —                                         | `dagsterplus_deployments`                | —       | **Ready**    | - |
| External Asset Connection  | `dagsterplus_external_asset_connection`   | —                                        | Yes     | Experimental | - |
| GitHub Integration         | `dagsterplus_github_integration`          | —                                        | Yes     | Experimental | - |
| Organization               | —                                         | `dagsterplus_organization`               | —       | **Ready**    | Yes |
| Organization Settings      | `dagsterplus_organization_settings`       | —                                        | Yes     | **Ready**    | - |
| Role                       | `dagsterplus_role`                        | `dagsterplus_role`                       | Yes     | **Ready**    | - |
| Roles (list)               | —                                         | `dagsterplus_roles`                      | —       | **Ready**    | - |
| SCIM Settings              | `dagsterplus_scim_settings`               | —                                        | Yes     | **Ready**    | - |
| Secret                     | `dagsterplus_secret`                      | `dagsterplus_secret`                     | Yes     | **Ready**    | - |
| Service Token              | `dagsterplus_service_token`               | —                                        | Yes     | **Ready**    | - |
| Service User               | `dagsterplus_service_user`                | `dagsterplus_service_user`               | Yes     | **Ready**    | - |
| Team                       | `dagsterplus_team`                        | `dagsterplus_team`                       | Yes     | **Ready**    | Yes |
| Team Deployment Grant      | `dagsterplus_team_deployment_grant`       | —                                        | Yes     | Experimental | Yes |
| Team Membership            | `dagsterplus_team_membership`             | —                                        | Yes     | Experimental | Yes |
| Teams (list)               | —                                         | `dagsterplus_teams`                      | —       | **Ready**    | Yes |
| User                       | `dagsterplus_user`                        | `dagsterplus_user`                       | Yes     | **Ready**    | Yes |
| User Token                 | `dagsterplus_user_token`                  | `dagsterplus_user_token`                 | Yes     | **Ready**    | - |
| Users (list)               | —                                         | `dagsterplus_users`                      | —       | **Ready**    | Yes |
| Version                    | —                                         | `dagsterplus_version`                    | —       | Experimental | Yes |

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) ≥ 1.0
- [Go](https://golang.org/) ≥ 1.21 (to build from source)

## Authentication

The provider requires a Dagster+ API token and your organization name. Both can be
supplied via HCL attributes or environment variables:

| Attribute       | Environment variable         |
|-----------------|------------------------------|
| `organization`  | `DAGSTER_CLOUD_ORGANIZATION` |
| `api_token`     | `DAGSTER_CLOUD_API_TOKEN`    |

Generate a token at **Dagster+ → Account Settings → API Tokens**.

## Provider Configuration

| Attribute      | Type   | Required | Description |
|----------------|--------|----------|-------------|
| `organization` | string | Yes      | Dagster+ org name (subdomain) |
| `api_token`    | string | Yes      | Dagster+ API token (sensitive) |
| `base_url`     | string | No       | Override API base URL |

## Local Development

A working Terraform configuration for manual testing lives in `local/main.tf`. Edit it to exercise the resources you're developing, then use the make targets below.

```bash
# 1. Build the provider binary and write dev.tfrc
make dev-setup

# 2. Set credentials (or create a .env file in the repo root)
export DAGSTER_CLOUD_ORGANIZATION=my-org
export DAGSTER_CLOUD_API_TOKEN=your-token

# 3. Iterate against local/main.tf
make dev-plan
make dev-apply
make dev-destroy
```

`make dev-setup` writes a `dev.tfrc` that points Terraform at the locally-built binary via `dev_overrides`. Both `dev.tfrc` and `.env` are gitignored.

### Running Tests

```bash
# Unit tests (no credentials required)
make test

# Acceptance tests (requires real credentials)
export TF_ACC=1
export DAGSTER_CLOUD_ORGANIZATION=my-org
export DAGSTER_CLOUD_API_TOKEN=your-token
make testacc

# Run a specific resource's acceptance tests
go test ./internal/provider/... -run TestAccAlertPolicy -v
```

## License

Apache 2.0
