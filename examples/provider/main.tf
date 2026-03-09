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
  # api_token is read from DAGSTER_CLOUD_API_TOKEN environment variable
}
