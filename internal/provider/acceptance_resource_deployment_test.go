package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeploymentResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccDeploymentConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "name", rName),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "status"),
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "deployment_id"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_deployment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDeploymentResource_agentType verifies agent_type is honoured on create
// and switched in place (no replacement) on update.
func TestAccDeploymentResource_agentType(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroyed(rName),
		Steps: []resource.TestStep{
			// Create with an explicit agent type.
			{
				Config: providerConfig() + testAccDeploymentAgentTypeConfig(rName, "SERVERLESS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "agent_type", "SERVERLESS"),
				),
			},
			// Switch to hybrid in place.
			{
				Config: providerConfig() + testAccDeploymentAgentTypeConfig(rName, "HYBRID"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dagsterplus_deployment.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_deployment.test", "agent_type", "HYBRID"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_deployment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccDeploymentResource_disappears verifies Terraform detects drift when the
// deployment is deleted outside of Terraform.
func TestAccDeploymentResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroyed(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccDeploymentConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_deployment.test", "id"),
					testAccDeploymentDisappears("dagsterplus_deployment.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccDeploymentDisappears deletes the deployment out-of-band during a Check step.
func testAccDeploymentDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		name := rs.Primary.Attributes["name"]
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		return c.DeleteDeployment(context.Background(), name)
	}
}

// testAccDeploymentConfig returns HCL for a deployment resource.
func testAccDeploymentConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment" "test" {
  name = %q
}
`, name)
}

// testAccDeploymentAgentTypeConfig returns HCL for a deployment with an explicit agent type.
func testAccDeploymentAgentTypeConfig(name, agentType string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment" "test" {
  name       = %q
  agent_type = %q
}
`, name, agentType)
}

// testAccCheckDeploymentDestroyed verifies the deployment no longer exists.
func testAccCheckDeploymentDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		_, err := c.GetDeployment(context.Background(), name)
		if err == nil {
			return fmt.Errorf("deployment %q still exists", name)
		}
		return nil
	}
}
