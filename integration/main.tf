terraform {
  required_providers {
    dagsterplus = {
      source = "dagster-io/dagsterplus"
    }
  }

  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "dagsterplus" {}

variable "test_user_email" {
  type        = string
  description = "Email of a non-owner org member to use in the integration test. Set via DAGSTER_CLOUD_TEST_USER_EMAIL or TF_VAR_test_user_email."
  default     = "dennis@dagsterlabs.com"
}

resource "dagsterplus_deployment" "test" {
  name = "acc-tf-test"
}

resource "dagsterplus_user" "dennis" {
  email = var.test_user_email
}

resource "dagsterplus_role" "observability" {
  name      = "acc-tf-observability"
  role_type = "deployment"

  permissions = ["edit_alerts", "edit_all_catalog_views"]
}

resource "dagsterplus_role" "org_admin" {
  name      = "acc-tf-org-admin"
  role_type = "organization"

  permissions = ["edit_users_and_teams", "edit_custom_roles", "read_audit_log"]
}

resource "dagsterplus_team" "data_engineering" {
  name = "acc-tf-data-engineering"

  organization_grant {
    custom_role_id = dagsterplus_role.org_admin.id
  }

  member {
    user_id = dagsterplus_user.dennis.id
  }
}

resource "dagsterplus_team" "data_engineering_2" {
  name = "acc-tf-data-engineering-2"

  deployment_grant {
    deployment     = "prod"
    custom_role_id = dagsterplus_role.observability.id
  }

  all_branch_deployments_grant {
    custom_role_id = dagsterplus_role.observability.id
  }

  member {
    user_id = dagsterplus_user.dennis.id
  }
}

resource "dagsterplus_agent_token" "test" {
  name = "acc-tf-agent-token"
}

resource "dagsterplus_user_token" "test" {
  name = "acc-tf-user-token"
}

resource "dagsterplus_code_location" "test" {
  deployment = dagsterplus_deployment.test.name
  name       = "acc-tf-code-location"
  image      = "ghcr.io/example/repo:v1"

  code_source {
    python_file = "repo.py"
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}

resource "dagsterplus_deployment_settings" "test" {
  deployment    = dagsterplus_deployment.test.name
  settings_json = jsonencode({ run_queue = { max_concurrent_runs = 5 } })
}

# Standalone team used to test team_deployment_grant and team_membership
# as separate resources (no inline grants or members).
resource "dagsterplus_team" "grants_only" {
  name = "acc-tf-grants-only"
}

resource "dagsterplus_team_deployment_grant" "test" {
  team_id        = dagsterplus_team.grants_only.id
  deployment     = dagsterplus_deployment.test.name
  custom_role_id = dagsterplus_role.observability.id
}

resource "dagsterplus_team_membership" "dennis" {
  team_id = dagsterplus_team.grants_only.id
  user_id = dagsterplus_user.dennis.id
}

resource "dagsterplus_alert_policy" "test_deployment" {
  deployment  = dagsterplus_deployment.test.name
  name        = "acc-tf-test-alerts"
  policy_type = "run"
  enabled     = true

  run {
    # Filter to the specific code location managed by this config.
    code_locations = [dagsterplus_code_location.test.name]
    on_failure     = true
  }

  notification_service {
    type            = "email"
    email_addresses = ["dennis@dagsterlabs.com"]
  }
}

resource "dagsterplus_alert_policy" "code_location" {
  deployment  = dagsterplus_deployment.test.name
  name        = "acc-tf-code-location-alerts"
  policy_type = "code_location"
  enabled     = true

  code_location {
    location_name = dagsterplus_code_location.test.name
  }

  notification_service {
    type            = "email"
    email_addresses = ["dennis@dagsterlabs.com"]
  }
}

resource "dagsterplus_alert_policy" "asset_health" {
  deployment  = "prod"
  name        = "asset-health-alerts"
  policy_type = "asset"
  enabled     = true

  asset {
    all_assets      = true
    specific_events = ["materialization_success", "materialization_failure"]
  }

  notification_service {
    type            = "email"
    email_addresses = ["dennis@dagsterlabs.com"]
  }
}

resource "dagsterplus_custom_metric" "test" {
  metadata_key = "acc_tf_integration_metric"
  display_name = "Acc TF Integration Metric"
  description  = "A custom metric managed by Terraform"
}

resource "dagsterplus_service_user" "ci_bot" {
  name        = "acc-tf-ci-bot"
  description = "CI/CD service user managed by Terraform"
}

resource "dagsterplus_service_token" "ci_bot_token" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  description     = "Primary token for acc-tf-ci-bot"
}

resource "dagsterplus_service_user_deployment_grant" "ci_bot_test" {
  service_user_id = dagsterplus_service_user.ci_bot.id
  deployment      = dagsterplus_deployment.test.name
  grant           = "LAUNCHER"
}

resource "dagsterplus_organization_settings" "org" {
  settings_json = "{}"
}

resource "dagsterplus_secret" "db_password" {
  secret_name           = "ACC_TF_DB_PASSWORD"
  secret_value          = "placeholder-value"
  full_deployment_scope = true
}

# ---------------------------------------------------------------------------
# Data sources — read back every resource created above to exercise the
# data source read path independently from the resource create path.
# ---------------------------------------------------------------------------

data "dagsterplus_user" "dennis" {
  email      = dagsterplus_user.dennis.email
  depends_on = [dagsterplus_user.dennis]
}

data "dagsterplus_deployment" "test" {
  name       = dagsterplus_deployment.test.name
  depends_on = [dagsterplus_deployment.test]
}

data "dagsterplus_role" "observability" {
  name       = dagsterplus_role.observability.name
  depends_on = [dagsterplus_role.observability]
}

data "dagsterplus_role" "org_admin" {
  name       = dagsterplus_role.org_admin.name
  depends_on = [dagsterplus_role.org_admin]
}

data "dagsterplus_team" "data_engineering" {
  name       = dagsterplus_team.data_engineering.name
  depends_on = [dagsterplus_team.data_engineering]
}

data "dagsterplus_agent_token" "test" {
  name       = dagsterplus_agent_token.test.name
  depends_on = [dagsterplus_agent_token.test]
}

data "dagsterplus_user_token" "test" {
  name       = dagsterplus_user_token.test.name
  user_id    = dagsterplus_user_token.test.user_id
  depends_on = [dagsterplus_user_token.test]
}

data "dagsterplus_code_location" "test" {
  deployment = dagsterplus_code_location.test.deployment
  name       = dagsterplus_code_location.test.name
  depends_on = [dagsterplus_code_location.test]
}

data "dagsterplus_alert_policy" "test_deployment" {
  deployment  = dagsterplus_alert_policy.test_deployment.deployment
  name        = dagsterplus_alert_policy.test_deployment.name
  policy_type = "run"
  depends_on  = [dagsterplus_alert_policy.test_deployment]
}

data "dagsterplus_alert_policy" "asset_health" {
  deployment  = dagsterplus_alert_policy.asset_health.deployment
  name        = dagsterplus_alert_policy.asset_health.name
  policy_type = "asset"
  depends_on  = [dagsterplus_alert_policy.asset_health]
}

data "dagsterplus_alert_policy" "code_location" {
  deployment  = dagsterplus_alert_policy.code_location.deployment
  name        = dagsterplus_alert_policy.code_location.name
  policy_type = "code_location"
  depends_on  = [dagsterplus_alert_policy.code_location]
}

data "dagsterplus_custom_metric" "test" {
  metadata_key = dagsterplus_custom_metric.test.metadata_key
  depends_on   = [dagsterplus_custom_metric.test]
}

data "dagsterplus_service_user" "ci_bot" {
  name       = dagsterplus_service_user.ci_bot.name
  depends_on = [dagsterplus_service_user.ci_bot]
}

# Organization (singleton — no lookup key needed)
data "dagsterplus_organization" "org" {}

# Secret
data "dagsterplus_secret" "db_password" {
  secret_name = dagsterplus_secret.db_password.secret_name
  depends_on  = [dagsterplus_secret.db_password]
}

# List data sources — verify the list read path returns at least the resources
# created above.

data "dagsterplus_deployments" "all" {
  depends_on = [dagsterplus_deployment.test]
}

data "dagsterplus_teams" "all" {
  depends_on = [
    dagsterplus_team.data_engineering,
    dagsterplus_team.data_engineering_2,
    dagsterplus_team.grants_only,
  ]
}

data "dagsterplus_roles" "all" {
  depends_on = [
    dagsterplus_role.observability,
    dagsterplus_role.org_admin,
  ]
}

data "dagsterplus_users" "all" {
  depends_on = [dagsterplus_user.dennis]
}

data "dagsterplus_code_locations" "test_deployment" {
  deployment = dagsterplus_deployment.test.name
  depends_on = [dagsterplus_code_location.test]
}

data "dagsterplus_alert_policies" "test_deployment" {
  deployment = dagsterplus_deployment.test.name
  depends_on = [
    dagsterplus_alert_policy.test_deployment,
    dagsterplus_alert_policy.code_location,
  ]
}
