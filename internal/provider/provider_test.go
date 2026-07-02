package provider_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used in every acceptance test.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"dagsterplus": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck verifies that the required environment variables are set.
// Acceptance tests are skipped automatically when TF_ACC is unset; this function
// provides a clearer error when credentials are missing.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DAGSTER_CLOUD_API_TOKEN") == "" {
		t.Fatal("DAGSTER_CLOUD_API_TOKEN must be set for acceptance tests")
	}
	if os.Getenv("DAGSTER_CLOUD_ORGANIZATION") == "" {
		t.Fatal("DAGSTER_CLOUD_ORGANIZATION must be set for acceptance tests")
	}
}

// testAccDeployment returns the name of the primary pre-existing deployment used
// by acceptance tests. Override with DAGSTER_CLOUD_TEST_DEPLOYMENT.
func testAccDeployment() string {
	if v := os.Getenv("DAGSTER_CLOUD_TEST_DEPLOYMENT"); v != "" {
		return v
	}
	return "prod"
}

// testAccDeployment2 returns the name of a second pre-existing deployment used
// by tests that need two deployments (e.g. team multi-grant tests).
// Override with DAGSTER_CLOUD_TEST_DEPLOYMENT_2.
func testAccDeployment2() string {
	if v := os.Getenv("DAGSTER_CLOUD_TEST_DEPLOYMENT_2"); v != "" {
		return v
	}
	return "staging"
}

// testAccUserEmail returns the email address used for user acceptance tests.
// It must be an address that is invitable in the test organisation (i.e. not
// already a full member). Set DAGSTER_CLOUD_TEST_USER_EMAIL; the test is
// skipped when the variable is absent so as not to fail on unconfigured envs.
func testAccUserEmail(t *testing.T) string {
	t.Helper()
	email := os.Getenv("DAGSTER_CLOUD_TEST_USER_EMAIL")
	if email == "" {
		t.Skip("DAGSTER_CLOUD_TEST_USER_EMAIL must be set to run user acceptance tests")
	}
	return email
}

// testAccExistingUserEmail returns the email of an already-accepted org member
// used by data source acceptance tests. Pending-invite users are not visible
// through the API, so data source tests require a user who has already joined.
// Set DAGSTER_CLOUD_TEST_EXISTING_USER_EMAIL; the test is skipped when absent.
func testAccExistingUserEmail(t *testing.T) string {
	t.Helper()
	email := os.Getenv("DAGSTER_CLOUD_TEST_EXISTING_USER_EMAIL")
	if email == "" {
		t.Skip("DAGSTER_CLOUD_TEST_EXISTING_USER_EMAIL must be set to run user data source acceptance tests")
	}
	return email
}

// testAccWebhookURL returns an allowlisted webhook endpoint used by the webhook
// notification acceptance test. Dagster+ only permits webhook alerts to domains
// that have been allowlisted for the organisation (a support-gated setting), so
// the URL cannot be an arbitrary value. Set DAGSTER_CLOUD_TEST_WEBHOOK_URL to an
// allowlisted endpoint; the test is skipped when the variable is absent.
func testAccWebhookURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("DAGSTER_CLOUD_TEST_WEBHOOK_URL")
	if url == "" {
		t.Skip("DAGSTER_CLOUD_TEST_WEBHOOK_URL (an allowlisted webhook endpoint) must be set to run the webhook notification test")
	}
	return url
}

// providerConfig returns the HCL provider block for acceptance tests.
// Credentials are read from environment variables.
func providerConfig() string {
	return `
provider "dagsterplus" {}
`
}

// testAccExternalAssetsPreCheck skips the test if external asset connections are not
// enabled for the test organisation. Set DAGSTER_CLOUD_TEST_EXTERNAL_ASSETS=1 to opt in.
func testAccExternalAssetsPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DAGSTER_CLOUD_TEST_EXTERNAL_ASSETS") == "" {
		t.Skip("DAGSTER_CLOUD_TEST_EXTERNAL_ASSETS must be set to run external asset connection tests")
	}
}

// testAccDeployment2PreCheck skips the test if DAGSTER_CLOUD_TEST_DEPLOYMENT_2 is not set.
// Tests requiring a second deployment are skipped unless a real second deployment is provided.
func testAccDeployment2PreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DAGSTER_CLOUD_TEST_DEPLOYMENT_2") == "" {
		t.Skip("DAGSTER_CLOUD_TEST_DEPLOYMENT_2 must be set to run multi-deployment tests")
	}
}

// testAccOrgGrantsPreCheck skips the test if org-level team grants are not enabled for the
// test organisation. Set DAGSTER_CLOUD_TEST_ORG_GRANTS=1 to opt in.
func testAccOrgGrantsPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DAGSTER_CLOUD_TEST_ORG_GRANTS") == "" {
		t.Skip("DAGSTER_CLOUD_TEST_ORG_GRANTS must be set to run org-level team grant tests")
	}
}

// testAccGithubIntegrationPreCheck skips the test if GitHub integration is not enabled for the
// test organisation. Set DAGSTER_CLOUD_TEST_GITHUB_INTEGRATION=1 to opt in.
func testAccGithubIntegrationPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("DAGSTER_CLOUD_TEST_GITHUB_INTEGRATION") == "" {
		t.Skip("DAGSTER_CLOUD_TEST_GITHUB_INTEGRATION must be set to run GitHub integration tests")
	}
}

// errRegexp compiles a regexp for use with resource.TestStep.ExpectError.
func errRegexp(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
