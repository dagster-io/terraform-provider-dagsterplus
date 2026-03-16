package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamDataSource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccTeamDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_team.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.dagsterplus_team.test", "id"),
				),
			},
		},
	})
}

func TestAccTeamDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccTeamDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccTeamDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_team" "test" {
  name = %q
}

data "dagsterplus_team" "test" {
  name       = dagsterplus_team.test.name
  depends_on = [dagsterplus_team.test]
}
`, name)
}

func testAccTeamDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_team" "test" {
  name = "this-team-does-not-exist-xyz"
}
`
}
