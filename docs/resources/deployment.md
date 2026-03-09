# dagsterplus_deployment

Manages a Dagster+ deployment.

## Example Usage

```hcl
resource "dagsterplus_deployment" "prod" {
  name = "prod"
  type = "PROD"
}
```

## Argument Reference

- `name` (Required, ForceNew) – The deployment name.
- `type` (Required, ForceNew) – `PROD` or `BRANCH`.

## Attributes Reference

- `id` – Same as `name`.
- `created_at` – ISO 8601 timestamp of when the deployment was created.

## Import

Deployments can be imported using the deployment name:

```
terraform import dagsterplus_deployment.prod prod
```
