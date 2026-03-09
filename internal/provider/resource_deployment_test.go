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

func TestAccDeploymentResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroyed("acc-tf-test"),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccDeploymentConfig("acc-tf-test", "BRANCH"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "name", "acc-tf-test"),
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "type", "BRANCH"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_deployment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// testAccDeploymentConfig returns HCL for a deployment resource.
func testAccDeploymentConfig(name, deploymentType string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment" "test" {
  name = %q
  type = %q
}
`, name, deploymentType)
}

// testAccCheckDeploymentDestroyed verifies the deployment no longer exists.
func testAccCheckDeploymentDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		_, err := c.GetDeployment(context.Background(), name)
		if err == nil {
			return fmt.Errorf("deployment %q still exists", name)
		}
		return nil
	}
}
