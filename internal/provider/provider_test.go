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

// providerConfig returns the HCL provider block for acceptance tests.
// Credentials are read from environment variables.
func providerConfig() string {
	return `
provider "dagsterplus" {}
`
}

// errRegexp compiles a regexp for use with resource.TestStep.ExpectError.
func errRegexp(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
