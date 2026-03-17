package datasources

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &organizationDataSource{}

func NewOrganizationDataSource() datasource.DataSource {
	return &organizationDataSource{}
}

type organizationDataSource struct {
	client *client.Client
}

type organizationDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	PublicID      types.String `tfsdk:"public_id"`
	Name          types.String `tfsdk:"name"`
	Status        types.String `tfsdk:"status"`
	AccountReview types.String `tfsdk:"account_review"`
	// GitHub App installation (nil when not installed)
	GitHubAccountName types.String `tfsdk:"github_account_name"`
	GitHubAppID       types.String `tfsdk:"github_app_id"`
	GitHubSettingsURL types.String `tfsdk:"github_settings_url"`
	GitHubRepos       types.List   `tfsdk:"github_repos"`
	// Slack App installation (nil when not installed)
	SlackTeamID   types.String `tfsdk:"slack_team_id"`
	SlackTeamName types.String `tfsdk:"slack_team_name"`
}

func (d *organizationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *organizationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads metadata about the Dagster+ organization configured in the provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Organization name (used as the data source identifier).",
				Computed:    true,
			},
			"public_id": schema.StringAttribute{
				Description: "The organization's public UUID.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The organization name.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The organization status: ACTIVE, READ_ONLY, SUSPENDED, or PENDING_DELETION.",
				Computed:    true,
			},
			"account_review": schema.StringAttribute{
				Description: "Account review status of the organization.",
				Computed:    true,
			},
			"github_account_name": schema.StringAttribute{
				Description: "GitHub account or organization name of the active GitHub App installation. Empty when not installed.",
				Computed:    true,
			},
			"github_app_id": schema.StringAttribute{
				Description: "GitHub App ID of the active installation. Empty when not installed.",
				Computed:    true,
			},
			"github_settings_url": schema.StringAttribute{
				Description: "URL to GitHub App installation settings. Empty when not installed.",
				Computed:    true,
				Sensitive:   true,
			},
			"github_repos": schema.ListAttribute{
				Description: "Repositories accessible via the GitHub App installation.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"slack_team_id": schema.StringAttribute{
				Description: "Slack workspace team ID of the active Slack App installation. Empty when not installed.",
				Computed:    true,
			},
			"slack_team_name": schema.StringAttribute{
				Description: "Slack workspace name of the active Slack App installation. Empty when not installed.",
				Computed:    true,
			},
		},
	}
}

func (d *organizationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *organizationDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	org, err := d.client.GetOrganization(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization", err.Error())
		return
	}

	state := organizationDataSourceModel{
		ID:            types.StringValue(org.Name),
		PublicID:      types.StringValue(org.PublicID),
		Name:          types.StringValue(org.Name),
		Status:        types.StringValue(org.Status),
		AccountReview: types.StringValue(org.AccountReview),
	}

	// GitHub App installation fields.
	if org.GitHub != nil {
		state.GitHubAccountName = types.StringValue(org.GitHub.AccountName)
		state.GitHubAppID = types.StringValue(org.GitHub.AppID)
		state.GitHubSettingsURL = types.StringValue(org.GitHub.SettingsURL)
		repos, diags := types.ListValueFrom(ctx, types.StringType, org.GitHub.Repos)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.GitHubRepos = repos
	} else {
		state.GitHubAccountName = types.StringValue("")
		state.GitHubAppID = types.StringValue("")
		state.GitHubSettingsURL = types.StringValue("")
		state.GitHubRepos, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	// Slack App installation fields.
	if org.Slack != nil {
		state.SlackTeamID = types.StringValue(org.Slack.TeamID)
		state.SlackTeamName = types.StringValue(org.Slack.TeamName)
	} else {
		state.SlackTeamID = types.StringValue("")
		state.SlackTeamName = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
