package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserTokenDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccUserTokenDataSourceConfig("acc-tf-ds-user-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_user_token.test", "name", "acc-tf-ds-user-token"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_user_token.test", "id"),
				),
			},
		},
	})
}

func TestAccUserTokenDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccUserTokenDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccUserTokenDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_user_token" "test" {
  name = %q
}

data "dagsterplus_user_token" "test" {
  name       = dagsterplus_user_token.test.name
  depends_on = [dagsterplus_user_token.test]
}
`, name)
}

func testAccUserTokenDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_user_token" "test" {
  name = "this-token-does-not-exist-xyz"
}
`
}
