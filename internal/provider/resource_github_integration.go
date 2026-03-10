package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &githubIntegrationResource{}
var _ resource.ResourceWithImportState = &githubIntegrationResource{}

func NewGithubIntegrationResource() resource.Resource {
	return &githubIntegrationResource{}
}

type githubIntegrationResource struct {
	client *client.Client
}

type githubIntegrationModel struct {
	ID          types.String `tfsdk:"id"`
	AccountName types.String `tfsdk:"account_name"`
	AppID       types.String `tfsdk:"app_id"`
	SettingsURL types.String `tfsdk:"settings_url"`
	Repos       types.List   `tfsdk:"repos"`
}

func (r *githubIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_github_integration"
}

func (r *githubIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Selects the active GitHub App installation for the Dagster+ organization. " +
			"This is a singleton resource — one per organization. " +
			"The GitHub App must already be installed via the GitHub OAuth flow before using this resource. " +
			"Destroying this resource deselects (deactivates) the GitHub App installation.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Organization name (used as the resource identifier).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_name": schema.StringAttribute{
				Description: "The GitHub account or organization name that has the app installed.",
				Required:    true,
			},
			"app_id": schema.StringAttribute{
				Description: "The GitHub App ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"settings_url": schema.StringAttribute{
				Description: "URL to the GitHub App installation settings.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"repos": schema.ListAttribute{
				Description: "List of repositories accessible via the GitHub App installation.",
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *githubIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *githubIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan githubIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gh, err := r.client.SelectGithubInstallation(ctx, plan.AccountName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error selecting GitHub installation", err.Error())
		return
	}

	state, diags := githubInstallationToModel(r.client.Organization, gh)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *githubIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state githubIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading GitHub integration", err.Error())
		return
	}
	if org.GitHub == nil {
		// Installation was removed outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := githubInstallationToModel(r.client.Organization, org.GitHub)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Preserve the id from existing state.
	newState.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *githubIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan githubIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gh, err := r.client.SelectGithubInstallation(ctx, plan.AccountName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating GitHub installation", err.Error())
		return
	}

	state, diags := githubInstallationToModel(r.client.Organization, gh)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *githubIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeselectGithubInstallation(ctx); err != nil {
		resp.Diagnostics.AddError("Error deselecting GitHub installation", err.Error())
	}
}

func (r *githubIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// githubInstallationToModel converts a client.GitHubInstallation to the Terraform state model.
func githubInstallationToModel(orgName string, gh *client.GitHubInstallation) (githubIntegrationModel, diag.Diagnostics) {
	repos, diags := types.ListValueFrom(context.Background(), types.StringType, gh.Repos)
	return githubIntegrationModel{
		ID:          types.StringValue(orgName),
		AccountName: types.StringValue(gh.AccountName),
		AppID:       types.StringValue(gh.AppID),
		SettingsURL: types.StringValue(gh.SettingsURL),
		Repos:       repos,
	}, diags
}
