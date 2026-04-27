package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserDataSource_basic(t *testing.T) {
	testEmail := testAccExistingUserEmail(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccUserDataSourceConfig(testEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_user.test", "email", testEmail),
					resource.TestCheckResourceAttrSet("data.dagsterplus_user.test", "id"),
				),
			},
		},
	})
}

func TestAccUserDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccUserDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccUserDataSourceConfig(email string) string {
	return fmt.Sprintf(`
data "dagsterplus_user" "test" {
  email = %q
}
`, email)
}

func testAccUserDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_user" "test" {
  email = "this-user-does-not-exist-xyz@example.com"
}
`
}
