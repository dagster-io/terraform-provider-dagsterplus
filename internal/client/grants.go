package client

import (
	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// LocationGrant represents a permission grant scoped to a specific code location.
// Used by all grant types (team, user, service user) for the deployment scope.
type LocationGrant struct {
	LocationName string
	Grant        string
	CustomRoleID string
}

// grantScopeToAPI maps the internal scope string (used in resource configs and
// client method arguments) to the GraphQL PermissionDeploymentScope enum value.
//
// "branch_deployments" and "all_branch_deployments" both map to ALL_BRANCH_DEPLOYMENTS;
// the per-parent variant is distinguished on the wire by a non-zero deploymentId.
var grantScopeToAPI = map[string]schema.PermissionDeploymentScope{
	"organization":           "ORGANIZATION",
	"deployment":             "DEPLOYMENT",
	"all_branch_deployments": "ALL_BRANCH_DEPLOYMENTS",
	"branch_deployments":     "ALL_BRANCH_DEPLOYMENTS",
}

// scopeForGrant maps an API enum + deploymentId back to an internal scope string.
// ALL_BRANCH_DEPLOYMENTS is ambiguous on the wire: when deploymentId is non-zero,
// it represents a per-base-branch grant under that parent deployment.
func scopeForGrant(deploymentScope schema.PermissionDeploymentScope, deploymentId int) string {
	switch deploymentScope {
	case "ORGANIZATION":
		return "organization"
	case "DEPLOYMENT":
		return "deployment"
	case "ALL_BRANCH_DEPLOYMENTS":
		if deploymentId != 0 {
			return "branch_deployments"
		}
		return "all_branch_deployments"
	}
	return string(deploymentScope)
}
