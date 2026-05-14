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

var _ resource.Resource = &userDeploymentGrantResource{}
var _ resource.ResourceWithImportState = &userDeploymentGrantResource{}

func NewUserDeploymentGrantResource() resource.Resource {
	return &userDeploymentGrantResource{}
}

type userDeploymentGrantResource struct {
	client *client.Client
}

type userDeploymentGrantResourceModel struct {
	ID             types.String         `tfsdk:"id"`
	UserID         types.String         `tfsdk:"user_id"`
	Deployment     types.String         `tfsdk:"deployment"`
	Grant          types.String         `tfsdk:"grant"`
	CustomRoleID   types.String         `tfsdk:"custom_role_id"`
	LocationGrants []locationGrantModel `tfsdk:"location_grants"`
}

func (r *userDeploymentGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_deployment_grant"
}

func (r *userDeploymentGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a deployment-level permission grant for a Dagster+ user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form `{user_id}/{deployment}`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to grant access.",
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
				Description: "Standard permission level: `VIEWER`, `LAUNCHER`, `EDITOR`, or `ADMIN`. Conflicts with custom_role_id.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("VIEWER", "LAUNCHER", "EDITOR", "ADMIN"),
					stringvalidator.ConflictsWith(path.MatchRoot("custom_role_id")),
				},
			},
			"custom_role_id": schema.StringAttribute{
				Description: "The ID of a custom role to assign. Conflicts with grant.",
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
							Description: "Standard permission level for this location. Conflicts with custom_role_id.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("VIEWER", "LAUNCHER", "EDITOR", "ADMIN"),
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("custom_role_id")),
							},
						},
						"custom_role_id": schema.StringAttribute{
							Description: "Custom role ID for this location. Conflicts with grant.",
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

func (r *userDeploymentGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userDeploymentGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetUserGrant(ctx, client.UserGrant{
		UserID:          user.ID,
		Email:           user.Email,
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

	plan.ID = types.StringValue(plan.UserID.ValueString() + "/" + plan.Deployment.ValueString())
	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userDeploymentGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grant, err := r.client.GetUserGrant(ctx, state.UserID.ValueString(), "deployment", intID)
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

func (r *userDeploymentGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetUserGrant(ctx, client.UserGrant{
		UserID:          user.ID,
		Email:           user.Email,
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
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userDeploymentGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.UserID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	if err := r.client.DeleteUserGrant(ctx, user.Email, "deployment", intID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing deployment grant", err.Error())
	}
}

func (r *userDeploymentGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{user_id}/{deployment}'.",
		)
		return
	}

	userID := parts[0]
	deployment := parts[1]

	intID, err := r.client.GetDeploymentIntID(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	grant, err := r.client.GetUserGrant(ctx, userID, "deployment", intID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment grant", err.Error())
		return
	}

	grantVal, customRoleIDVal := grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state := userDeploymentGrantResourceModel{
		ID:             types.StringValue(req.ID),
		UserID:         types.StringValue(userID),
		Deployment:     types.StringValue(deployment),
		Grant:          grantVal,
		CustomRoleID:   customRoleIDVal,
		LocationGrants: locationGrantsFromClient(grant.LocationGrants),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
