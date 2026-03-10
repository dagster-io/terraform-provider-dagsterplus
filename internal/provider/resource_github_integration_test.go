package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccGithubIntegrationResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGithubIntegrationDeselected(),
		Steps: []resource.TestStep{
			// Create: select the GitHub installation
			{
				Config: providerConfig() + testAccGithubIntegrationConfig("my-github-org"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_github_integration.test", "account_name", "my-github-org"),
					resource.TestCheckResourceAttrSet("dagsterplus_github_integration.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_github_integration.test", "app_id"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_github_integration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGithubIntegrationConfig(accountName string) string {
	return fmt.Sprintf(`
resource "dagsterplus_github_integration" "test" {
  account_name = %q
}
`, accountName)
}

// testAccCheckGithubIntegrationDeselected verifies no GitHub App installation is active after destroy.
func testAccCheckGithubIntegrationDeselected() resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		org, err := c.GetOrganization(context.Background())
		if err != nil {
			return fmt.Errorf("error checking GitHub integration: %w", err)
		}
		if org.GitHub != nil {
			return fmt.Errorf("GitHub integration still active after destroy (account: %s)", org.GitHub.AccountName)
		}
		return nil
	}
}
