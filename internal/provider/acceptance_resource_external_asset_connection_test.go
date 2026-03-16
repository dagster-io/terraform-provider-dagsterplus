package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccExternalAssetConnectionResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccExternalAssetsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckExternalAssetConnectionDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccExternalAssetConnectionConfig(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_external_asset_connection.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_external_asset_connection.test", "description", "initial description"),
					resource.TestCheckResourceAttr("dagsterplus_external_asset_connection.test", "connection_type", "SNOWFLAKE"),
					resource.TestCheckResourceAttr("dagsterplus_external_asset_connection.test", "source_config_yaml", "account: my_test_account\nuser: my_test_user"),
					resource.TestCheckResourceAttrSet("dagsterplus_external_asset_connection.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_external_asset_connection.test", "schedule_status"),
				),
			},
			// Update description
			{
				Config: providerConfig() + testAccExternalAssetConnectionConfig(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_external_asset_connection.test", "description", "updated description"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_external_asset_connection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccExternalAssetConnectionConfig(name, description string) string {
	return fmt.Sprintf(`
resource "dagsterplus_external_asset_connection" "test" {
  name               = %q
  description        = %q
  connection_type    = "SNOWFLAKE"
  source_config_yaml = "account: my_test_account\nuser: my_test_user"
}
`, name, description)
}

func testAccCheckExternalAssetConnectionDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		conns, err := c.ListExternalAssetConnections(context.Background())
		if err != nil {
			return fmt.Errorf("listing external asset connections: %w", err)
		}
		for _, conn := range conns {
			if conn.Name == name {
				return fmt.Errorf("external asset connection %q still exists with ID %q", name, conn.ID)
			}
		}
		return nil
	}
}

// TestAccExternalAssetConnectionResource_disappears verifies Terraform detects drift when the
// connection is deleted outside of Terraform.
func TestAccExternalAssetConnectionResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t); testAccExternalAssetsPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckExternalAssetConnectionDestroyed(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccExternalAssetConnectionConfig(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_external_asset_connection.test", "id"),
					testAccExternalAssetConnectionDisappears("dagsterplus_external_asset_connection.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccExternalAssetConnectionDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		return c.DeleteExternalAssetConnection(context.Background(), rs.Primary.ID)
	}
}
