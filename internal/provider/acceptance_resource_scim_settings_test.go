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

func TestAccScimSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckScimSyncDisabled(),
		Steps: []resource.TestStep{
			// Create: enable SCIM
			{
				Config: providerConfig() + testAccScimSettingsConfig(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_scim_settings.test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("dagsterplus_scim_settings.test", "id"),
				),
			},
			// Update: disable SCIM
			{
				Config: providerConfig() + testAccScimSettingsConfig(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_scim_settings.test", "enabled", "false"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_scim_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccScimSettingsConfig(enabled bool) string {
	return fmt.Sprintf(`
resource "dagsterplus_scim_settings" "test" {
  enabled = %t
}
`, enabled)
}

// testAccCheckScimSyncDisabled verifies that SCIM sync is disabled after destroy.
// Destroy calls SetScimSyncEnabled(false), so we verify the API reflects that.
func testAccCheckScimSyncDisabled() resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		enabled, err := c.GetScimSyncEnabled(context.Background())
		if err != nil {
			return fmt.Errorf("error checking SCIM sync status: %w", err)
		}
		if enabled {
			return fmt.Errorf("SCIM sync is still enabled after destroy")
		}
		return nil
	}
}
