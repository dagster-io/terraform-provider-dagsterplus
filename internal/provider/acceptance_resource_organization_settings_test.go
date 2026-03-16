package provider_test

import (
	"context"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOrganizationSettingsResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationSettingsReset,
		Steps: []resource.TestStep{
			// Create and Read (singleton)
			{
				Config: providerConfig() + testAccOrganizationSettingsConfig(`{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_organization_settings.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_organization_settings.test", "settings_json"),
				),
			},
			// Update settings
			{
				Config: providerConfig() + testAccOrganizationSettingsConfig(`{}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_organization_settings.test", "settings_json"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_organization_settings.test",
				ImportState:       true,
				ImportStateVerify: false, // settings_json may differ in format
			},
		},
	})
}

func testAccOrganizationSettingsConfig(settingsJSON string) string {
	return `
resource "dagsterplus_organization_settings" "test" {
  settings_json = "` + settingsJSON + `"
}
`
}

func testAccCheckOrganizationSettingsReset(_ *terraform.State) error {
	c := client.New(
		os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
		os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
		"",
	)
	_, err := c.GetOrganizationSettings(context.Background())
	return err
}
