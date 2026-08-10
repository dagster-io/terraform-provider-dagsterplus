package resources

import "strings"

// Enum value sets shared between schema validators and attribute descriptions.
//
// Each slice is the single source of truth for one enum: the same values feed
// both stringvalidator.OneOf(...) and the generated documentation (via
// withEnumValues), so the validator and the docs can never drift.
var (
	// grantLevels are the standard (non-custom-role) permission levels.
	grantLevels = []string{"VIEWER", "LAUNCHER", "EDITOR", "ADMIN"}

	// orgGrantLevels are the permission levels accepted for organization-level grants.
	orgGrantLevels = []string{"ADMIN"}

	// roleTypes are the scopes a custom role can target.
	roleTypes = []string{"deployment", "organization"}

	// externalAssetConnectionTypes are the supported external connection sources.
	externalAssetConnectionTypes = []string{"SNOWFLAKE", "BIGQUERY", "POSTGRES", "DATABRICKS"}

	// alertPolicyTypes are the categories of alert policy.
	alertPolicyTypes = []string{"asset", "run", "code_location", "automation", "agent_downtime", "budget", "insight_metric"}

	// alertComparisonOperators are the threshold comparison operators.
	alertComparisonOperators = []string{"greater_than", "less_than"}

	// notificationTypes are the supported alert notification channels.
	notificationTypes = []string{"email", "email_owners", "slack", "microsoft_teams", "pagerduty", "webhook"}

	// customRolePermissions is the full set of valid permission values for a custom role.
	customRolePermissions = []string{
		"delete_runs",
		"edit_alerts",
		"edit_all_catalog_views",
		"edit_code_locations",
		"edit_concurrency_limits",
		"edit_custom_roles",
		"edit_deployment_permissions",
		"edit_deployment_settings",
		"edit_dynamic_partitions",
		"edit_external_asset_connections",
		"edit_insights_metrics",
		"edit_issues",
		"edit_secrets",
		"edit_sensor_cursors",
		"edit_users_and_teams",
		"manage_billing",
		"manage_branch_deployments",
		"manage_full_deployments",
		"manage_service_users",
		"manage_sso_and_scim",
		"read_and_edit_agent_tokens",
		"read_and_edit_all_user_tokens",
		"read_audit_log",
		"read_secret_values",
		"redeploy_code_locations",
		"report_asset_events",
		"start_and_stop_runs",
		"toggle_schedules",
		"toggle_sensors",
		"wipe_assets",
	}
)

// Shared base descriptions for enum attributes that recur across resources.
// The list of valid values is appended at render time via withEnumValues, so
// these constants describe only the attribute's meaning.
const (
	// grantLevelDescription is the base description for a standard permission-level attribute.
	grantLevelDescription = "Standard permission level. Exactly one of grant or custom_role_id must be set."

	// orgGrantDescription is the base description for an organization-level grant attribute.
	orgGrantDescription = "Organization-level permission. For any non-admin role use custom_role_id with an organization-scoped custom role. Exactly one of grant or custom_role_id must be set."
)

// withEnumValues appends the list of valid enum values to a base attribute
// description, formatted as inline backticked values (e.g. "Valid values:
// `a`, `b`."). Passing the same slice that feeds stringvalidator.OneOf
// guarantees the generated documentation matches the validator exactly.
func withEnumValues(base string, values []string) string {
	return base + " Valid values: `" + strings.Join(values, "`, `") + "`."
}
