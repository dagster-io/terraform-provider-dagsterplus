package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccAlertPolicyDataSource_basic creates an asset alert policy via the resource, then reads
// it back via the data source and verifies the attributes match.
func TestAccAlertPolicyDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccAlertPolicyDataSourceConfig(
					"acc-tf-ds-alert", testAccAlertDeployment,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Resource attributes
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", "acc-tf-ds-alert"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					// Data source attributes match the resource
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test", "name", "acc-tf-ds-alert"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test", "asset.0.all_assets", "true"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test", "asset.0.health_status", "degraded"),
					resource.TestCheckResourceAttr("data.dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_alert_policy.test", "id"),
				),
			},
		},
	})
}

// TestAccAlertPolicyDataSource_notFound verifies that looking up a nonexistent policy errors.
func TestAccAlertPolicyDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccAlertPolicyDataSourceNotFoundConfig(testAccAlertDeployment),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccAlertPolicyDataSourceConfig(name, deployment string) string {
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "asset"
  enabled     = true

  asset {
    all_assets    = true
    health_status = "degraded"
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}

data "dagsterplus_alert_policy" "test" {
  deployment  = dagsterplus_alert_policy.test.deployment
  name        = dagsterplus_alert_policy.test.name
  policy_type = "asset"
  depends_on  = [dagsterplus_alert_policy.test]
}
`, name, deployment)
}

func testAccAlertPolicyDataSourceNotFoundConfig(deployment string) string {
	return fmt.Sprintf(`
data "dagsterplus_alert_policy" "test" {
  deployment  = %q
  name        = "this-policy-does-not-exist-xyz"
  policy_type = "asset"
}
`, deployment)
}
