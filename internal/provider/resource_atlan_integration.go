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

var _ resource.Resource = &atlanIntegrationResource{}
var _ resource.ResourceWithImportState = &atlanIntegrationResource{}

func NewAtlanIntegrationResource() resource.Resource {
	return &atlanIntegrationResource{}
}

type atlanIntegrationResource struct {
	client *client.Client
}

type atlanIntegrationModel struct {
	ID     types.String `tfsdk:"id"`
	Token  types.String `tfsdk:"token"`
	Domain types.String `tfsdk:"domain"`
}

func (r *atlanIntegrationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_atlan_integration"
}

func (r *atlanIntegrationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the Atlan integration for the Dagster+ organization. " +
			"This is a singleton resource — one per organization. " +
			"Destroying this resource removes the Atlan integration settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Organization name (used as the resource identifier).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The Atlan API token.",
				Required:    true,
				Sensitive:   true,
			},
			"domain": schema.StringAttribute{
				Description: "The Atlan domain (e.g. my-org.atlan.com).",
				Required:    true,
			},
		},
	}
}

func (r *atlanIntegrationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *atlanIntegrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan atlanIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetAtlanIntegration(ctx, plan.Token.ValueString(), plan.Domain.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error creating Atlan integration", err.Error())
		return
	}

	plan.ID = types.StringValue(r.client.Organization)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *atlanIntegrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state atlanIntegrationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	atlan, err := r.client.GetAtlanIntegration(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Atlan integration", err.Error())
		return
	}
	if atlan == nil {
		// Integration was deleted outside Terraform.
		resp.State.RemoveResource(ctx)
		return
	}

	state.Token = types.StringValue(atlan.Token)
	state.Domain = types.StringValue(atlan.Domain)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *atlanIntegrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan atlanIntegrationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.SetAtlanIntegration(ctx, plan.Token.ValueString(), plan.Domain.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error updating Atlan integration", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *atlanIntegrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if err := r.client.DeleteAtlanIntegration(ctx); err != nil {
		resp.Diagnostics.AddError("Error deleting Atlan integration", err.Error())
	}
}

func (r *atlanIntegrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
