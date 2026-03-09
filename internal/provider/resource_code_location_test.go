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

func TestAccCodeLocationResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed("acc-tf-test", "acc-tf-location"),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccCodeLocationConfig(
					"acc-tf-test", "acc-tf-location", "ghcr.io/example/repo:v1", "repo.py",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "deployment_name", "acc-tf-test"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "name", "acc-tf-location"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "image", "ghcr.io/example/repo:v1"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "code_source.python_file", "repo.py"),
					resource.TestCheckResourceAttrSet("dagsterplus_code_location.test", "id"),
				),
			},
			// Update image in-place
			{
				Config: providerConfig() + testAccCodeLocationConfig(
					"acc-tf-test", "acc-tf-location", "ghcr.io/example/repo:v2", "repo.py",
				),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_code_location.test", "image", "ghcr.io/example/repo:v2",
				),
			},
			// Import
			{
				ResourceName:      "dagsterplus_code_location.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeLocationResource_packageName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed("acc-tf-test", "acc-tf-pkg-location"),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment_name = "acc-tf-test"
  name            = "acc-tf-pkg-location"
  image           = "ghcr.io/example/repo:latest"

  code_source {
    package_name = "my_dagster_package"
  }
}
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "code_source.package_name", "my_dagster_package"),
				),
			},
		},
	})
}

// testAccCodeLocationConfig returns HCL for a code location resource using a python_file source.
func testAccCodeLocationConfig(deployment, name, image, pythonFile string) string {
	return fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment_name = %q
  name            = %q
  image           = %q

  code_source {
    python_file = %q
  }

  working_directory = "/app"
  executable_path   = "/usr/bin/python3"
}
`, deployment, name, image, pythonFile)
}

// testAccCheckCodeLocationDestroyed verifies the code location no longer exists.
func testAccCheckCodeLocationDestroyed(deployment, name string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		c := client.New(
			os.Getenv("DAGSTER_CLOUD_ORGANIZATION"),
			os.Getenv("DAGSTER_CLOUD_API_TOKEN"),
			"",
		)
		_, err := c.GetCodeLocation(context.Background(), deployment, name)
		if err == nil {
			return fmt.Errorf("code location %q in %q still exists", name, deployment)
		}
		return nil
	}
}
