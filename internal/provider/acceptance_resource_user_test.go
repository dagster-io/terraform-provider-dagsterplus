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

func TestAccUserResource_basic(t *testing.T) {
	testEmail := testAccUserEmail(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroyed(testEmail),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccUserConfig(testEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_user.test", "email", testEmail),
					resource.TestCheckResourceAttrSet("dagsterplus_user.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_user.test", "role"),
				),
			},
			// Import
			{
				ResourceName:            "dagsterplus_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name", "role"},
			},
		},
	})
}

func testAccUserConfig(email string) string {
	return fmt.Sprintf(`
resource "dagsterplus_user" "test" {
  email = %q
}
`, email)
}

func testAccCheckUserDestroyed(email string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		users, err := c.ListUsers(context.Background())
		if err != nil {
			return fmt.Errorf("listing users: %w", err)
		}
		for _, u := range users {
			if u.Email == email {
				return fmt.Errorf("user %q still exists", email)
			}
		}
		return nil
	}
}
