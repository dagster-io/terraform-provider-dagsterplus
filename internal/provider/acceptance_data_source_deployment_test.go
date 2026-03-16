package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccDeploymentDataSourceConfig("acc-tf-ds-dep"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_deployment.test", "name", "acc-tf-ds-dep"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployment.test", "id"),
				),
			},
		},
	})
}

func TestAccDeploymentDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccDeploymentDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccDeploymentDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment" "test" {
  name = %q
}

data "dagsterplus_deployment" "test" {
  name       = dagsterplus_deployment.test.name
  depends_on = [dagsterplus_deployment.test]
}
`, name)
}

func testAccDeploymentDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_deployment" "test" {
  name = "this-deployment-does-not-exist-xyz"
}
`
}
