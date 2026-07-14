# Terraform Provider for Dagster+

> **Early Development** — this provider is a work in progress.

Manage [Dagster+](https://dagster.cloud) (Dagster Cloud) resources declaratively with Terraform.

## Installation

```hcl
terraform {
  required_providers {
    dagsterplus = {
      source  = "dagster-io/dagsterplus"
    }
  }
}
```

The latest version is available on the [Terraform Registry](https://registry.terraform.io/providers/dagster-io/dagsterplus/latest).

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

## Usage

### Managing resources

You can manage resources using the `terraform apply` command. For example, to create a new deployment named `production`:

```hcl
resource "dagsterplus_deployment" "production" {
  name = "production"
}
```

### Data sources

You can use data sources to retrieve information about existing resources. For example, to list all deployments in your organization:

```hcl
data "dagsterplus_deployments" "all" {}

output "deployments" {
  value = data.dagsterplus_deployments.all
}
```

### Importing existing resources

You can import existing resources into your Terraform state using the `terraform import` command. For example, to import an existing deployment:

```hcl
resource "dagsterplus_deployment" "production" {
  name = "production"
}
```

More examples are available in the [`examples`](./examples/) directory and in the provider documentation on the Terraform Registry.

## Status

| Entity | Resource | Data Source | Imports | Status |
|--------|----------|-------------|---------|--------|
| Deployment | `dagsterplus_deployment` | `dagsterplus_deployment` / `dagsterplus_deployments` | Yes | Experimental |
| Code Location | `dagsterplus_code_location` | `dagsterplus_code_location` / `dagsterplus_code_locations` | Yes | Experimental |
| Team | `dagsterplus_team` | `dagsterplus_team` / `dagsterplus_teams` | Yes | Experimental |
| Team Membership | `dagsterplus_team_membership` | — | Yes | Experimental |
| Team Deployment Grant | `dagsterplus_team_deployment_grant` | — | Yes | Experimental |
| User | `dagsterplus_user` | `dagsterplus_user` / `dagsterplus_users` | Yes | Experimental |
| Alert Policy | `dagsterplus_alert_policy` | `dagsterplus_alert_policy` / `dagsterplus_alert_policies` | Yes | Experimental |
| Agent Token | `dagsterplus_agent_token` | `dagsterplus_agent_token` | Yes | Experimental |
| Agent Token Deployment Grant | `dagsterplus_agent_token_deployment_grant` | — | Yes | Experimental |
| User Token | `dagsterplus_user_token` | `dagsterplus_user_token` | Yes | Experimental |
| Role | `dagsterplus_role` | `dagsterplus_role` / `dagsterplus_roles` | Yes | Experimental |
| SCIM Settings | `dagsterplus_scim_settings` | — | Yes | Experimental |
| Atlan Integration | `dagsterplus_atlan_integration` | — | Yes | Experimental |
| GitHub Integration | `dagsterplus_github_integration` | — | Yes | Experimental |
| Secret | `dagsterplus_secret` | `dagsterplus_secret` | Yes | Experimental |
| Deployment Settings | `dagsterplus_deployment_settings` | — | Yes | Experimental |
| Concurrency Pool | `dagsterplus_concurrency_pool` | — | Yes | Experimental |
| Service User | `dagsterplus_service_user` | `dagsterplus_service_user` | Yes | Experimental |
| Service User Deployment Grant | `dagsterplus_service_user_deployment_grant` | — | Yes | Experimental |
| Service Token | `dagsterplus_service_token` | — | Yes | Experimental |
| Organization Settings | `dagsterplus_organization_settings` | `dagsterplus_organization` | Yes | Experimental |
| Custom Metric | `dagsterplus_custom_metric` | `dagsterplus_custom_metric` | Yes | Experimental |
| External Asset Connection | `dagsterplus_external_asset_connection` | — | Yes | Experimental |

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute to this provider.

## License

This provider is distributed under the [Mozilla Public License, Version 2.0](LICENSE).
