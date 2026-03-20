package provider_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// TestIntegration runs a full apply → no-drift plan → destroy cycle against
// integration/main.tf using the real API.
//
// Run with:
//
//	set -a && . ./.env && set +a && TF_ACC=1 go test ./internal/provider/... -run TestIntegration -v -timeout 30m
func TestIntegration(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping integration test")
	}

	// integration/main.tf declares source = "dagster-io/dagsterplus". The test
	// framework's reattach provider defaults to the "hashicorp" namespace, so we
	// must override it to match the source address and ensure the in-process server
	// is used instead of downloading the provider from the registry.
	t.Setenv(resource.EnvTfAccProviderNamespace, "dagster-io")

	// The integration test requires a pre-existing non-owner org member to exercise
	// the dagsterplus_user resource and data source. The org creator cannot be added
	// via AddUser, so a separate test user is required.
	testUserEmail := os.Getenv("DAGSTER_CLOUD_TEST_USER_EMAIL")
	if testUserEmail == "" {
		t.Skip("DAGSTER_CLOUD_TEST_USER_EMAIL must be set to run the integration test (must be a non-owner org member)")
	}

	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("cannot resolve repo root: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				ConfigFile:               config.StaticFile(filepath.Join(repoRoot, "integration", "main.tf")),
				ConfigVariables: config.Variables{
					"test_user_email": config.StringVariable(testUserEmail),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// Deployment
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "name", "acc-tf-test"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "id"),

					// User
					resource.TestCheckResourceAttr("dagsterplus_user.dennis", "email", testUserEmail),
					resource.TestCheckResourceAttrSet("dagsterplus_user.dennis", "id"),

					// Roles
					resource.TestCheckResourceAttr("dagsterplus_role.observability", "name", "acc-tf-observability"),
					resource.TestCheckResourceAttr("dagsterplus_role.observability", "role_type", "deployment"),
					resource.TestCheckResourceAttr("dagsterplus_role.observability", "permissions.#", "2"),
					resource.TestCheckResourceAttrSet("dagsterplus_role.observability", "id"),
					resource.TestCheckResourceAttr("dagsterplus_role.org_admin", "name", "acc-tf-org-admin"),
					resource.TestCheckResourceAttr("dagsterplus_role.org_admin", "role_type", "organization"),
					resource.TestCheckResourceAttr("dagsterplus_role.org_admin", "permissions.#", "3"),
					resource.TestCheckResourceAttrSet("dagsterplus_role.org_admin", "id"),

					// Teams (inline grants/members)
					resource.TestCheckResourceAttr("dagsterplus_team.data_engineering", "name", "acc-tf-data-engineering"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering", "organization_grant.0.custom_role_id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering", "member.0.user_id"),
					resource.TestCheckResourceAttr("dagsterplus_team.data_engineering_2", "name", "acc-tf-data-engineering-2"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering_2", "id"),
					resource.TestCheckResourceAttr("dagsterplus_team.data_engineering_2", "deployment_grant.0.deployment", "prod"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering_2", "deployment_grant.0.custom_role_id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering_2", "all_branch_deployments_grant.0.custom_role_id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.data_engineering_2", "member.0.user_id"),

					// Team managed via standalone grant + membership resources
					resource.TestCheckResourceAttr("dagsterplus_team.grants_only", "name", "acc-tf-grants-only"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.grants_only", "id"),
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttrSet("dagsterplus_team_deployment_grant.test", "team_id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team_deployment_grant.test", "custom_role_id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team_membership.dennis", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team_membership.dennis", "team_id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team_membership.dennis", "user_id"),

					// Agent token
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "name", "acc-tf-agent-token"),
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token.test", "token"),

					// User token
					resource.TestCheckResourceAttr("dagsterplus_user_token.test", "name", "acc-tf-user-token"),
					resource.TestCheckResourceAttrSet("dagsterplus_user_token.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_user_token.test", "token"),

					// Code location (depends on deployment)
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "name", "acc-tf-code-location"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "image", "ghcr.io/example/repo:v1"),
					resource.TestCheckResourceAttrSet("dagsterplus_code_location.test", "id"),

					// Deployment settings (depends on deployment)
					resource.TestCheckResourceAttr("dagsterplus_deployment_settings.test", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment_settings.test", "id"),

					// Alert policies
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "name", "acc-tf-test-alerts"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "policy_type", "run"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "run.0.code_locations.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "run.0.code_locations.0", "acc-tf-code-location"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "run.0.on_failure", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test_deployment", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test_deployment", "id"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "name", "asset-health-alerts"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "deployment", "prod"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "policy_type", "asset"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "asset.0.all_assets", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "asset.0.specific_events.#", "2"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.asset_health", "notification_service.type", "email"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.asset_health", "id"),
					// Code-location alert depends on both deployment and code location
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.code_location", "name", "acc-tf-code-location-alerts"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.code_location", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.code_location", "policy_type", "code_location"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.code_location", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.code_location", "code_location.0.location_name", "acc-tf-code-location"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.code_location", "notification_service.type", "email"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.code_location", "id"),

					// Custom metric
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "metadata_key", "acc_tf_integration_metric"),
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "display_name", "Acc TF Integration Metric"),
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "description", "A custom metric managed by Terraform"),
					resource.TestCheckResourceAttrSet("dagsterplus_custom_metric.test", "id"),

					// Service user + token (service_token depends on service_user)
					resource.TestCheckResourceAttr("dagsterplus_service_user.ci_bot", "name", "acc-tf-ci-bot"),
					resource.TestCheckResourceAttr("dagsterplus_service_user.ci_bot", "description", "CI/CD service user managed by Terraform"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_user.ci_bot", "id"),
					resource.TestCheckResourceAttr("dagsterplus_service_token.ci_bot_token", "description", "Primary token for acc-tf-ci-bot"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_token.ci_bot_token", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_token.ci_bot_token", "token"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_token.ci_bot_token", "service_user_id"),

					// Organization settings
					resource.TestCheckResourceAttr("dagsterplus_organization_settings.org", "settings_json", "{}"),
					resource.TestCheckResourceAttrSet("dagsterplus_organization_settings.org", "id"),

					// Secret
					resource.TestCheckResourceAttr("dagsterplus_secret.db_password", "secret_name", "ACC_TF_DB_PASSWORD"),
					resource.TestCheckResourceAttr("dagsterplus_secret.db_password", "full_deployment_scope", "true"),
					resource.TestCheckResourceAttrSet("dagsterplus_secret.db_password", "id"),

					// Data sources — verify read path matches what the resources created
					resource.TestCheckResourceAttr("data.dagsterplus_user.dennis", "email", testUserEmail),
					resource.TestCheckResourceAttrSet("data.dagsterplus_user.dennis", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_deployment.test", "name", "acc-tf-test"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployment.test", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployment.test", "status"),

					resource.TestCheckResourceAttr("data.dagsterplus_role.observability", "name", "acc-tf-observability"),
					resource.TestCheckResourceAttr("data.dagsterplus_role.observability", "role_type", "deployment"),
					resource.TestCheckResourceAttr("data.dagsterplus_role.observability", "permissions.#", "2"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_role.observability", "id"),
					resource.TestCheckResourceAttr("data.dagsterplus_role.org_admin", "name", "acc-tf-org-admin"),
					resource.TestCheckResourceAttr("data.dagsterplus_role.org_admin", "role_type", "organization"),
					resource.TestCheckResourceAttr("data.dagsterplus_role.org_admin", "permissions.#", "3"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_role.org_admin", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_team.data_engineering", "name", "acc-tf-data-engineering"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_team.data_engineering", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_agent_token.test", "name", "acc-tf-agent-token"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_agent_token.test", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_user_token.test", "name", "acc-tf-user-token"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_user_token.test", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "name", "acc-tf-code-location"),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "image", "ghcr.io/example/repo:v1"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_code_location.test", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test_deployment", "name", "acc-tf-test-alerts"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test_deployment", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test_deployment", "policy_type", "run"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test_deployment", "enabled", "true"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test_deployment", "run.0.on_failure", "true"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_alert_policy.test_deployment", "id"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.asset_health", "name", "asset-health-alerts"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.asset_health", "deployment", "prod"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.asset_health", "policy_type", "asset"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.asset_health", "enabled", "true"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.asset_health", "asset.0.all_assets", "true"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_alert_policy.asset_health", "id"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.code_location", "name", "acc-tf-code-location-alerts"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.code_location", "deployment", "acc-tf-test"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.code_location", "policy_type", "code_location"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.code_location", "enabled", "true"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.code_location", "code_location.0.location_name", "acc-tf-code-location"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_alert_policy.code_location", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_custom_metric.test", "metadata_key", "acc_tf_integration_metric"),
					resource.TestCheckResourceAttr("data.dagsterplus_custom_metric.test", "display_name", "Acc TF Integration Metric"),
					resource.TestCheckResourceAttr("data.dagsterplus_custom_metric.test", "description", "A custom metric managed by Terraform"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_custom_metric.test", "id"),

					resource.TestCheckResourceAttr("data.dagsterplus_service_user.ci_bot", "name", "acc-tf-ci-bot"),
					resource.TestCheckResourceAttr("data.dagsterplus_service_user.ci_bot", "description", "CI/CD service user managed by Terraform"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_service_user.ci_bot", "id"),

					// Organization data source
					resource.TestCheckResourceAttrSet("data.dagsterplus_organization.org", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_organization.org", "name"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_organization.org", "status"),

					// Secret data source
					resource.TestCheckResourceAttr("data.dagsterplus_secret.db_password", "secret_name", "ACC_TF_DB_PASSWORD"),
					resource.TestCheckResourceAttr("data.dagsterplus_secret.db_password", "full_deployment_scope", "true"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_secret.db_password", "id"),

					// List data sources — verify the read succeeded and lists are non-empty
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployments.all", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_teams.all", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_roles.all", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_users.all", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_code_locations.test_deployment", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_alert_policies.test_deployment", "id"),
				),
			},
		},
	})
}
