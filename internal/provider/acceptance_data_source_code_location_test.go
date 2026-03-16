package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCodeLocationDataSource_basic(t *testing.T) {
	rDep := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)
	rLoc := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccCodeLocationDataSourceConfig(
					rDep, rLoc, "ghcr.io/example/repo:latest", "repo.py",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "deployment", rDep),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "name", rLoc),
					resource.TestCheckResourceAttr("data.dagsterplus_code_location.test", "image", "ghcr.io/example/repo:latest"),
					// code_source.python_file is not asserted — the API does not reliably return code_source via list endpoint.
					resource.TestCheckResourceAttrSet("data.dagsterplus_code_location.test", "id"),
				),
			},
		},
	})
}

func TestAccCodeLocationDataSource_notFound(t *testing.T) {
	rDep := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      providerConfig() + testAccCodeLocationDataSourceNotFoundConfig(rDep),
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
  name       = %q
  image      = %q

  code_source {
    python_file = %q
  }
}

data "dagsterplus_code_location" "test" {
  deployment = dagsterplus_code_location.test.deployment
  name       = dagsterplus_code_location.test.name
  depends_on = [dagsterplus_code_location.test]
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
  name       = "this-location-does-not-exist-xyz"
  depends_on = [dagsterplus_deployment.test]
}
`, deployment)
}
