package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserTokenDataSource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccUserTokenDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_user_token.test", "name", rName),
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
  user_id    = dagsterplus_user_token.test.user_id
  depends_on = [dagsterplus_user_token.test]
}
`, name)
}

func testAccUserTokenDataSourceNotFoundConfig() string {
	return `
resource "dagsterplus_user_token" "helper" {
  name = "acc-tf-ds-helper-token"
}

data "dagsterplus_user_token" "test" {
  name       = "no-such-token"
  user_id    = dagsterplus_user_token.helper.user_id
  depends_on = [dagsterplus_user_token.helper]
}
`
}
