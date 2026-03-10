# Migrating from datarootsio/dagster to dagster-io/dagsterplus

This guide covers migrating an existing Terraform configuration from the community
[`datarootsio/dagster`](https://registry.terraform.io/providers/datarootsio/dagster) provider
to the official [`dagster-io/dagsterplus`](https://registry.terraform.io/providers/dagster-io/dagsterplus) provider.

Terraform has no built-in mechanism to migrate state across provider namespaces. The migration
is manual: you rewrite your configuration, then use `terraform state mv` or `terraform import`
to adopt the existing remote resources into the new provider's state entries — **without
destroying or recreating anything in Dagster+**.

---

## Before you start

1. **Back up your state.** For remote state (S3, Terraform Cloud, etc.) take a snapshot. For
   local state, copy `terraform.tfstate` somewhere safe.
2. **Credentials.** Both providers use the same Dagster+ API token. Make sure
   `DAGSTER_CLOUD_ORGANIZATION` and `DAGSTER_CLOUD_API_TOKEN` are set (or configured in the
   provider block).
3. **Work one resource type at a time.** Run `terraform plan` after each section to confirm
   zero drift before moving on.

---

## Provider configuration

### Before (datarootsio)

```hcl
terraform {
  required_providers {
    dagster = {
      source  = "datarootsio/dagster"
      version = "~> 0.1"
    }
  }
}

provider "dagster" {
  organization = "my-org"
  api_token    = var.dagster_token
  deployment   = "prod"   # sets a default deployment for all resources
}
```

### After (dagster-io)

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
  api_token    = var.dagster_token
  # No default deployment — specify deployment per resource where required
}
```

Run `terraform init -upgrade` after updating `required_providers`.

---

## Step-by-step migration

Migrate in dependency order: deployments first, then users, then teams, then code locations.

### 1. Deployments

`dagster_deployment` → `dagsterplus_deployment`

**Schema differences:**
- `settings_document` (JSON blob in datarootsio) does not exist in dagsterplus. Use the
  separate `dagsterplus_deployment_settings` resource instead.

**Config rewrite:**

```hcl
# Before
resource "dagster_deployment" "prod" {
  name = "prod"
  type = "SERVERLESS"
}

# After
resource "dagsterplus_deployment" "prod" {
  name = "prod"
  type = "SERVERLESS"
}
```

**State migration** (run for each deployment):

```bash
terraform state mv \
  'dagster_deployment.prod' \
  'dagsterplus_deployment.prod'
```

Verify:

```bash
terraform plan   # should show no changes for this resource
```

If you had a `settings_document` on the deployment, create a separate
`dagsterplus_deployment_settings` resource and import it:

```bash
terraform import dagsterplus_deployment_settings.prod prod
```

### 2. Users

`dagster_user` → `dagsterplus_user`

**Schema differences:** None significant. datarootsio used numeric IDs internally; dagsterplus
uses string IDs. After `terraform state mv` a plan may show an ID value change — this is
cosmetic and resolves on the next apply.

**Config rewrite:**

```hcl
# Before
resource "dagster_user" "alice" {
  email = "alice@example.com"
}

# After
resource "dagsterplus_user" "alice" {
  email = "alice@example.com"
}
```

**State migration:**

```bash
terraform state mv \
  'dagster_user.alice' \
  'dagsterplus_user.alice'
```

### 3. Teams

datarootsio splits team configuration across three separate resources:
`dagster_team`, `dagster_team_membership`, and `dagster_team_deployment_grant`.
dagsterplus consolidates all three into a single `dagsterplus_team` resource using
`member {}` and `deployment_grant {}` blocks.

**Config rewrite:**

```hcl
# Before — three separate resources
resource "dagster_team" "platform" {
  name = "platform"
}

resource "dagster_team_membership" "alice_platform" {
  team_id = dagster_team.platform.id
  user_id = dagster_user.alice.id
}

resource "dagster_team_deployment_grant" "platform_prod" {
  team_id        = dagster_team.platform.id
  deployment     = "prod"
  deployment_access = "EDITOR"
}

# After — single consolidated resource
resource "dagsterplus_team" "platform" {
  name = "platform"

  member {
    user_id = dagsterplus_user.alice.id
  }

  deployment_grant {
    deployment = "prod"
    grant      = "EDITOR"
  }
}
```

**State migration:**

Because the consolidated resource includes membership and grant data that the old
`dagster_team` resource did not, `terraform state mv` alone will not produce a clean plan.
The safest approach is to remove the old entries from state and import the consolidated team:

```bash
# Remove the three old state entries
terraform state rm 'dagster_team.platform'
terraform state rm 'dagster_team_membership.alice_platform'
terraform state rm 'dagster_team_deployment_grant.platform_prod'

# Import the team (the remote resource is unchanged)
terraform import dagsterplus_team.platform platform
```

Run `terraform plan` — it should show no changes if your new config matches the live state.

### 4. Code locations

`dagster_code_location` → `dagsterplus_code_location`

**Schema differences:**

| datarootsio attribute | dagsterplus attribute | Notes |
|---|---|---|
| `deployment` | `deployment` | Unchanged |
| `code_source.python_file` | `code_source { python_file = ... }` | Same, nested block |
| `code_source.package_name` | `code_source { package_name = ... }` | Same, nested block |
| `code_source.module_name` | `code_source { module_name = ... }` | Same, nested block |
| `working_directory` | `working_directory` | Unchanged |
| `executable_path` | `executable_path` | Unchanged |

The attribute names are identical. The provider is inferred from context rather than from the
provider-level `deployment` default.

**Config rewrite:**

```hcl
# Before
resource "dagster_code_location" "my_pipeline" {
  deployment = "prod"
  name       = "my-pipeline"
  image      = "ghcr.io/my-org/my-pipeline:latest"

  code_source {
    python_file = "repo.py"
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}

# After
resource "dagsterplus_code_location" "my_pipeline" {
  deployment = "prod"
  name       = "my-pipeline"
  image      = "ghcr.io/my-org/my-pipeline:latest"

  code_source {
    python_file = "repo.py"
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}
```

**State migration** (run for each code location):

```bash
terraform state mv \
  'dagster_code_location.my_pipeline' \
  'dagsterplus_code_location.my_pipeline'
```

Verify:

```bash
terraform plan   # should show no changes for this resource
```

#### `dagster_code_location_from_document`

This resource (YAML-based code location document) has no equivalent in dagsterplus.
Remove it from your state and rewrite the config as a `dagsterplus_code_location` resource,
then create it fresh:

```bash
terraform state rm 'dagster_code_location_from_document.my_location'
# Update config to use dagsterplus_code_location, then:
terraform apply
```

---

## Data sources

Data sources hold no state — migration is a config rewrite only.

| datarootsio data source | dagsterplus data source | Notes |
|---|---|---|
| `data.dagster_deployment` | `data.dagsterplus_deployment` | Lookup by `name` |
| `data.dagster_user` | `data.dagsterplus_user` | Lookup by `email` |
| `data.dagster_users` | `data.dagsterplus_users` | No `email_regex` filter in dagsterplus — filter in locals |
| `data.dagster_team` | `data.dagsterplus_team` | Lookup by `name` |
| `data.dagster_teams` | `data.dagsterplus_teams` | No `regex_filter` in dagsterplus — filter in locals |
| `data.dagster_organization` | `data.dagsterplus_organization` | |
| `data.dagster_current_deployment` | No equivalent | Use the deployment name string directly |
| `data.dagster_configuration_document` | No equivalent | Not needed — dagsterplus does not use YAML config documents |
| `data.dagster_version` | No equivalent | Not provided |

Example rewrite for a user lookup:

```hcl
# Before
data "dagster_user" "alice" {
  email = "alice@example.com"
}

# After
data "dagsterplus_user" "alice" {
  email = "alice@example.com"
}
```

---

## Final verification

After completing all sections, run a full plan:

```bash
terraform plan
```

A successful migration shows **no changes** (no creates, updates, or destroys). If you see
unexpected diffs:

- **ID type drift** (numeric vs string): run `terraform apply` once to converge; no remote
  change will occur.
- **Unexpected attribute drift**: compare the live API state (use `terraform show`) against
  your config. Adjust the config to match, or use `terraform apply` to push the desired value.

---

## Resources with no equivalent

| datarootsio resource | Recommendation |
|---|---|
| `dagster_code_location_from_document` | Rewrite as `dagsterplus_code_location` (see above) |
| `data.dagster_current_deployment` | Replace with a literal string or a `var.deployment_name` variable |
| `data.dagster_configuration_document` | Remove — dagsterplus does not use YAML documents |
| `data.dagster_version` | Remove — provider version is not exposed as a data source |
