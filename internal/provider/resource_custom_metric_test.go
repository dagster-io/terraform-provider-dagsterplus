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

func TestAccCustomMetricResource_basic(t *testing.T) {
	// metadata_key uses underscores (not hyphens) per API requirements.
	rName := "acc_tf_" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCustomMetricDestroyed(rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccCustomMetricConfig(rName, "Acc TF Test Metric", "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "metadata_key", rName),
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "display_name", "Acc TF Test Metric"),
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "description", "initial description"),
					resource.TestCheckResourceAttrSet("dagsterplus_custom_metric.test", "id"),
				),
			},
			// Update display_name and description
			{
				Config: providerConfig() + testAccCustomMetricConfig(rName, "Updated Metric", "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "display_name", "Updated Metric"),
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "description", "updated description"),
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_custom_metric.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCustomMetricConfig(metadataKey, displayName, description string) string {
	return fmt.Sprintf(`
resource "dagsterplus_custom_metric" "test" {
  metadata_key = %q
  display_name = %q
  description  = %q
}
`, metadataKey, displayName, description)
}

func testAccCheckCustomMetricDestroyed(metadataKey string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		metrics, err := c.ListCustomMetrics(context.Background())
		if err != nil {
			return nil
		}
		for _, m := range metrics {
			if m.MetadataKey == metadataKey {
				return fmt.Errorf("custom metric %q still exists with ID %q", metadataKey, m.ID)
			}
		}
		return nil
	}
}
