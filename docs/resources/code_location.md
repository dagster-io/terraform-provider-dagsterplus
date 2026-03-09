# dagsterplus_code_location

Manages a Dagster+ code location within a deployment.

## Example Usage

```hcl
resource "dagsterplus_code_location" "my_pipeline" {
  deployment_name = dagsterplus_deployment.prod.name
  name            = "my-pipeline"
  image           = "ghcr.io/my-org/my-pipeline:latest"

  code_source {
    python_file = "repo.py"
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}
```

## Argument Reference

- `deployment_name` (Required, ForceNew) – The deployment this code location belongs to.
- `name` (Required, ForceNew) – The code location name.
- `image` (Required) – Docker image reference.
- `code_source` (Required block) – Where Dagster finds the code. Exactly one of:
  - `python_file` – Path to a Python file.
  - `package_name` – Python package name.
  - `module_name` – Python module name.
- `working_directory` (Optional) – Working directory inside the container.
- `executable_path` (Optional) – Python executable path inside the container.

## Attributes Reference

- `id` – `{deployment_name}/{name}`.

## Import

Code locations can be imported using `{deployment_name}/{location_name}`:

```
terraform import dagsterplus_code_location.my_pipeline prod/my-pipeline
```
