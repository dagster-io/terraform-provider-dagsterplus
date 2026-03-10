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

func TestAccRoleResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroyedByName(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccRoleConfig(rName, "deployment", `["edit_alerts"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_role.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_role.test", "role_type", "deployment"),
					resource.TestCheckResourceAttrSet("dagsterplus_role.test", "id"),
				),
			},
			// Update permissions
			{
				Config: providerConfig() + testAccRoleConfig(rName, "deployment", `["edit_alerts", "delete_runs"]`),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_role.test", "permissions.#", "2",
				),
			},
			// Import by role ID
			{
				ResourceName:      "dagsterplus_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoleConfig(name, roleType, permissions string) string {
	return fmt.Sprintf(`
resource "dagsterplus_role" "test" {
  name        = %q
  role_type   = %q
  permissions = %s
}
`, name, roleType, permissions)
}

func testAccCheckRoleDestroyedByName(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		roles, err := c.ListRoles(context.Background())
		if err != nil {
			return nil
		}
		for _, r := range roles {
			if r.Name == name {
				return fmt.Errorf("role %q still exists with ID %q", name, r.ID)
			}
		}
		return nil
	}
}
