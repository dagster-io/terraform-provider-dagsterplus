package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCustomMetricDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccCustomMetricDataSourceConfig("acc_tf_ds_metric", "Acc TF DS Metric", "ds test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_custom_metric.test", "metadata_key", "acc_tf_ds_metric"),
					resource.TestCheckResourceAttr("data.dagsterplus_custom_metric.test", "metadata_key", "acc_tf_ds_metric"),
					resource.TestCheckResourceAttr("data.dagsterplus_custom_metric.test", "display_name", "Acc TF DS Metric"),
					resource.TestCheckResourceAttr("data.dagsterplus_custom_metric.test", "description", "ds test description"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_custom_metric.test", "id"),
				),
			},
		},
	})
}

func TestAccCustomMetricDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccCustomMetricDataSourceNotFoundConfig(),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccCustomMetricDataSourceConfig(metadataKey, displayName, description string) string {
	return fmt.Sprintf(`
resource "dagsterplus_custom_metric" "test" {
  metadata_key = %q
  display_name = %q
  description  = %q
}

data "dagsterplus_custom_metric" "test" {
  metadata_key = dagsterplus_custom_metric.test.metadata_key
  depends_on   = [dagsterplus_custom_metric.test]
}
`, metadataKey, displayName, description)
}

func testAccCustomMetricDataSourceNotFoundConfig() string {
	return `
data "dagsterplus_custom_metric" "test" {
  metadata_key = "this_metric_does_not_exist_xyz"
}
`
}
