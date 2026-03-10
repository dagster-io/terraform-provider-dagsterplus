package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &scimSettingsResource{}
var _ resource.ResourceWithImportState = &scimSettingsResource{}

func NewScimSettingsResource() resource.Resource {
	return &scimSettingsResource{}
}

type scimSettingsResource struct {
	client *client.Client
}

type scimSettingsModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (r *scimSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scim_settings"
}

func (r *scimSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages SCIM sync settings for the Dagster+ organization. " +
			"This is a singleton resource — one per organization. " +
			"Destroying this resource sets SCIM sync to disabled.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Organization name (used as the resource identifier).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether SCIM sync is enabled for the organization.",
				Required:    true,
			},
		},
	}
}

func (r *scimSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *scimSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan scimSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetScimSyncEnabled(ctx, plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Error setting SCIM sync", err.Error())
		return
	}

	plan.ID = types.StringValue(r.client.Organization)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *scimSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state scimSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	enabled, err := r.client.GetScimSyncEnabled(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SCIM sync settings", err.Error())
		return
	}

	state.Enabled = types.BoolValue(enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *scimSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan scimSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetScimSyncEnabled(ctx, plan.Enabled.ValueBool()); err != nil {
		resp.Diagnostics.AddError("Error updating SCIM sync", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *scimSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.SetScimSyncEnabled(ctx, false); err != nil {
		resp.Diagnostics.AddError("Error disabling SCIM sync", err.Error())
	}
}

func (r *scimSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
