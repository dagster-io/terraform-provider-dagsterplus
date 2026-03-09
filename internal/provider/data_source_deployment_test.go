package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Reads an existing deployment — assumes "prod" exists in the test org.
				// Adjust to a deployment name that exists in your Dagster+ org.
				Config: providerConfig() + testAccDeploymentDataSourceConfig("prod"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_deployment.test", "name", "prod"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployment.test", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployment.test", "type"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_deployment.test", "created_at"),
				),
			},
		},
	})
}

func TestAccDeploymentDataSource_notFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccDeploymentDataSourceConfig("this-deployment-does-not-exist-xyz"),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccDeploymentDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "dagsterplus_deployment" "test" {
  name = %q
}
`, name)
}
