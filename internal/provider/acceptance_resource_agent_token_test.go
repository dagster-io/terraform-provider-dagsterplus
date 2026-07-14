package provider_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

// TestAccAgentTokenResource_grants exercises the inline grant attributes:
// removing the default organization grant, granting a deployment, then updating
// to add an all-branch-deployments grant.
func TestAccAgentTokenResource_grants(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTokenDestroyed(rName),
		Steps: []resource.TestStep{
			// Create: drop the default org grant, grant a single deployment.
			{
				Config: providerConfig() + testAccAgentTokenGrantsConfig(rName, "false", "false", fmt.Sprintf("[%q]", testAccDeployment())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "organization", "false"),
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "all_branch_deployments", "false"),
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "deployment_grants.#", "1"),
					resource.TestCheckTypeSetElemAttr("dagsterplus_agent_token.test", "deployment_grants.*", testAccDeployment()),
				),
			},
			// Update: re-enable org grant and add an all-branch-deployments grant.
			{
				Config: providerConfig() + testAccAgentTokenGrantsConfig(rName, "true", "true", fmt.Sprintf("[%q]", testAccDeployment())),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "organization", "true"),
					resource.TestCheckResourceAttr("dagsterplus_agent_token.test", "all_branch_deployments", "true"),
				),
			},
			// Import: token/name are not recoverable and the grant name sets are
			// not repopulated on import (only organization/all_branch are read back).
			{
				ResourceName:            "dagsterplus_agent_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "name", "deployment_grants"},
			},
		},
	})
}

func testAccAgentTokenGrantsConfig(name, organization, allBranch, deploymentGrants string) string {
	return fmt.Sprintf(`
resource "dagsterplus_agent_token" "test" {
  name                   = %q
  organization           = %s
  all_branch_deployments = %s
  deployment_grants      = %s
}
`, name, organization, allBranch, deploymentGrants)
}

// TestAccAgentTokenResource_partialCreateNoLeak reproduces issue #44: the token
// is created in Dagster+ before its grants are applied, so a grant failure during
// Create leaves a real token behind. Here the failure is induced deterministically
// by pointing deployment_grants at a deployment that does not exist, which errors
// in applyGrants after the token already exists remotely.
//
// The fix persists the token to Terraform state before applying grants, so the
// failed create is tracked; the corrected re-apply reconciles that same token
// rather than creating a new one. The token-count check is the discriminator:
// with the bug, state was never written, so the re-apply would create a SECOND
// token with the same name (count 2). With the fix, exactly one token exists.
func TestAccAgentTokenResource_partialCreateNoLeak(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	missingDeployment := "acc-tf-nonexistent-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAgentTokenDestroyed(rName),
		Steps: []resource.TestStep{
			// Create fails while resolving the (nonexistent) deployment grant,
			// after the token has already been created remotely.
			{
				Config:      providerConfig() + testAccAgentTokenGrantsConfig(rName, "true", "false", fmt.Sprintf("[%q]", missingDeployment)),
				ExpectError: regexp.MustCompile("Error resolving deployment"),
			},
			// Re-apply with a valid config converges on a single token; no orphan
			// was leaked by the failed create.
			{
				Config: providerConfig() + testAccAgentTokenConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_agent_token.test", "id"),
					testAccCheckAgentTokenCount(rName, 1),
				),
			},
		},
	})
}

// testAccCheckAgentTokenCount asserts exactly `want` non-revoked tokens carry the
// given name. Used to prove a failed create did not leak an untracked token.
func testAccCheckAgentTokenCount(name string, want int) resource.TestCheckFunc {
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
		count := 0
		for _, tok := range tokens {
			if tok.Name == name {
				count++
			}
		}
		if count != want {
			return fmt.Errorf("found %d agent tokens named %q, want %d (a leaked token indicates a partial-create failure was not tracked in state)", count, name, want)
		}
		return nil
	}
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
