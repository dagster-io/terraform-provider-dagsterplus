# dagsterplus_team

Manages a Dagster+ team and its deployment-level permissions.

## Example Usage

```hcl
resource "dagsterplus_team" "data_engineering" {
  name = "data-engineering"

  deployment_permission {
    deployment_name = "prod"
    role            = "EDITOR"
  }

  deployment_permission {
    deployment_name = "staging"
    role            = "ADMIN"
  }
}
```

## Argument Reference

- `name` (Required, ForceNew) – The team name.
- `deployment_permission` (Optional, repeatable block):
  - `deployment_name` (Required) – The deployment to grant access to.
  - `role` (Required) – One of `VIEWER`, `LAUNCHER`, `EDITOR`, `ADMIN`.

## Attributes Reference

- `id` – The team ID assigned by Dagster+.

## Import

Teams can be imported using the team ID (visible in the Dagster+ UI or API):

```
terraform import dagsterplus_team.data_engineering <team-id>
```
