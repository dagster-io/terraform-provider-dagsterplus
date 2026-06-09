package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgentTokenDeploymentGrantResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// Destroying the agent token removes its grants; verifying the token is
		// gone is sufficient to confirm the grant is gone too.
		CheckDestroy: testAccCheckAgentTokenDestroyed(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccAgentTokenDeploymentGrantConfig(rName, testAccDeployment()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token_deployment_grant.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token_deployment_grant.test", "agent_token_id"),
					resource.TestCheckResourceAttr("dagsterplus_agent_token_deployment_grant.test", "deployment", testAccDeployment()),
				),
			},
			{
				ResourceName:      "dagsterplus_agent_token_deployment_grant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgentTokenDeploymentGrantConfig(name, deployment string) string {
	return fmt.Sprintf(`
resource "dagsterplus_agent_token" "test" {
  name         = %q
  organization = false
}

resource "dagsterplus_agent_token_deployment_grant" "test" {
  agent_token_id = dagsterplus_agent_token.test.id
  deployment     = %q
}
`, name, deployment)
}
