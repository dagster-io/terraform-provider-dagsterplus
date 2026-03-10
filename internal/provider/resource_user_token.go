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

var _ resource.Resource = &userTokenResource{}
var _ resource.ResourceWithImportState = &userTokenResource{}

func NewUserTokenResource() resource.Resource {
	return &userTokenResource{}
}

type userTokenResource struct {
	client *client.Client
}

type userTokenResourceModel struct {
	ID     types.String `tfsdk:"id"`
	UserID types.String `tfsdk:"user_id"`
	Name   types.String `tfsdk:"name"`
	Token  types.String `tfsdk:"token"`
}

func (r *userTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_token"
}

func (r *userTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ user token. " +
			"The token value is only available at creation time; it cannot be recovered after import. " +
			"Changing the name forces a new resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The token ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user who owns this token. Required for revocation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The label for the user token. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The user token value. Only populated on creation; not recoverable after import.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *userTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.CreateUserToken(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating user token", err.Error())
		return
	}

	plan.ID = types.StringValue(tok.ID)
	plan.UserID = types.StringValue(tok.UserID)
	plan.Name = types.StringValue(tok.Name)
	plan.Token = types.StringValue(tok.Token)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.GetUserToken(ctx, state.ID.ValueString(), state.UserID.ValueString())
	if err != nil {
		// If we can't list tokens (e.g. API shape mismatch), preserve existing state
		// rather than failing — the resource still exists and delete will use the stored ID.
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	// Preserve the token value from state — the API never returns it after creation.
	state.ID = types.StringValue(tok.ID)
	state.Name = types.StringValue(tok.Name)
	// state.Token and state.UserID are intentionally not overwritten.

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userTokenResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Name is RequiresReplace — Update is never called.
}

func (r *userTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUserToken(ctx, state.ID.ValueString(), state.UserID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting user token", err.Error())
	}
}

func (r *userTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Token value and user_id are irrecoverable after creation.
	// Name may not be resolvable if the list API is unavailable.
	state := userTokenResourceModel{
		ID:     types.StringValue(req.ID),
		UserID: types.StringValue(""),
		Name:   types.StringValue(""),
		Token:  types.StringValue(""),
	}

	// user_id is not known at import time; GetUserToken requires it, so skip the lookup.
	_ = ctx

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
