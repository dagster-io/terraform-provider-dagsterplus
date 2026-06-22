package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &teamDeploymentGrantResource{}
var _ resource.ResourceWithImportState = &teamDeploymentGrantResource{}

func NewTeamDeploymentGrantResource() resource.Resource {
	return &teamDeploymentGrantResource{}
}

type teamDeploymentGrantResource struct {
	client *client.Client
}

type locationGrantModel struct {
	LocationName types.String `tfsdk:"location_name"`
	Grant        types.String `tfsdk:"grant"`
	CustomRoleID types.String `tfsdk:"custom_role_id"`
}

type teamDeploymentGrantResourceModel struct {
	ID             types.String         `tfsdk:"id"`
	TeamID         types.String         `tfsdk:"team_id"`
	Deployment     types.String         `tfsdk:"deployment"`
	Grant          types.String         `tfsdk:"grant"`
	CustomRoleID   types.String         `tfsdk:"custom_role_id"`
	LocationGrants []locationGrantModel `tfsdk:"location_grants"`
}

func (r *teamDeploymentGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_deployment_grant"
}

func (r *teamDeploymentGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a deployment-level permission grant for a Dagster+ team. " +
			"Use this resource when the team is managed separately from its deployment permissions. " +
			"If you manage the team with dagsterplus_team, prefer the inline deployment_grant block instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form `{team_id}/{deployment}`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team_id": schema.StringAttribute{
				Description: "The ID of the team to grant access.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment to grant access to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"grant": schema.StringAttribute{
				Description: withEnumValues(grantLevelDescription, grantLevels),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(grantLevels...),
					stringvalidator.ExactlyOneOf(path.MatchRoot("custom_role_id")),
				},
			},
			"custom_role_id": schema.StringAttribute{
				Description: "The ID of a custom role to assign. Exactly one of grant or custom_role_id must be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("grant")),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"location_grants": schema.ListNestedBlock{
				Description: "Per-code-location permission overrides within this deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"location_name": schema.StringAttribute{
							Description: "The name of the code location.",
							Required:    true,
						},
						"grant": schema.StringAttribute{
							Description: withEnumValues(grantLevelDescription, grantLevels),
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(grantLevels...),
								stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("custom_role_id")),
							},
						},
						"custom_role_id": schema.StringAttribute{
							Description: "Custom role ID for this location. Exactly one of grant or custom_role_id must be set.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("grant")),
							},
						},
					},
				},
			},
		},
	}
}

func (r *teamDeploymentGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func locationGrantsToClient(models []locationGrantModel) []client.LocationGrant {
	grants := make([]client.LocationGrant, len(models))
	for i, m := range models {
		g, roleID := resolveGrantFields(m.Grant, m.CustomRoleID)
		grants[i] = client.LocationGrant{
			LocationName: m.LocationName.ValueString(),
			Grant:        g,
			CustomRoleID: roleID,
		}
	}
	return grants
}

func locationGrantsFromClient(grants []client.LocationGrant) []locationGrantModel {
	models := make([]locationGrantModel, len(grants))
	for i, g := range grants {
		grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
		models[i] = locationGrantModel{
			LocationName: types.StringValue(g.LocationName),
			Grant:        grantVal,
			CustomRoleID: customRoleIDVal,
		}
	}
	return models
}

func (r *teamDeploymentGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
		TeamID:          plan.TeamID.ValueString(),
		DeploymentScope: "deployment",
		DeploymentID:    intID,
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
		LocationGrants:  locationGrantsToClient(plan.LocationGrants),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting deployment grant", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.TeamID.ValueString() + "/" + plan.Deployment.ValueString())
	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	// The mutation response does not return locationGrants; preserve the plan values
	// since the API accepted the mutation without error.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamDeploymentGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grant, err := r.client.GetTeamGrant(ctx, state.TeamID.ValueString(), "deployment", intID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading deployment grant", err.Error())
		return
	}

	state.Grant, state.CustomRoleID = grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state.LocationGrants = locationGrantsFromClient(grant.LocationGrants)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *teamDeploymentGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
		TeamID:          plan.TeamID.ValueString(),
		DeploymentScope: "deployment",
		DeploymentID:    intID,
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
		LocationGrants:  locationGrantsToClient(plan.LocationGrants),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating deployment grant", err.Error())
		return
	}

	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	// The mutation response does not return locationGrants; preserve the plan values
	// since the API accepted the mutation without error.
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamDeploymentGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	if err := r.client.DeleteTeamGrant(ctx, state.TeamID.ValueString(), "deployment", intID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing deployment grant", err.Error())
	}
}

func (r *teamDeploymentGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{team_id}/{deployment}'.",
		)
		return
	}

	teamID := parts[0]
	deployment := parts[1]

	intID, err := r.client.GetDeploymentIntID(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grant, err := r.client.GetTeamGrant(ctx, teamID, "deployment", intID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment grant", err.Error())
		return
	}

	grantVal, customRoleIDVal := grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state := teamDeploymentGrantResourceModel{
		ID:             types.StringValue(req.ID),
		TeamID:         types.StringValue(teamID),
		Deployment:     types.StringValue(deployment),
		Grant:          grantVal,
		CustomRoleID:   customRoleIDVal,
		LocationGrants: locationGrantsFromClient(grant.LocationGrants),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
