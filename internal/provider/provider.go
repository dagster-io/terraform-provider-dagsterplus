package provider

import (
	"context"
	"os"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/dagster-io/terraform-provider-dagsterplus/internal/datasources"
	"github.com/dagster-io/terraform-provider-dagsterplus/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure dagsterplusProvider satisfies the provider.Provider interface.
var _ provider.Provider = &dagsterplusProvider{}
var _ provider.ProviderWithFunctions = &dagsterplusProvider{}

// dagsterplusProvider defines the provider implementation.
type dagsterplusProvider struct {
	version string
}

// dagsterplusProviderModel describes the provider data model.
type dagsterplusProviderModel struct {
	Organization types.String `tfsdk:"organization"`
	APIToken     types.String `tfsdk:"api_token"`
	BaseURL      types.String `tfsdk:"base_url"`
}

// New returns a provider.Provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &dagsterplusProvider{version: version}
	}
}

func (p *dagsterplusProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dagsterplus"
	resp.Version = p.version
}

func (p *dagsterplusProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Dagster+ (Dagster Cloud) to manage deployments, code locations, and teams.",
		Attributes: map[string]schema.Attribute{
			"organization": schema.StringAttribute{
				Description: "The Dagster+ organization name (e.g. `my-org` for `my-org.dagster.cloud`). " +
					"May also be set via the `DAGSTER_CLOUD_ORGANIZATION` environment variable.",
				Optional: true,
			},
			"api_token": schema.StringAttribute{
				Description: "A Dagster+ API token. " +
					"May also be set via the `DAGSTER_CLOUD_API_TOKEN` environment variable. " +
					"Generate a token at Dagster+ → Account Settings → API Tokens.",
				Optional:  true,
				Sensitive: true,
			},
			"base_url": schema.StringAttribute{
				Description: "Override the base URL for the Dagster+ API. " +
					"Defaults to `https://{organization}.dagster.cloud`. " +
					"Only needed for on-prem or custom deployments.",
				Optional: true,
			},
		},
	}
}

func (p *dagsterplusProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config dagsterplusProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	organization := os.Getenv("DAGSTER_CLOUD_ORGANIZATION")
	if !config.Organization.IsNull() {
		organization = config.Organization.ValueString()
	}

	apiToken := os.Getenv("DAGSTER_CLOUD_API_TOKEN")
	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}

	baseURL := ""
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	if organization == "" {
		resp.Diagnostics.AddError(
			"Missing organization",
			"The 'organization' attribute or DAGSTER_CLOUD_ORGANIZATION environment variable must be set.",
		)
	}

	if apiToken == "" {
		resp.Diagnostics.AddError(
			"Missing api_token",
			"The 'api_token' attribute or DAGSTER_CLOUD_API_TOKEN environment variable must be set.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(organization, apiToken, baseURL)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *dagsterplusProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewDeploymentResource,
		resources.NewCodeLocationResource,
		resources.NewTeamResource,
		resources.NewUserResource,
		resources.NewAlertPolicyResource,
		resources.NewAgentTokenResource,
		resources.NewAgentTokenDeploymentGrantResource,
		resources.NewUserTokenResource,
		resources.NewRoleResource,
		resources.NewScimSettingsResource,
		resources.NewAtlanIntegrationResource,
		resources.NewGithubIntegrationResource,
		resources.NewSecretResource,
		resources.NewDeploymentSettingsResource,
		resources.NewServiceUserResource,
		resources.NewServiceTokenResource,
		resources.NewOrganizationSettingsResource,
		resources.NewCustomMetricResource,
		resources.NewExternalAssetConnectionResource,
		resources.NewTeamMembershipResource,
		resources.NewTeamOrganizationGrantResource,
		resources.NewTeamDeploymentGrantResource,
		resources.NewTeamAllBranchDeploymentsGrantResource,
		resources.NewTeamBranchDeploymentsGrantResource,
		resources.NewUserOrganizationGrantResource,
		resources.NewUserDeploymentGrantResource,
		resources.NewUserAllBranchDeploymentsGrantResource,
		resources.NewUserBranchDeploymentsGrantResource,
		resources.NewServiceUserOrganizationGrantResource,
		resources.NewServiceUserDeploymentGrantResource,
		resources.NewServiceUserAllBranchDeploymentsGrantResource,
		resources.NewServiceUserBranchDeploymentsGrantResource,
		resources.NewCodeLocationFromDocumentResource,
	}
}

func (p *dagsterplusProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewDeploymentDataSource,
		datasources.NewDeploymentsDataSource,
		datasources.NewAlertPolicyDataSource,
		datasources.NewAlertPoliciesDataSource,
		datasources.NewUserDataSource,
		datasources.NewUsersDataSource,
		datasources.NewTeamDataSource,
		datasources.NewTeamsDataSource,
		datasources.NewCodeLocationDataSource,
		datasources.NewCodeLocationsDataSource,
		datasources.NewAgentTokenDataSource,
		datasources.NewUserTokenDataSource,
		datasources.NewRoleDataSource,
		datasources.NewRolesDataSource,
		datasources.NewOrganizationDataSource,
		datasources.NewSecretDataSource,
		datasources.NewServiceUserDataSource,
		datasources.NewCustomMetricDataSource,
		datasources.NewVersionDataSource,
		datasources.NewConfigurationDocumentDataSource,
	}
}

func (p *dagsterplusProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}
