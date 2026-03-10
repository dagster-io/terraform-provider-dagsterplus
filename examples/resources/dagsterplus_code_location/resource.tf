resource "dagsterplus_code_location" "my_pipeline" {
  deployment = "prod"
  name            = "my-pipeline"
  image           = "ghcr.io/my-org/my-pipeline:latest"

  code_source {
    python_file = "repo.py"
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}

# Using a package name instead of a Python file
resource "dagsterplus_code_location" "my_package" {
  deployment = "prod"
  name            = "my-package"
  image           = "ghcr.io/my-org/my-package:latest"

  code_source {
    package_name = "my_dagster_package"
  }
}
