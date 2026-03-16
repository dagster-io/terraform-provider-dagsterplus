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
					resource.TestCheckResourceAttr("dagsterplus_role.test", "permissions.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_role.test", "description", ""),
					resource.TestCheckResourceAttr("dagsterplus_role.test", "icon", ""),
					resource.TestCheckResourceAttrSet("dagsterplus_role.test", "id"),
				),
			},
			// No-op plan after create
			{
				Config:             providerConfig() + testAccRoleConfig(rName, "deployment", `["edit_alerts"]`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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

func TestAccRoleResource_organizationType(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroyedByName(rName),
		Steps: []resource.TestStep{
			// Create an organization-scoped role with org-only permissions
			{
				Config: providerConfig() + testAccRoleConfig(rName, "organization", `["manage_full_deployments"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_role.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_role.test", "role_type", "organization"),
					resource.TestCheckResourceAttr("dagsterplus_role.test", "permissions.#", "1"),
					resource.TestCheckResourceAttrSet("dagsterplus_role.test", "id"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccRoleResource_disappears verifies Terraform detects drift when the role is
// deleted outside of Terraform.
func TestAccRoleResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroyedByName(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccRoleConfig(rName, "deployment", `["edit_alerts"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_role.test", "id"),
					testAccRoleDisappears("dagsterplus_role.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoleDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteRole(context.Background(), rs.Primary.ID)
	}
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
			return fmt.Errorf("listing roles: %w", err)
		}
		for _, r := range roles {
			if r.Name == name {
				return fmt.Errorf("role %q still exists with ID %q", name, r.ID)
			}
		}
		return nil
	}
}
