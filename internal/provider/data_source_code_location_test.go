package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCodeLocationDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccCodeLocationDataSourceConfig(
					"acc-tf-ds-dep", "acc-tf-ds-loc", "ghcr.io/example/repo:latest", "repo.py",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "deployment", "acc-tf-ds-dep"),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "name", "acc-tf-ds-loc"),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "image", "ghcr.io/example/repo:latest"),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "code_source.python_file", "repo.py"),
					resource.TestCheckResourceAttrSet("data.dagsterplus_code_location.test", "id"),
				),
			},
		},
	})
}

func TestAccCodeLocationDataSource_notFound(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccCodeLocationDataSourceNotFoundConfig("acc-tf-ds-dep"),
				ExpectError: errRegexp("not found"),
			},
		},
	})
}

func testAccCodeLocationDataSourceConfig(deployment, name, image, pythonFile string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment" "test" {
  name = %q

}

resource "dagsterplus_code_location" "test" {
  deployment = dagsterplus_deployment.test.name
  name            = %q
  image           = %q

  code_source {
    python_file = %q
  }
}

data "dagsterplus_code_location" "test" {
  deployment = dagsterplus_code_location.test.deployment
  name            = dagsterplus_code_location.test.name
  depends_on      = [dagsterplus_code_location.test]
}
`, deployment, name, image, pythonFile)
}

func testAccCodeLocationDataSourceNotFoundConfig(deployment string) string {
	return fmt.Sprintf(`
resource "dagsterplus_deployment" "test" {
  name = %q

}

data "dagsterplus_code_location" "test" {
  deployment = dagsterplus_deployment.test.name
  name            = "this-location-does-not-exist-xyz"
  depends_on      = [dagsterplus_deployment.test]
}
`, deployment)
}
