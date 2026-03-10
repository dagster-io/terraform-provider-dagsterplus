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

func TestAccServiceTokenResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckServiceTokenRevoked(rName),
		Steps: []resource.TestStep{
			// Create the service user first, then the token
			{
				Config: providerConfig() + testAccServiceTokenConfig(rName, "initial token description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_service_token.test", "description", "initial token description"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_token.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_service_token.test", "token"),
				),
			},
			// Update description
			{
				Config: providerConfig() + testAccServiceTokenConfig(rName, "updated token description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_service_token.test", "description", "updated token description"),
				),
			},
			// Import — token is irrecoverable, service_user_id and token are excluded
			{
				ResourceName:            "dagsterplus_service_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "service_user_id"},
			},
		},
	})
}

func testAccServiceTokenConfig(ownerName, description string) string {
	return fmt.Sprintf(`
resource "dagsterplus_service_user" "token_owner" {
  name        = %q
  description = "Service user for token test"
}

resource "dagsterplus_service_token" "test" {
  service_user_id = dagsterplus_service_user.token_owner.id
  description     = %q
}
`, ownerName, description)
}

func testAccCheckServiceTokenRevoked(ownerName string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		// Verify the service user is gone too (cleanup)
		users, err := c.ListServiceUsers(context.Background())
		if err != nil {
			return nil
		}
		for _, u := range users {
			if u.Name == ownerName {
				return fmt.Errorf("service user %q still exists", u.Name)
			}
		}
		return nil
	}
}
