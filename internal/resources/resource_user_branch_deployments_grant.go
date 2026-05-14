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

var _ resource.Resource = &userBranchDeploymentsGrantResource{}
var _ resource.ResourceWithImportState = &userBranchDeploymentsGrantResource{}

func NewUserBranchDeploymentsGrantResource() resource.Resource {
	return &userBranchDeploymentsGrantResource{}
}

type userBranchDeploymentsGrantResource struct {
	client *client.Client
}

type userBranchDeploymentsGrantResourceModel struct {
	ID               types.String `tfsdk:"id"`
	UserID           types.String `tfsdk:"user_id"`
	ParentDeployment types.String `tfsdk:"parent_deployment"`
	Grant            types.String `tfsdk:"grant"`
	CustomRoleID     types.String `tfsdk:"custom_role_id"`
}

func (r *userBranchDeploymentsGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_branch_deployments_grant"
}

func (r *userBranchDeploymentsGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission grant for a Dagster+ user across all branch deployments " +
			"of a specific parent (full) deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form `{user_id}/{parent_deployment}`.",
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
			"parent_deployment": schema.StringAttribute{
				Description: "The name of the full (parent) deployment whose branch deployments this grant applies to.",
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
	}
}

func (r *userBranchDeploymentsGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *userBranchDeploymentsGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetUserGrant(ctx, client.UserGrant{
		UserID:          user.ID,
		Email:           user.Email,
		DeploymentScope: "branch_deployments",
		DeploymentID:    intID,
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting branch deployments grant", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.UserID.ValueString() + "/" + plan.ParentDeployment.ValueString())
	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userBranchDeploymentsGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grant, err := r.client.GetUserGrant(ctx, state.UserID.ValueString(), "branch_deployments", intID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading branch deployments grant", err.Error())
		return
	}

	state.Grant, state.CustomRoleID = grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *userBranchDeploymentsGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, plan.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving user", err.Error())
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetUserGrant(ctx, client.UserGrant{
		UserID:          user.ID,
		Email:           user.Email,
		DeploymentScope: "branch_deployments",
		DeploymentID:    intID,
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch deployments grant", err.Error())
		return
	}

	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userBranchDeploymentsGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userBranchDeploymentsGrantResourceModel
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

	intID, err := r.client.GetDeploymentIntID(ctx, state.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	if err := r.client.DeleteUserGrant(ctx, user.Email, "branch_deployments", intID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing branch deployments grant", err.Error())
	}
}

func (r *userBranchDeploymentsGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{user_id}/{parent_deployment}'.",
		)
		return
	}

	userID := parts[0]
	parentDeployment := parts[1]

	intID, err := r.client.GetDeploymentIntID(ctx, parentDeployment)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grant, err := r.client.GetUserGrant(ctx, userID, "branch_deployments", intID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading branch deployments grant", err.Error())
		return
	}

	grantVal, customRoleIDVal := grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state := userBranchDeploymentsGrantResourceModel{
		ID:               types.StringValue(req.ID),
		UserID:           types.StringValue(userID),
		ParentDeployment: types.StringValue(parentDeployment),
		Grant:            grantVal,
		CustomRoleID:     customRoleIDVal,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
