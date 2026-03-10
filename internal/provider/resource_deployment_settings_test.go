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

func TestAccDeploymentSettingsResource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentSettingsReset("prod"),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccDeploymentSettingsConfig("prod", `{"run_queue":{"max_concurrent_runs":5}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_deployment_settings.test", "deployment", "prod"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment_settings.test", "settings_json"),
				),
			},
			// Update settings
			{
				Config: providerConfig() + testAccDeploymentSettingsConfig("prod", `{"run_queue":{"max_concurrent_runs":10}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_deployment_settings.test", "deployment", "prod"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_deployment_settings.test",
				ImportState:       true,
				ImportStateVerify: false, // settings_json may be reformatted by API
			},
		},
	})
}

func testAccDeploymentSettingsConfig(deployment, settingsJSON string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment_settings" "test" {
  deployment    = %q
  settings_json = %q
}
`, deployment, settingsJSON)
}

func testAccCheckDeploymentSettingsReset(deployment string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		// After destroy, settings should be reset to {} - just verify we can read them
		_, err := c.GetDeploymentSettings(context.Background(), deployment)
		return err
	}
}
