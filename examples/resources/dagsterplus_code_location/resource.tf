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

# Using a package name instead of a Python file
resource "dagsterplus_code_location" "my_package" {
  deployment = "prod"
  name       = "my-package"
  image      = "ghcr.io/my-org/my-package:latest"

  code_source {
    package_name = "my_dagster_package"
  }
}

# With Kubernetes container context
resource "dagsterplus_code_location" "with_k8s" {
  deployment = "prod"
  name       = "my-k8s-location"
  image      = "ghcr.io/my-org/my-pipeline:latest"

  code_source {
    module_name = "my_dagster_module"
  }

  container_context = jsonencode({
    k8s = {
      namespace    = "dagster"
      env_vars     = ["ENV=production"]
      env_secrets  = ["my-secret"]
      resources = {
        requests = { memory = "512Mi", cpu = "250m" }
        limits   = { memory = "1Gi", cpu = "500m" }
      }
    }
  })
}
