package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccOrganizationDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dagsterplus_organization.test", "id"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_organization.test", "name"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_organization.test", "status"),
				),
			},
		},
	})
}

func testAccOrganizationDataSourceConfig() string {
	return `
data "dagsterplus_organization" "test" {}
`
}
