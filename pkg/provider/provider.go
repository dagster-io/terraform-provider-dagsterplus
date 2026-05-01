package provider

import (
	internalprovider "github.com/dagster-io/terraform-provider-dagsterplus/internal/provider"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
)

func New(version string) func() tfprovider.Provider {
	return internalprovider.New(version)
}
