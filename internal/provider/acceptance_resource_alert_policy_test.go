package provider_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccAlertDeployment is the deployment used for alert policy and related acceptance tests.
// Reads from DAGSTER_CLOUD_TEST_DEPLOYMENT; falls back to "prod".
var testAccAlertDeployment = testAccDeployment()

// TestAccAlertPolicyResource_assetHealthStatus tests create, update, and destroy for an
// asset alert policy driven by health status transitions.
func TestAccAlertPolicyResource_assetHealthStatus(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: all assets, degraded health status, enabled
			{
				Config: providerConfig() + testAccAlertPolicyHealthStatusConfig(
					rName, testAccAlertDeployment, true, "degraded",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "asset"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.all_assets", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.health_status", "degraded"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "event_types.0", "ASSET_HEALTH_DEGRADED"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Update: change health_status to warning
			{
				Config: providerConfig() + testAccAlertPolicyHealthStatusConfig(
					rName, testAccAlertDeployment, true, "warning",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.health_status", "warning"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "event_types.0", "ASSET_HEALTH_WARNING"),
				),
			},
			// Update: disable the policy
			{
				Config: providerConfig() + testAccAlertPolicyHealthStatusConfig(
					rName, testAccAlertDeployment, false, "warning",
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "false"),
			},
			// Import by {deployment}/{name}
			{
				ResourceName:      "dagsterplus_alert_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				// policy_type is not returned by the API so it cannot be verified after import.
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

// TestAccAlertPolicyResource_assetSpecificEvents tests create, update, and destroy for an
// asset alert policy driven by specific event types.
func TestAccAlertPolicyResource_assetSpecificEvents(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: single specific event
			{
				Config: providerConfig() + testAccAlertPolicySpecificEventsConfig(
					rName, testAccAlertDeployment, true,
					[]string{"materialization_success"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "asset"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.all_assets", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.specific_events.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.specific_events.0", "materialization_success"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Update: add a second event
			{
				Config: providerConfig() + testAccAlertPolicySpecificEventsConfig(
					rName, testAccAlertDeployment, true,
					[]string{"materialization_success", "materialization_failure"},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.specific_events.#", "2"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.specific_events.0", "materialization_success"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "asset.0.specific_events.1", "materialization_failure"),
				),
			},
			// Update: disable
			{
				Config: providerConfig() + testAccAlertPolicySpecificEventsConfig(
					rName, testAccAlertDeployment, false,
					[]string{"materialization_success", "materialization_failure"},
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "false"),
			},
			// Import by {deployment}/{name}
			{
				ResourceName:            "dagsterplus_alert_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

// TestAccAlertPolicyResource_run tests create, update, and destroy for a run alert policy.
func TestAccAlertPolicyResource_run(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: all runs, on_failure only
			{
				Config: providerConfig() + testAccAlertPolicyRunConfig(
					rName, testAccAlertDeployment, true, true, false,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "run"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "run.0.all_runs", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "run.0.on_failure", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "run.0.on_success", "false"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "event_types.0", "JOB_FAILURE"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Update: also on_success
			{
				Config: providerConfig() + testAccAlertPolicyRunConfig(
					rName, testAccAlertDeployment, true, true, true,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "run.0.on_success", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "run.0.on_failure", "true"),
				),
			},
			// Update: disable
			{
				Config: providerConfig() + testAccAlertPolicyRunConfig(
					rName, testAccAlertDeployment, false, true, true,
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "false"),
			},
			// Import
			{
				ResourceName:            "dagsterplus_alert_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

// TestAccAlertPolicyResource_codeLocation tests create and destroy for a code_location alert policy.
func TestAccAlertPolicyResource_codeLocation(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: all locations
			{
				Config: providerConfig() + testAccAlertPolicyCodeLocationConfig(
					rName, testAccAlertDeployment, true,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "code_location"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "code_location.0.all_locations", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "event_types.0", "CODE_LOCATION_ERROR"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Import
			{
				ResourceName:            "dagsterplus_alert_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

// TestAccAlertPolicyResource_automation tests create, update, and destroy for an automation alert policy.
func TestAccAlertPolicyResource_automation(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: all schedules and sensors
			{
				Config: providerConfig() + testAccAlertPolicyAutomationConfig(
					rName, testAccAlertDeployment, true, 0,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "automation"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "automation.0.all_schedules_and_sensors", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "automation.0.include_schedules", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "automation.0.include_sensors", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "event_types.0", "TICK_FAILURE"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Update: add min_consecutive_failures
			{
				Config: providerConfig() + testAccAlertPolicyAutomationConfig(
					rName, testAccAlertDeployment, true, 2,
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "automation.0.min_consecutive_failures", "2"),
			},
			// Import
			{
				ResourceName:            "dagsterplus_alert_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

// TestAccAlertPolicyResource_budget tests create, update, and destroy for a budget alert policy.
func TestAccAlertPolicyResource_budget(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: greater_than threshold
			{
				Config: providerConfig() + testAccAlertPolicyBudgetConfig(
					rName, testAccAlertDeployment, true, "greater_than", 500.0, 30,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "budget"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "budget.0.operator", "greater_than"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "budget.0.threshold", "500"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "budget.0.period_days", "30"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Update: raise threshold
			{
				Config: providerConfig() + testAccAlertPolicyBudgetConfig(
					rName, testAccAlertDeployment, true, "greater_than", 1000.0, 30,
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "budget.0.threshold", "1000"),
			},
			// Update: disable
			{
				Config: providerConfig() + testAccAlertPolicyBudgetConfig(
					rName, testAccAlertDeployment, false, "greater_than", 1000.0, 30,
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "false"),
			},
			// Import
			{
				ResourceName:            "dagsterplus_alert_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"policy_type"},
			},
		},
	})
}

// TestAccAlertPolicyResource_insightMetric tests create, update, and destroy for an insight_metric alert policy.
func TestAccAlertPolicyResource_insightMetric(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			// Create: deployment-wide metric threshold
			{
				Config: providerConfig() + testAccAlertPolicyInsightMetricConfig(
					rName, testAccAlertDeployment, true, "dagster_credits", "greater_than", 100.0, 7,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "deployment", testAccAlertDeployment),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "policy_type", "insight_metric"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "insight_metric.0.metric", "dagster_credits"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "insight_metric.0.operator", "greater_than"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "insight_metric.0.threshold", "100"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "insight_metric.0.period_days", "7"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.type", "email"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.#", "1"),
					resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "notification_service.email_addresses.0", "acc-test@example.com"),
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
				),
			},
			// Update: increase threshold
			{
				Config: providerConfig() + testAccAlertPolicyInsightMetricConfig(
					rName, testAccAlertDeployment, true, "dagster_credits", "greater_than", 250.0, 7,
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "insight_metric.0.threshold", "250"),
			},
			// Update: disable
			{
				Config: providerConfig() + testAccAlertPolicyInsightMetricConfig(
					rName, testAccAlertDeployment, false, "dagster_credits", "greater_than", 250.0, 7,
				),
				Check: resource.TestCheckResourceAttr("dagsterplus_alert_policy.test", "enabled", "false"),
			},
			// Import — deployment-level insight_metric policies are indistinguishable
			// from budget policies in the API response, so import infers "budget".
			// Ignore all block-specific fields that differ between the two schemas.
			{
				ResourceName:      "dagsterplus_alert_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"policy_type",
					"insight_metric.#", "insight_metric.0.%",
					"insight_metric.0.metric", "insight_metric.0.operator",
					"insight_metric.0.threshold", "insight_metric.0.period_days",
					"insight_metric.0.asset_group", "insight_metric.0.asset_key", "insight_metric.0.job_name",
					"budget.#", "budget.0.%",
					"budget.0.operator", "budget.0.threshold", "budget.0.period_days",
				},
			},
		},
	})
}

// TestAccAlertPolicyResource_disappears verifies Terraform detects drift when the
// policy is deleted outside of Terraform.
func TestAccAlertPolicyResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAlertPolicyDestroyed(testAccAlertDeployment, rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccAlertPolicyHealthStatusConfig(
					rName, testAccAlertDeployment, true, "degraded",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_alert_policy.test", "id"),
					testAccAlertPolicyDisappears("dagsterplus_alert_policy.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccAlertPolicyResource_invalidPolicyType verifies the schema validator rejects
// an unknown policy_type at plan time.
func TestAccAlertPolicyResource_invalidPolicyType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "dagsterplus_alert_policy" "test" {
  name        = "acc-tf-invalid"
  deployment  = "prod"
  policy_type = "not_a_real_type"
  enabled     = true

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`,
				ExpectError: errRegexp(`value must be one of`),
			},
		},
	})
}

// TestAccAlertPolicyResource_invalidBudgetOperator verifies the schema validator rejects
// an unknown budget operator at plan time.
func TestAccAlertPolicyResource_invalidBudgetOperator(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "dagsterplus_alert_policy" "test" {
  name        = "acc-tf-invalid-budget"
  deployment  = "prod"
  policy_type = "budget"
  enabled     = true

  budget {
    operator    = "not_an_operator"
    threshold   = 100
    period_days = 30
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`,
				ExpectError: errRegexp(`value must be one of`),
			},
		},
	})
}

// TestAccAlertPolicyResource_conflictingAssetFields verifies the ConflictsWith validator
// rejects health_status and specific_events being set simultaneously.
func TestAccAlertPolicyResource_conflictingAssetFields(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "dagsterplus_alert_policy" "test" {
  name        = "acc-tf-invalid-asset"
  deployment  = "prod"
  policy_type = "asset"
  enabled     = true

  asset {
    all_assets      = true
    health_status   = "degraded"
    specific_events = ["materialization_success"]
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`,
				ExpectError: errRegexp(`cannot be specified`),
			},
		},
	})
}

// testAccAlertPolicyDisappears deletes the alert policy out-of-band during a Check step,
// causing Terraform to detect drift on the subsequent implicit plan.
func testAccAlertPolicyDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		deployment := rs.Primary.Attributes["deployment"]
		name := rs.Primary.Attributes["name"]
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		return c.DeleteAlertPolicy(context.Background(), deployment, name)
	}
}

// testAccAlertPolicyHealthStatusConfig returns HCL for an asset alert policy using health_status.
func testAccAlertPolicyHealthStatusConfig(name, deployment string, enabled bool, healthStatus string) string {
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "asset"
  enabled     = %t

  asset {
    all_assets    = true
    health_status = %q
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled, healthStatus)
}

// testAccAlertPolicySpecificEventsConfig returns HCL for an asset alert policy using specific_events.
func testAccAlertPolicySpecificEventsConfig(name, deployment string, enabled bool, events []string) string {
	quoted := make([]string, len(events))
	for i, e := range events {
		quoted[i] = fmt.Sprintf("%q", e)
	}
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "asset"
  enabled     = %t

  asset {
    all_assets      = true
    specific_events = [%s]
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled, strings.Join(quoted, ", "))
}

// testAccCheckAlertPolicyDestroyed verifies the alert policy no longer exists.
func testAccCheckAlertPolicyDestroyed(deployment, name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		_, err := c.GetAlertPolicy(context.Background(), deployment, name)
		if err == nil {
			return fmt.Errorf("alert policy %q in deployment %q still exists", name, deployment)
		}
		return nil
	}
}

func testAccAlertPolicyRunConfig(name, deployment string, enabled, onFailure, onSuccess bool) string {
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "run"
  enabled     = %t

  run {
    all_runs   = true
    on_failure = %t
    on_success = %t
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled, onFailure, onSuccess)
}

func testAccAlertPolicyCodeLocationConfig(name, deployment string, enabled bool) string {
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "code_location"
  enabled     = %t

  code_location {
    all_locations = true
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled)
}

func testAccAlertPolicyAutomationConfig(name, deployment string, enabled bool, minConsecutiveFailures int) string {
	minFailures := ""
	if minConsecutiveFailures > 0 {
		minFailures = fmt.Sprintf("    min_consecutive_failures = %d\n", minConsecutiveFailures)
	}
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "automation"
  enabled     = %t

  automation {
    all_schedules_and_sensors = true
    include_schedules         = true
    include_sensors           = true
%s  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled, minFailures)
}

func testAccAlertPolicyBudgetConfig(name, deployment string, enabled bool, operator string, threshold float64, periodDays int) string {
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "budget"
  enabled     = %t

  budget {
    operator    = %q
    threshold   = %g
    period_days = %d
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled, operator, threshold, periodDays)
}

func testAccAlertPolicyInsightMetricConfig(name, deployment string, enabled bool, metric, operator string, threshold float64, periodDays int) string {
	return fmt.Sprintf(`
resource "dagsterplus_alert_policy" "test" {
  name        = %q
  deployment  = %q
  policy_type = "insight_metric"
  enabled     = %t

  insight_metric {
    metric      = %q
    operator    = %q
    threshold   = %g
    period_days = %d
  }

  notification_service {
    type            = "email"
    email_addresses = ["acc-test@example.com"]
  }
}
`, name, deployment, enabled, metric, operator, threshold, periodDays)
}
