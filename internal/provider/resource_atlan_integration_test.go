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

func TestAccAtlanIntegrationResource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAtlanIntegrationDeleted(),
		Steps: []resource.TestStep{
			// Create
			{
				Config: providerConfig() + testAccAtlanIntegrationConfig("initial-token", "my-org.atlan.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_atlan_integration.test", "domain", "my-org.atlan.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_atlan_integration.test", "id"),
				),
			},
			// Update domain
			{
				Config: providerConfig() + testAccAtlanIntegrationConfig("updated-token", "updated-org.atlan.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_atlan_integration.test", "domain", "updated-org.atlan.com"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_atlan_integration.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAtlanIntegrationConfig(token, domain string) string {
	return fmt.Sprintf(`
resource "dagsterplus_atlan_integration" "test" {
  token  = %q
  domain = %q
}
`, token, domain)
}

// testAccCheckAtlanIntegrationDeleted verifies the Atlan integration is removed after destroy.
func testAccCheckAtlanIntegrationDeleted() resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		atlan, err := c.GetAtlanIntegration(context.Background())
		if err != nil {
			return fmt.Errorf("error checking Atlan integration: %w", err)
		}
		if atlan != nil {
			return fmt.Errorf("Atlan integration still exists after destroy")
		}
		return nil
	}
}
