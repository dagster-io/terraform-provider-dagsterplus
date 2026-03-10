package provider

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

var _ resource.Resource = &teamMembershipResource{}
var _ resource.ResourceWithImportState = &teamMembershipResource{}

func NewTeamMembershipResource() resource.Resource {
	return &teamMembershipResource{}
}

type teamMembershipResource struct {
	client *client.Client
}

type teamMembershipResourceModel struct {
	ID     types.String `tfsdk:"id"`
	TeamID types.String `tfsdk:"team_id"`
	UserID types.String `tfsdk:"user_id"`
}

func (r *teamMembershipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_membership"
}

func (r *teamMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single user's membership in a Dagster+ team. " +
			"Use this resource when the team itself is managed separately. " +
			"If you manage the team with dagsterplus_team, prefer the inline member block instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form '{team_id}/{user_id}'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				Description: "The ID of the team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to add to the team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *teamMembershipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamMembershipResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AddUserToTeam(ctx, plan.TeamID.ValueString(), plan.UserID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error adding user to team", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.TeamID.ValueString() + "/" + plan.UserID.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inTeam, err := r.client.IsUserInTeam(ctx, state.TeamID.ValueString(), state.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error checking team membership", err.Error())
		return
	}

	if !inTeam {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *teamMembershipResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All fields require replacement; Update is never called.
}

func (r *teamMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamMembershipResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RemoveUserFromTeam(ctx, state.TeamID.ValueString(), state.UserID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error removing user from team", err.Error())
	}
}

func (r *teamMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{team_id}/{user_id}'.",
		)
		return
	}

	state := teamMembershipResourceModel{
		ID:     types.StringValue(req.ID),
		TeamID: types.StringValue(parts[0]),
		UserID: types.StringValue(parts[1]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
