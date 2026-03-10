package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &deploymentSettingsResource{}
var _ resource.ResourceWithImportState = &deploymentSettingsResource{}

func NewDeploymentSettingsResource() resource.Resource {
	return &deploymentSettingsResource{}
}

type deploymentSettingsResource struct {
	client *client.Client
}

type deploymentSettingsResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Deployment   types.String `tfsdk:"deployment"`
	SettingsJSON types.String `tfsdk:"settings_json"`
}

func (r *deploymentSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment_settings"
}

func (r *deploymentSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Dagster+ deployment-level settings as a JSON blob. " +
			"On destroy, the settings are reset to an empty configuration ({}).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The deployment name (used as the resource identifier).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment to configure. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"settings_json": schema.StringAttribute{
				Description: "JSON string of deployment settings configuration.",
				Required:    true,
			},
		},
	}
}

func (r *deploymentSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deploymentSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetDeploymentSettings(ctx, plan.Deployment.ValueString(), plan.SettingsJSON.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error setting deployment settings", err.Error())
		return
	}

	plan.ID = plan.Deployment
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *deploymentSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settingsJSON, err := r.client.GetDeploymentSettings(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment settings", err.Error())
		return
	}

	state.SettingsJSON = types.StringValue(settingsJSON)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *deploymentSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deploymentSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetDeploymentSettings(ctx, plan.Deployment.ValueString(), plan.SettingsJSON.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating deployment settings", err.Error())
		return
	}

	plan.ID = plan.Deployment
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *deploymentSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reset to empty settings on destroy.
	if err := r.client.SetDeploymentSettings(ctx, state.Deployment.ValueString(), "{}"); err != nil {
		resp.Diagnostics.AddError("Error resetting deployment settings", err.Error())
	}
}

func (r *deploymentSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	deployment := req.ID
	settingsJSON, err := r.client.GetDeploymentSettings(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error importing deployment settings", err.Error())
		return
	}

	state := deploymentSettingsResourceModel{
		ID:           types.StringValue(deployment),
		Deployment:   types.StringValue(deployment),
		SettingsJSON: types.StringValue(settingsJSON),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
