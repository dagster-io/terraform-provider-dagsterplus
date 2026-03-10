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

func TestAccSecretResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSecretDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccSecretConfig(rName, "initial-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_secret.test", "secret_name", rName),
					resource.TestCheckResourceAttr("dagsterplus_secret.test", "secret_value", "initial-value"),
					resource.TestCheckResourceAttrSet("dagsterplus_secret.test", "id"),
				),
			},
			// Update secret value
			{
				Config: providerConfig() + testAccSecretConfig(rName, "updated-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_secret.test", "secret_value", "updated-value"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_secret.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSecretConfig(name, value string) string {
	return fmt.Sprintf(`
resource "dagsterplus_secret" "test" {
  secret_name          = %q
  secret_value         = %q
  full_deployment_scope = true
}
`, name, value)
}

func testAccCheckSecretDestroyed(name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		secrets, err := c.ListSecrets(context.Background())
		if err != nil {
			return nil
		}
		for _, s := range secrets {
			if s.SecretName == name {
				return fmt.Errorf("secret %q still exists with ID %q", name, s.ID)
			}
		}
		return nil
	}
}
