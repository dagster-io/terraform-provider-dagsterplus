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

resource "dagsterplus_deployment" "test" {
  name = "acc-tf-test"
}

resource "dagsterplus_user" "colton" {
  email = "colton@dagsterlabs.com"
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
    user_id = dagsterplus_user.colton.id
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
    user_id = dagsterplus_user.colton.id
  }
}

resource "dagsterplus_agent_token" "test" {
  name = "acc-tf-agent-token"
}

resource "dagsterplus_user_token" "test" {
  name = "acc-tf-user-token"
}

resource "dagsterplus_alert_policy" "test_deployment" {
  deployment  = dagsterplus_deployment.test.name
  name        = "acc-tf-test-alerts"
  policy_type = "run"
  enabled     = true

  run {
    all_runs   = true
    on_failure = true
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
  metadata_key = "acc_tf_test_metric"
  display_name = "Acc TF Test Metric"
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

resource "dagsterplus_organization_settings" "org" {
  settings_json = "{}"
}

resource "dagsterplus_secret" "db_password" {
  secret_name          = "ACC_TF_DB_PASSWORD"
  secret_value         = "placeholder-value"
  full_deployment_scope = true
}
