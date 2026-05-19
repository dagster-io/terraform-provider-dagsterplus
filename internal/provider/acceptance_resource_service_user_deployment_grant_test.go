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

func TestAccServiceUserDeploymentGrantResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroyedByName(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccServiceUserDeploymentGrantConfig(rName, testAccDeployment(), "VIEWER"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_service_user_deployment_grant.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_user_deployment_grant.test", "service_user_id"),
					resource.TestCheckResourceAttr("dagsterplus_service_user_deployment_grant.test", "deployment", testAccDeployment()),
					resource.TestCheckResourceAttr("dagsterplus_service_user_deployment_grant.test", "grant", "VIEWER"),
				),
			},
			{
				Config: providerConfig() + testAccServiceUserDeploymentGrantConfig(rName, testAccDeployment(), "EDITOR"),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_service_user_deployment_grant.test", "grant", "EDITOR",
				),
			},
			{
				ResourceName:      "dagsterplus_service_user_deployment_grant.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckServiceUserDestroyedByName(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		users, err := c.ListServiceUsers(context.Background())
		if err != nil {
			return fmt.Errorf("listing service users: %w", err)
		}
		for _, u := range users {
			if u.Name == name {
				return fmt.Errorf("service user %q still exists", name)
			}
		}
		return nil
	}
}

func testAccServiceUserDeploymentGrantConfig(suName, deployment, grant string) string {
	return fmt.Sprintf(`
resource "dagsterplus_service_user" "test" {
  name = %q
}

resource "dagsterplus_service_user_deployment_grant" "test" {
  service_user_id = dagsterplus_service_user.test.id
  deployment      = %q
  grant           = %q
}
`, suName, deployment, grant)
}
