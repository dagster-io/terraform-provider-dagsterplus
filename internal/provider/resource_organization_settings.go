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

var _ resource.Resource = &organizationSettingsResource{}
var _ resource.ResourceWithImportState = &organizationSettingsResource{}

func NewOrganizationSettingsResource() resource.Resource {
	return &organizationSettingsResource{}
}

type organizationSettingsResource struct {
	client *client.Client
}

type organizationSettingsResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SettingsJSON types.String `tfsdk:"settings_json"`
}

func (r *organizationSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_settings"
}

func (r *organizationSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Dagster+ organization-level settings as a JSON blob. " +
			"This is a singleton resource — only one instance should exist per organization. " +
			"On destroy, settings are reset to an empty configuration ({}).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Always set to the organization name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"settings_json": schema.StringAttribute{
				Description: "JSON string of organization settings configuration.",
				Required:    true,
			},
		},
	}
}

func (r *organizationSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetOrganizationSettings(ctx, plan.SettingsJSON.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error setting organization settings", err.Error())
		return
	}

	plan.ID = types.StringValue(r.client.Organization)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *organizationSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settingsJSON, err := r.client.GetOrganizationSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading organization settings", err.Error())
		return
	}

	state.SettingsJSON = types.StringValue(settingsJSON)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *organizationSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetOrganizationSettings(ctx, plan.SettingsJSON.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating organization settings", err.Error())
		return
	}

	plan.ID = types.StringValue(r.client.Organization)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *organizationSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Reset to empty settings on destroy.
	if err := r.client.SetOrganizationSettings(ctx, "{}"); err != nil {
		resp.Diagnostics.AddError("Error resetting organization settings", err.Error())
	}
}

func (r *organizationSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	settingsJSON, err := r.client.GetOrganizationSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error importing organization settings", err.Error())
		return
	}

	state := organizationSettingsResourceModel{
		ID:           types.StringValue(r.client.Organization),
		SettingsJSON: types.StringValue(settingsJSON),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
