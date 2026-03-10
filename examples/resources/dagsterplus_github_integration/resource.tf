# Select the active GitHub App installation for the organization.
# Pre-requisite: the GitHub App must already be installed via the GitHub OAuth
# flow in the Dagster+ UI before using this resource.
resource "dagsterplus_github_integration" "example" {
  account_name = "my-github-org"
}
