# Migrating from datarootsio/dagster to dagster-io/dagsterplus

Both providers talk to the same Dagster+ API. No remote resources are destroyed or recreated
during migration. You are only updating which provider manages each resource in Terraform state.

Two strategies are available. Choose the one that fits your situation.

---

## Credentials

Both providers use the same environment variables:

```
DAGSTER_CLOUD_ORGANIZATION=<your-org>
DAGSTER_CLOUD_API_TOKEN=<your-token>
```

---

## Strategy 1 — Clean cut (recommended for small configs)

Best when you want to drop the `datarootsio` provider in a single pass.

**Steps (illustrated with `dagsterplus_deployment`):**

### 1. Add the new provider to `required_providers`

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
}
```

Run `terraform init`.

### 2. Write the new resource block alongside the old one

```hcl
# Keep this until migration is complete
resource "dagster_deployment" "prod" {
  name = "prod"
  type = "SERVERLESS"
}

# New resource to take over
resource "dagsterplus_deployment" "prod" {
  name = "prod"
  type = "SERVERLESS"
}
```

### 3. Remove the old resource from state

```bash
terraform state rm 'dagster_deployment.prod'
```

### 4. Import the existing resource into the new provider

```bash
terraform import dagsterplus_deployment.prod prod
```

The import ID is the deployment name string.

### 5. Delete the old resource block and provider

Remove `dagster_deployment.prod` from your config and remove `datarootsio/dagster` from
`required_providers`.

### 6. Verify

```bash
terraform plan   # should show no changes
```

---

## Strategy 2 — Side-by-side (indefinite coexistence)

Best when you want to keep existing resources where they are and simply use `dagsterplus`
for new resources going forward, including new resources of types that also exist in the
`datarootsio` provider (e.g. you can add a new `dagsterplus_deployment` while existing
`dagster_deployment` resources stay managed by datarootsio indefinitely).

Both providers talk to the same API, so there is no conflict. You are only choosing which
provider manages each Terraform resource object; the underlying Dagster+ resource is the same.

### 1. Add the new provider alongside the existing one

```hcl
terraform {
  required_providers {
    dagster = {
      source  = "datarootsio/dagster"
      version = "~> 0.1"
    }
    dagsterplus = {
      source  = "dagster-io/dagsterplus"
      version = "~> 0.1"
    }
  }
}

provider "dagster" {
  organization = "my-org"
  api_token    = var.dagster_token
  deployment   = "prod"
}

provider "dagsterplus" {
  organization = "my-org"
  api_token    = var.dagster_token
}
```

Run `terraform init`.

### 2. Use dagsterplus for new resources

Write new resources using `dagsterplus_*` types. Existing `dagster_*` resources are untouched.

```hcl
# Existing resource — stays with datarootsio, no changes needed
resource "dagster_deployment" "prod" {
  name = "prod"
  type = "SERVERLESS"
}

# New resource — managed by dagster-io/dagsterplus from day one
resource "dagsterplus_deployment" "staging" {
  name = "staging"
  type = "SERVERLESS"
}
```

### 3. Optionally migrate existing resources

If you later want to move an existing datarootsio resource to dagsterplus, use the clean cut
steps from Strategy 1 (`terraform state rm` + `terraform import`) for that resource.
You can do this for any subset of resources at any time — there is no requirement to migrate
everything.
