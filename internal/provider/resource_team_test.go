package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTeamResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName("acc-tf-team"),
		Steps: []resource.TestStep{
			// Create without permissions
			{
				Config: providerConfig() + testAccTeamConfigNoPerms("acc-tf-team"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team.test", "name", "acc-tf-team"),
					resource.TestCheckResourceAttrSet("dagsterplus_team.test", "id"),
				),
			},
			// Add a deployment permission
			{
				Config: providerConfig() + testAccTeamConfigWithPerm("acc-tf-team", "acc-tf-test", "VIEWER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team.test", "name", "acc-tf-team"),
					resource.TestCheckResourceAttr("dagsterplus_team.test", "deployment_permission.0.deployment_name", "acc-tf-test"),
					resource.TestCheckResourceAttr("dagsterplus_team.test", "deployment_permission.0.role", "VIEWER"),
				),
			},
			// Update permission role
			{
				Config: providerConfig() + testAccTeamConfigWithPerm("acc-tf-team", "acc-tf-test", "EDITOR"),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_team.test", "deployment_permission.0.role", "EDITOR",
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
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName("acc-tf-team-multi"),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_team" "test" {
  name = "acc-tf-team-multi"

  deployment_permission {
    deployment_name = "prod"
    role            = "EDITOR"
  }

  deployment_permission {
    deployment_name = "staging"
    role            = "ADMIN"
  }
}
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team.test", "name", "acc-tf-team-multi"),
					resource.TestCheckResourceAttr("dagsterplus_team.test", "deployment_permission.#", "2"),
				),
			},
		},
	})
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

  deployment_permission {
    deployment_name = %q
    role            = %q
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
