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

func TestAccConcurrencyPoolResource_basic(t *testing.T) {
	poolName := "acc-tf-pool"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckConcurrencyPoolDestroy(testAccDeployment(), poolName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccConcurrencyPoolConfig(testAccDeployment(), poolName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_concurrency_pool.test", "deployment", testAccDeployment()),
					resource.TestCheckResourceAttr("dagsterplus_concurrency_pool.test", "name", poolName),
					resource.TestCheckResourceAttr("dagsterplus_concurrency_pool.test", "limit", "1"),
					resource.TestCheckResourceAttr("dagsterplus_concurrency_pool.test", "id", testAccDeployment()+"/"+poolName),
				),
			},
			// Update the limit
			{
				Config: providerConfig() + testAccConcurrencyPoolConfig(testAccDeployment(), poolName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_concurrency_pool.test", "limit", "10"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_concurrency_pool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccConcurrencyPoolConfig(deployment, name string, limit int) string {
	return fmt.Sprintf(`
resource "dagsterplus_concurrency_pool" "test" {
  deployment = %q
  name       = %q
  limit      = %d
}
`, deployment, name, limit)
}

func testAccCheckConcurrencyPoolDestroy(deployment, name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		pool, err := c.GetConcurrencyPool(context.Background(), deployment, name)
		if err != nil {
			return err
		}
		if pool.IsSet {
			return fmt.Errorf("concurrency pool %q still has an explicit limit after destroy", name)
		}
		return nil
	}
}
