# Terraform Provider for Dagster+

Manage [Dagster+](https://dagster.cloud) (Dagster Cloud) resources declaratively with Terraform.

## Features

- **Deployments** – create and destroy production/branch deployments
- **Code Locations** – register container-based code locations within a deployment
- **Teams & Permissions** – manage teams and grant deployment-level roles

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) ≥ 1.0
- [Go](https://golang.org/) ≥ 1.21 (to build from source)

## Authentication

The provider requires a Dagster+ API token and your organization name. Both can be
supplied via HCL attributes or environment variables:

| Attribute       | Environment variable            |
|-----------------|---------------------------------|
| `organization`  | `DAGSTER_CLOUD_ORGANIZATION`    |
| `api_token`     | `DAGSTER_CLOUD_API_TOKEN`       |

Generate a token at **Dagster+ → Account Settings → API Tokens**.

## Quick Start

```hcl
terraform {
  required_providers {
    dagsterplus = {
      source  = "dagster-io/dagsterplus"
      version = "~> 0.1"
    }
  }
}

provider "dagsterplus" {
  organization = "my-org"
  # api_token read from DAGSTER_CLOUD_API_TOKEN
}

resource "dagsterplus_deployment" "prod" {
  name = "prod"
  type = "PROD"
}

resource "dagsterplus_code_location" "pipeline" {
  deployment_name = dagsterplus_deployment.prod.name
  name            = "my-pipeline"
  image           = "ghcr.io/my-org/my-pipeline:latest"

  code_source {
    python_file = "repo.py"
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}

resource "dagsterplus_team" "data_eng" {
  name = "data-engineering"

  deployment_permission {
    deployment_name = dagsterplus_deployment.prod.name
    role            = "EDITOR"
  }
}
```

## Provider Configuration

| Attribute      | Type   | Required | Description |
|----------------|--------|----------|-------------|
| `organization` | string | Yes      | Dagster+ org name (subdomain) |
| `api_token`    | string | Yes      | Dagster+ API token (sensitive) |
| `base_url`     | string | No       | Override API base URL |

## Supported Operations

| Resource                    | Create | Read | Update | Delete | Import | Data Source |
|-----------------------------|:------:|:----:|:------:|:------:|:------:|:-----------:|
| `dagsterplus_deployment`    | ✓      | ✓    | —¹     | ✓      | ✓      | ✓           |
| `dagsterplus_code_location` | ✓      | ✓    | ✓      | ✓      | ✓      | —           |
| `dagsterplus_team`          | ✓      | ✓    | ✓      | ✓      | ✓      | —           |
| `dagsterplus_user`          | ✓      | ✓    | ✓      | ✓      | ✓      | —           |

¹ All deployment attributes (`name`, `type`) are ForceNew — any change destroys and recreates the deployment.

## Resources

### `dagsterplus_deployment`

| Attribute    | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `name`       | string | Yes      | Deployment name (forces new) |
| `type`       | string | Yes      | `PROD` or `BRANCH` (forces new) |
| `id`         | string | Computed | Same as `name` |
| `created_at` | string | Computed | ISO 8601 creation timestamp |

**Import:** `terraform import dagsterplus_deployment.prod prod`

### `dagsterplus_code_location`

| Attribute           | Type   | Required | Description |
|---------------------|--------|----------|-------------|
| `deployment_name`   | string | Yes      | Target deployment (forces new) |
| `name`              | string | Yes      | Location name (forces new) |
| `image`             | string | Yes      | Docker image reference |
| `code_source`       | block  | Yes      | Where Dagster finds the code |
| `working_directory` | string | No       | Working directory in container |
| `executable_path`   | string | No       | Python executable path |

`code_source` block attributes (at least one required):

| Attribute      | Description |
|----------------|-------------|
| `python_file`  | Path to a Python file |
| `package_name` | Python package name |
| `module_name`  | Python module name |

**Import:** `terraform import dagsterplus_code_location.pipeline prod/my-pipeline`

### `dagsterplus_team`

| Attribute              | Type   | Required | Description |
|------------------------|--------|----------|-------------|
| `name`                 | string | Yes      | Team name (forces new) |
| `id`                   | string | Computed | Team ID |
| `deployment_permission`| block  | No       | Repeated; grants a role on a deployment |

`deployment_permission` block:

| Attribute         | Description |
|-------------------|-------------|
| `deployment_name` | Deployment to grant access to |
| `role`            | `VIEWER`, `LAUNCHER`, `EDITOR`, or `ADMIN` |

**Import:** `terraform import dagsterplus_team.data_eng <team-id>`

### `dagsterplus_user`

| Attribute | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `email`   | string | Yes      | Email address of the user to invite (forces new) |
| `role`    | string | Yes      | Org-level role: `VIEWER`, `EDITOR`, `ADMIN`, or `OWNER` |
| `id`      | string | Computed | User ID assigned by Dagster+ |
| `name`    | string | Computed | Display name (set once the user accepts the invite) |

**Import:** `terraform import dagsterplus_user.alice <user-id>`

## Data Sources

### `dagsterplus_deployment`

```hcl
data "dagsterplus_deployment" "existing" {
  name = "prod"
}
```

## Local Development

```bash
# Build
make build

# Install to local Terraform plugins directory
make install

# Add dev_overrides to ~/.terraformrc
cat > ~/.terraformrc <<EOF
provider_installation {
  dev_overrides {
    "dagster-io/dagsterplus" = "$HOME/.terraform.d/plugins/registry.terraform.io/dagster-io/dagsterplus/0.1.0/$(go env GOOS)_$(go env GOARCH)"
  }
  direct {}
}
EOF

# Run acceptance tests (requires real credentials)
export DAGSTER_CLOUD_API_TOKEN=your-token
export DAGSTER_CLOUD_ORGANIZATION=your-org
make testacc
```

## License

Apache 2.0
