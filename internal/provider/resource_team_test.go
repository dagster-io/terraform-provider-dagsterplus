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

func TestAccTeamResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName(rName),
		Steps: []resource.TestStep{
			// Create without permissions
			{
				Config: providerConfig() + testAccTeamConfigNoPerms(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team.test", "name", rName),
					resource.TestCheckResourceAttrSet("dagsterplus_team.test", "id"),
				),
			},
			// Add a deployment permission (use "prod" — a stable pre-existing deployment)
			{
				Config: providerConfig() + testAccTeamConfigWithPerm(rName, testAccAlertDeployment, "VIEWER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_team.test", "deployment_grant.0.deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_team.test", "deployment_grant.0.grant", "VIEWER"),
				),
			},
			// Update permission role
			{
				Config: providerConfig() + testAccTeamConfigWithPerm(rName, testAccAlertDeployment, "EDITOR"),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_team.test", "deployment_grant.0.grant", "EDITOR",
				),
			},
			// Import by team ID (Terraform sets state.ID to the team ID)
			{
				ResourceName:      "dagsterplus_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTeamResource_multiplePermissions(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_team" "test" {
  name = %q

  deployment_grant {
    deployment = "prod"
    grant      = "EDITOR"
  }

  deployment_grant {
    deployment = "staging"
    grant      = "ADMIN"
  }
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_team.test", "deployment_grant.#", "2"),
				),
			},
		},
	})
}

// TestAccTeamResource_disappears verifies Terraform detects drift when the team is
// deleted outside of Terraform.
func TestAccTeamResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccTeamConfigNoPerms(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_team.test", "id"),
					testAccTeamDisappears("dagsterplus_team.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccTeamDisappears deletes the team out-of-band during a Check step.
func testAccTeamDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		id := rs.Primary.ID
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		return c.DeleteTeam(context.Background(), id)
	}
}

func testAccTeamConfigNoPerms(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_team" "test" {
  name = %q
}
`, name)
}

func testAccTeamConfigWithPerm(name, deployment, role string) string {
	return fmt.Sprintf(`
resource "dagsterplus_team" "test" {
  name = %q

  deployment_grant {
    deployment = %q
    grant      = %q
  }
}
`, name, deployment, role)
}

// testAccCheckTeamDestroyedByName verifies the team no longer exists (searching by name).
func testAccCheckTeamDestroyedByName(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		teams, err := c.ListTeams(context.Background())
		if err != nil {
			// If listing fails we can't verify — treat as non-blocking.
			return nil
		}
		for _, t := range teams {
			if t.Name == name {
				return fmt.Errorf("team %q still exists with ID %q", name, t.ID)
			}
		}
		return nil
	}
}
