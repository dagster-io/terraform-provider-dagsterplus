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

# ---------------------------------------------------------------------------
# Dependencies for the grant tests.
# ---------------------------------------------------------------------------

resource "dagsterplus_deployment" "test" {
  name = "acc-tf-test"
}

resource "dagsterplus_user" "dennis" {
  email = var.test_user_email
}

resource "dagsterplus_role" "observability" {
  name        = "acc-tf-observability"
  role_type   = "deployment"
  permissions = ["edit_alerts", "edit_all_catalog_views"]
}

resource "dagsterplus_role" "org_admin" {
  name        = "acc-tf-org-admin"
  role_type   = "organization"
  permissions = ["edit_users_and_teams", "edit_custom_roles", "read_audit_log"]
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

# ---------------------------------------------------------------------------
# Team A: inline grant blocks — all 4 scopes.
# Exercises the inline lifecycle code paths on dagsterplus_team.
# ---------------------------------------------------------------------------

resource "dagsterplus_team" "inline" {
  name = "acc-tf-team-inline"

  organization_grant {
    custom_role_id = dagsterplus_role.org_admin.id
  }

  deployment_grant {
    deployment     = dagsterplus_deployment.test.name
    custom_role_id = dagsterplus_role.observability.id
  }

  all_branch_deployments_grant {
    grant = "LAUNCHER"
  }

  branch_deployments_grant {
    parent_deployment = dagsterplus_deployment.test.name
    grant             = "EDITOR"
  }

  member {
    user_id = dagsterplus_user.dennis.id
  }
}

# ---------------------------------------------------------------------------
# Team B: standalone grant resources — all 4 scopes.
# Exercises the standalone {team}_*_grant resources.
# ---------------------------------------------------------------------------

resource "dagsterplus_team" "standalone" {
  name = "acc-tf-team-standalone"
}

resource "dagsterplus_team_organization_grant" "standalone" {
  team_id = dagsterplus_team.standalone.id
  grant   = "ADMIN"
}

resource "dagsterplus_team_deployment_grant" "standalone" {
  team_id        = dagsterplus_team.standalone.id
  deployment     = dagsterplus_deployment.test.name
  custom_role_id = dagsterplus_role.observability.id
}

resource "dagsterplus_team_all_branch_deployments_grant" "standalone" {
  team_id = dagsterplus_team.standalone.id
  grant   = "LAUNCHER"
}

resource "dagsterplus_team_branch_deployments_grant" "standalone" {
  team_id           = dagsterplus_team.standalone.id
  parent_deployment = dagsterplus_deployment.test.name
  grant             = "EDITOR"
}

# ---------------------------------------------------------------------------
# Service user A: inline grant blocks — all 4 scopes, including location_grants
# inside the deployment_grant block to exercise per-location overrides.
# ---------------------------------------------------------------------------

resource "dagsterplus_service_user" "inline" {
  name        = "acc-tf-bot-inline"
  description = "Service user with inline grants"

  organization_grant {
    custom_role_id = dagsterplus_role.org_admin.id
  }

  deployment_grant {
    deployment = dagsterplus_deployment.test.name
    grant      = "VIEWER"

    # Per-code-location override; only valid inside deployment_grant.
    # (Asserted in test plan section C3 — note location_grants don't
    # round-trip through the mutation response so the second plan should
    # still be no-op.)
    # location_grants are currently inline only — uncomment when ready:
    # location_grants {
    #   location_name = dagsterplus_code_location.test.name
    #   grant         = "EDITOR"
    # }
  }

  all_branch_deployments_grant {
    grant = "LAUNCHER"
  }

  branch_deployments_grant {
    parent_deployment = dagsterplus_deployment.test.name
    grant             = "EDITOR"
  }
}

# ---------------------------------------------------------------------------
# Service user B: standalone grant resources — all 4 scopes.
# ---------------------------------------------------------------------------

resource "dagsterplus_service_user" "standalone" {
  name        = "acc-tf-bot-standalone"
  description = "Service user with standalone grants"
}

resource "dagsterplus_service_user_organization_grant" "standalone" {
  service_user_id = dagsterplus_service_user.standalone.id
  grant           = "ADMIN"
}

resource "dagsterplus_service_user_deployment_grant" "standalone" {
  service_user_id = dagsterplus_service_user.standalone.id
  deployment      = dagsterplus_deployment.test.name
  custom_role_id  = dagsterplus_role.observability.id
}

resource "dagsterplus_service_user_all_branch_deployments_grant" "standalone" {
  service_user_id = dagsterplus_service_user.standalone.id
  grant           = "LAUNCHER"
}

resource "dagsterplus_service_user_branch_deployments_grant" "standalone" {
  service_user_id   = dagsterplus_service_user.standalone.id
  parent_deployment = dagsterplus_deployment.test.name
  grant             = "EDITOR"
}

# ---------------------------------------------------------------------------
# User (dennis): standalone grant resources — all 4 scopes.
#
# Note: inline grant blocks on dagsterplus_user are intentionally NOT
# exercised here — they share the same model as the service_user inline
# path above. To smoke-test inline-on-user, follow test plan section B1
# (swap to inline blocks temporarily).
# ---------------------------------------------------------------------------

resource "dagsterplus_user_organization_grant" "dennis" {
  user_id = dagsterplus_user.dennis.id
  grant   = "ADMIN"
}

resource "dagsterplus_user_deployment_grant" "dennis" {
  user_id    = dagsterplus_user.dennis.id
  deployment = dagsterplus_deployment.test.name
  grant      = "EDITOR"
}

resource "dagsterplus_user_all_branch_deployments_grant" "dennis" {
  user_id = dagsterplus_user.dennis.id
  grant   = "LAUNCHER"
}

resource "dagsterplus_user_branch_deployments_grant" "dennis" {
  user_id           = dagsterplus_user.dennis.id
  parent_deployment = dagsterplus_deployment.test.name
  grant             = "EDITOR"
}
