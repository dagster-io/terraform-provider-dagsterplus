package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgentTokenDataSource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccAgentTokenDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_agent_token.test", "name", rName),
					resource.TestCheckResourceAttrSet("data.dagsterplus_agent_token.test", "id"),
				),
			},
		},
	})
}

func TestAccAgentTokenDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccAgentTokenDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccAgentTokenDataSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_agent_token" "test" {
  name = %q
}

data "dagsterplus_agent_token" "test" {
  name       = dagsterplus_agent_token.test.name
  depends_on = [dagsterplus_agent_token.test]
}
`, name)
}

func testAccAgentTokenDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_agent_token" "test" {
  name = "this-token-does-not-exist-xyz"
}
`
}
