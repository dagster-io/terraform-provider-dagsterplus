package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &serviceTokenResource{}
var _ resource.ResourceWithImportState = &serviceTokenResource{}

func NewServiceTokenResource() resource.Resource {
	return &serviceTokenResource{}
}

type serviceTokenResource struct {
	client *client.Client
}

type serviceTokenResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ServiceUserID types.String `tfsdk:"service_user_id"`
	Description   types.String `tfsdk:"description"`
	Token         types.String `tfsdk:"token"`
}

func (r *serviceTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_token"
}

func (r *serviceTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ service token. " +
			"The token value is only available at creation time; it cannot be recovered after import. " +
			"Changing service_user_id forces a new resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The token ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_user_id": schema.StringAttribute{
				Description: "The ID of the service user this token belongs to. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "An optional description for the token.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The service token value. Only populated on creation; not recoverable after import.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *serviceTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.CreateServiceToken(ctx, plan.ServiceUserID.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating service token", err.Error())
		return
	}

	plan.ID = types.StringValue(tok.ID)
	plan.Description = types.StringValue(tok.Description)
	plan.Token = types.StringValue(tok.Token)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// service_user_id is not stored after import (the API cannot recover it).
	// Skip the lookup so the resource remains in state for the user to manually fix.
	if state.ServiceUserID.ValueString() == "" {
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	tok, err := r.client.GetServiceToken(ctx, state.ServiceUserID.ValueString(), state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading service token", err.Error())
		return
	}

	state.ID = types.StringValue(tok.ID)
	state.Description = types.StringValue(tok.Description)
	// state.Token is intentionally not overwritten — the API never returns it after creation.
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serviceTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.UpdateServiceToken(ctx, state.ID.ValueString(), plan.ServiceUserID.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating service token", err.Error())
		return
	}

	plan.ID = types.StringValue(tok.ID)
	plan.Description = types.StringValue(tok.Description)
	// Preserve token from state.
	plan.Token = state.Token
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RevokeServiceToken(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error revoking service token", err.Error())
	}
}

func (r *serviceTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Token value is irrecoverable after creation.
	// Service user ID must be supplied by the user after import.
	state := serviceTokenResourceModel{
		ID:            types.StringValue(req.ID),
		ServiceUserID: types.StringValue(""),
		Description:   types.StringValue(""),
		Token:         types.StringValue(""),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
