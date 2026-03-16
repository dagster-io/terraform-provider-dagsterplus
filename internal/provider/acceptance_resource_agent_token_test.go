package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAgentTokenResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTokenDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccAgentTokenConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "name", rName),
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token.test", "token"),
				),
			},
			// Import by token ID (token and name not recoverable via API)
			{
				ResourceName:            "dagsterplus_agent_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "name"},
			},
		},
	})
}

// TestAccAgentTokenResource_disappears verifies Terraform detects drift when the token is
// deleted outside of Terraform.
func TestAccAgentTokenResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTokenDestroyed(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccAgentTokenConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token.test", "id"),
					testAccAgentTokenDisappears("dagsterplus_agent_token.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAgentTokenDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		return c.DeleteAgentToken(context.Background(), rs.Primary.ID)
	}
}

func testAccAgentTokenConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_agent_token" "test" {
  name = %q
}
`, name)
}

func testAccCheckAgentTokenDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		tokens, err := c.ListAgentTokens(context.Background())
		if err != nil {
			return fmt.Errorf("listing agent tokens: %w", err)
		}
		for _, tok := range tokens {
			if tok.Name == name {
				return fmt.Errorf("agent token %q still exists with ID %q", name, tok.ID)
			}
		}
		return nil
	}
}
