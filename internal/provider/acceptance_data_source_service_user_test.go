package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceUserDataSource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccServiceUserDataSourceConfig(rName, "ds test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_service_user.test", "name", rName),
					resource.TestCheckResourceAttr("data.dagsterplus_service_user.test", "name", rName),
					resource.TestCheckResourceAttr("data.dagsterplus_service_user.test", "description", "ds test description"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_service_user.test", "id"),
				),
			},
		},
	})
}

func TestAccServiceUserDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccServiceUserDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccServiceUserDataSourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "dagsterplus_service_user" "test" {
  name        = %q
  description = %q
}

data "dagsterplus_service_user" "test" {
  name       = dagsterplus_service_user.test.name
  depends_on = [dagsterplus_service_user.test]
}
`, name, description)
}

func testAccServiceUserDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_service_user" "test" {
  name = "no-such-user"
}
`
}
