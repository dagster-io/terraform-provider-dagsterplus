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

func TestAccServiceUserResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccServiceUserConfig(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_service_user.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_service_user.test", "description", "initial description"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_user.test", "id"),
				),
			},
			// Update description
			{
				Config: providerConfig() + testAccServiceUserConfig(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_service_user.test", "description", "updated description"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_service_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccServiceUserConfig(name, description string) string {
	return fmt.Sprintf(`
resource "dagsterplus_service_user" "test" {
  name        = %q
  description = %q
}
`, name, description)
}

func testAccCheckServiceUserDestroyed(name string) resource.TestCheckFunc {
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
				return fmt.Errorf("service user %q still exists with ID %q", name, u.ID)
			}
		}
		return nil
	}
}

// TestAccServiceUserResource_disappears verifies Terraform detects drift when the service user
// is deleted outside of Terraform.
func TestAccServiceUserResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceUserDestroyed(rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccServiceUserConfig(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_service_user.test", "id"),
					testAccServiceUserDisappears("dagsterplus_service_user.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccServiceUserDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteServiceUser(context.Background(), rs.Primary.ID)
	}
}
