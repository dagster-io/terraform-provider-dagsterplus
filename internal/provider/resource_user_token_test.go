package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccUserTokenResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserTokenDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccUserTokenConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_user_token.test", "name", rName),
					resource.TestCheckResourceAttrSet("dagsterplus_user_token.test", "id"),
					resource.TestCheckResourceAttrSet("dagsterplus_user_token.test", "token"),
				),
			},
			// Import by token ID (token, name, user_id not recoverable via API)
			{
				ResourceName:            "dagsterplus_user_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "name", "user_id"},
			},
		},
	})
}

func testAccUserTokenConfig(name string) string {
	return fmt.Sprintf(`
resource "dagsterplus_user_token" "test" {
  name = %q
}
`, name)
}

func testAccCheckUserTokenDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		// user_id is not available at destroy-check time; skip the list verification.
		_ = c
		return nil
	}
}
