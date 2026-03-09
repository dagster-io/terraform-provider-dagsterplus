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

// Ensure teamResource satisfies the resource.Resource interface.
var _ resource.Resource = &teamResource{}
var _ resource.ResourceWithImportState = &teamResource{}

// NewTeamResource returns a new team resource.
func NewTeamResource() resource.Resource {
	return &teamResource{}
}

// teamResource defines the resource implementation.
type teamResource struct {
	client *client.Client
}

// deploymentPermissionModel represents a single deployment_permission block.
type deploymentPermissionModel struct {
	DeploymentName types.String `tfsdk:"deployment_name"`
	Role           types.String `tfsdk:"role"`
}

// teamResourceModel describes the resource data model.
type teamResourceModel struct {
	ID                    types.String                `tfsdk:"id"`
	Name                  types.String                `tfsdk:"name"`
	DeploymentPermissions []deploymentPermissionModel `tfsdk:"deployment_permission"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ team and its deployment permissions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The team ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"deployment_permission": schema.ListNestedBlock{
				Description: "Grants a role to this team on a specific deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"deployment_name": schema.StringAttribute{
							Description: "The name of the deployment.",
							Required:    true,
						},
						"role": schema.StringAttribute{
							Description: "The role to grant: VIEWER, LAUNCHER, EDITOR, or ADMIN.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func permModelsToClient(perms []deploymentPermissionModel) []client.Permission {
	out := make([]client.Permission, len(perms))
	for i, p := range perms {
		out[i] = client.Permission{
			DeploymentName: p.DeploymentName.ValueString(),
			Role:           p.Role.ValueString(),
		}
	}
	return out
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.CreateTeam(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating team", err.Error())
		return
	}

	plan.ID = types.StringValue(team.ID)

	if len(plan.DeploymentPermissions) > 0 {
		if err := r.client.UpdateTeamPermissions(ctx, team.ID, permModelsToClient(plan.DeploymentPermissions)); err != nil {
			resp.Diagnostics.AddError("Error setting team permissions", err.Error())
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading team", err.Error())
		return
	}

	state.ID = types.StringValue(team.ID)
	state.Name = types.StringValue(team.Name)

	perms, err := r.client.GetTeamPermissions(ctx, team.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading team permissions", err.Error())
		return
	}

	state.DeploymentPermissions = make([]deploymentPermissionModel, len(perms))
	for i, p := range perms {
		state.DeploymentPermissions[i] = deploymentPermissionModel{
			DeploymentName: types.StringValue(p.DeploymentName),
			Role:           types.StringValue(p.Role),
		}
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateTeamPermissions(ctx, plan.ID.ValueString(), permModelsToClient(plan.DeploymentPermissions)); err != nil {
		resp.Diagnostics.AddError("Error updating team permissions", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTeam(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting team", err.Error())
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by team ID.
	team, err := r.client.GetTeam(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing team", err.Error())
		return
	}

	perms, err := r.client.GetTeamPermissions(ctx, team.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading team permissions during import", err.Error())
		return
	}

	state := teamResourceModel{
		ID:   types.StringValue(team.ID),
		Name: types.StringValue(team.Name),
	}

	state.DeploymentPermissions = make([]deploymentPermissionModel, len(perms))
	for i, p := range perms {
		state.DeploymentPermissions[i] = deploymentPermissionModel{
			DeploymentName: types.StringValue(p.DeploymentName),
			Role:           types.StringValue(p.Role),
		}
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
