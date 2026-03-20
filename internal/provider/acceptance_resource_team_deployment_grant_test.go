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

func TestAccTeamDeploymentGrantResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName(rName),
		Steps: []resource.TestStep{
			// Create
			{
				Config: providerConfig() + testAccTeamDeploymentGrantConfig(rName, testAccDeployment(), "VIEWER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_team_deployment_grant.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_team_deployment_grant.test", "team_id"),
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "deployment", testAccDeployment()),
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "grant", "VIEWER"),
				),
			},
			// Update grant level
			{
				Config: providerConfig() + testAccTeamDeploymentGrantConfig(rName, testAccDeployment(), "EDITOR"),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_team_deployment_grant.test", "grant", "EDITOR",
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_team_deployment_grant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccTeamDeploymentGrantResource_locationGrants tests that per-location grants
// are created, updated, and imported correctly.
// A real code location is created first because the API only persists location grants
// for locations that actually exist in the deployment.
func TestAccTeamDeploymentGrantResource_locationGrants(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	locName := "acc-tf-loc-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	deployment := testAccDeployment()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName(rName),
		Steps: []resource.TestStep{
			// Create with a location grant
			{
				Config: providerConfig() + testAccTeamDeploymentGrantWithLocationConfig(rName, deployment, "VIEWER", locName, "LAUNCHER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "grant", "VIEWER"),
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "location_grants.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "location_grants.0.location_name", locName),
					resource.TestCheckResourceAttr("dagsterplus_team_deployment_grant.test", "location_grants.0.grant", "LAUNCHER"),
				),
			},
			// Update location grant level
			{
				Config: providerConfig() + testAccTeamDeploymentGrantWithLocationConfig(rName, deployment, "VIEWER", locName, "EDITOR"),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_team_deployment_grant.test", "location_grants.0.grant", "EDITOR",
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_team_deployment_grant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccTeamDeploymentGrantResource_disappears verifies Terraform detects drift when
// the deployment grant is removed outside of Terraform.
func TestAccTeamDeploymentGrantResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTeamDestroyedByName(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccTeamDeploymentGrantConfig(rName, testAccDeployment(), "VIEWER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_team_deployment_grant.test", "id"),
					testAccTeamDeploymentGrantDisappears("dagsterplus_team_deployment_grant.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTeamDeploymentGrantDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		teamID := rs.Primary.Attributes["team_id"]
		deployment := rs.Primary.Attributes["deployment"]
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		intID, err := c.GetDeploymentIntID(context.Background(), deployment)
		if err != nil {
			return fmt.Errorf("resolving deployment: %w", err)
		}
		return c.DeleteTeamGrant(context.Background(), teamID, "deployment", intID)
	}
}

func testAccTeamDeploymentGrantConfig(teamName, deployment, grant string) string {
	return fmt.Sprintf(`
resource "dagsterplus_team" "test" {
  name = %q
}

resource "dagsterplus_team_deployment_grant" "test" {
  team_id    = dagsterplus_team.test.id
  deployment = %q
  grant      = %q
}
`, teamName, deployment, grant)
}

func testAccTeamDeploymentGrantWithLocationConfig(teamName, deployment, grant, locationName, locationGrant string) string {
	return fmt.Sprintf(`
resource "dagsterplus_code_location" "loc" {
  deployment = %q
  name       = %q
  image      = "ghcr.io/example/repo:v1"

  code_source {
    python_file = "repo.py"
  }
}

resource "dagsterplus_team" "test" {
  name = %q
}

resource "dagsterplus_team_deployment_grant" "test" {
  team_id    = dagsterplus_team.test.id
  deployment = %q
  grant      = %q

  location_grants {
    location_name = dagsterplus_code_location.loc.name
    grant         = %q
  }
}
`, deployment, locationName, teamName, deployment, grant, locationGrant)
}
