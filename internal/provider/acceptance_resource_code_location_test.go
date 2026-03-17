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

func TestAccCodeLocationResource_basic(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed(testAccDeployment(), rName),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig() + testAccCodeLocationConfig(
					testAccDeployment(), rName, "ghcr.io/example/repo:v1", "repo.py",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "deployment", testAccDeployment()),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "name", rName),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "image", "ghcr.io/example/repo:v1"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "code_source.python_file", "repo.py"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "working_directory", "/app"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "executable_path", "/usr/bin/python3"),
					resource.TestCheckResourceAttrSet("dagsterplus_code_location.test", "id"),
				),
			},
			// Update image in-place
			{
				Config: providerConfig() + testAccCodeLocationConfig(
					testAccDeployment(), rName, "ghcr.io/example/repo:v2", "repo.py",
				),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_code_location.test", "image", "ghcr.io/example/repo:v2",
				),
			},
			// Import — the API serialized metadata does not reliably return code_source/path fields.
			{
				ResourceName:      "dagsterplus_code_location.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"code_source.python_file",
					"working_directory",
					"executable_path",
					"attribute",
					"agent_queue",
					"git",
				},
			},
		},
	})
}

func TestAccCodeLocationResource_packageName(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed(testAccDeployment(), rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment = %q
  name       = %q
  image      = "ghcr.io/example/repo:latest"

  code_source {
    package_name = "my_dagster_package"
  }
}
`, testAccDeployment(), rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "code_source.package_name", "my_dagster_package"),
				),
			},
		},
	})
}

// TestAccCodeLocationResource_agentQueue tests that agent_queue and attribute are
// persisted and read back correctly.
func TestAccCodeLocationResource_agentQueue(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed(testAccDeployment(), rName),
		Steps: []resource.TestStep{
			// Create with agent_queue and attribute
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment  = %q
  name        = %q
  image       = "ghcr.io/example/repo:latest"
  agent_queue = "default"
  attribute   = "my_defs"

  code_source {
    python_file = "repo.py"
  }
}
`, testAccDeployment(), rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "agent_queue", "default"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "attribute", "my_defs"),
				),
			},
			// Update attribute value
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment  = %q
  name        = %q
  image       = "ghcr.io/example/repo:latest"
  agent_queue = "default"
  attribute   = "updated_defs"

  code_source {
    python_file = "repo.py"
  }
}
`, testAccDeployment(), rName),
				Check: resource.TestCheckResourceAttr(
					"dagsterplus_code_location.test", "attribute", "updated_defs",
				),
			},
		},
	})
}

// TestAccCodeLocationResource_git tests a git-based code location (no image).
func TestAccCodeLocationResource_git(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed(testAccDeployment(), rName),
		Steps: []resource.TestStep{
			// Create with git block instead of image
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment = %q
  name       = %q

  git {
    commit_hash = "abc123def456"
    url         = "https://github.com/example/repo"
  }

  code_source {
    python_file = "repo.py"
  }
}
`, testAccDeployment(), rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "git.commit_hash", "abc123def456"),
					resource.TestCheckResourceAttr("dagsterplus_code_location.test", "git.url", "https://github.com/example/repo"),
					resource.TestCheckNoResourceAttr("dagsterplus_code_location.test", "image"),
				),
			},
		},
	})
}

// TestAccCodeLocationResource_disappears verifies Terraform detects drift when the
// code location is deleted outside of Terraform.
func TestAccCodeLocationResource_disappears(t *testing.T) {
	rName := "acc-tf-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCodeLocationDestroyed(testAccDeployment(), rName),
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + testAccCodeLocationConfig(
					testAccDeployment(), rName, "ghcr.io/example/repo:v1", "repo.py",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dagsterplus_code_location.test", "id"),
					testAccCodeLocationDisappears("dagsterplus_code_location.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccCodeLocationDisappears deletes the code location out-of-band during a Check step.
func testAccCodeLocationDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteCodeLocation(context.Background(), deployment, name)
	}
}

// testAccCodeLocationConfig returns HCL for a code location resource using a python_file source.
func testAccCodeLocationConfig(deployment, name, image, pythonFile string) string {
	return fmt.Sprintf(`
resource "dagsterplus_code_location" "test" {
  deployment = %q
  name       = %q
  image      = %q

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
