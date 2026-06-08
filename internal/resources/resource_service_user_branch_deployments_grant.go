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

var _ resource.Resource = &serviceUserBranchDeploymentsGrantResource{}
var _ resource.ResourceWithImportState = &serviceUserBranchDeploymentsGrantResource{}

func NewServiceUserBranchDeploymentsGrantResource() resource.Resource {
	return &serviceUserBranchDeploymentsGrantResource{}
}

type serviceUserBranchDeploymentsGrantResource struct {
	client *client.Client
}

type serviceUserBranchDeploymentsGrantResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ServiceUserID    types.String `tfsdk:"service_user_id"`
	ParentDeployment types.String `tfsdk:"parent_deployment"`
	Grant            types.String `tfsdk:"grant"`
	CustomRoleID     types.String `tfsdk:"custom_role_id"`
}

func (r *serviceUserBranchDeploymentsGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user_branch_deployments_grant"
}

func (r *serviceUserBranchDeploymentsGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission grant for a Dagster+ service user across all branch deployments " +
			"of a specific parent (full) deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form `{service_user_id}/{parent_deployment}`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_user_id": schema.StringAttribute{
				Description: "The ID of the service user to grant access.",
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
				Description: withEnumValues(grantLevelDescription, grantLevels),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(grantLevels...),
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

func (r *serviceUserBranchDeploymentsGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceUserBranchDeploymentsGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceUserBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
		ServiceUserID:   plan.ServiceUserID.ValueString(),
		DeploymentScope: "branch_deployments",
		DeploymentID:    intID,
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting branch deployments grant", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServiceUserID.ValueString() + "/" + plan.ParentDeployment.ValueString())
	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceUserBranchDeploymentsGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceUserBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grant, err := r.client.GetServiceUserGrant(ctx, state.ServiceUserID.ValueString(), "branch_deployments", intID)
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

func (r *serviceUserBranchDeploymentsGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceUserBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
		ServiceUserID:   plan.ServiceUserID.ValueString(),
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

func (r *serviceUserBranchDeploymentsGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceUserBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.ParentDeployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	if err := r.client.DeleteServiceUserGrant(ctx, state.ServiceUserID.ValueString(), "branch_deployments", intID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing branch deployments grant", err.Error())
	}
}

func (r *serviceUserBranchDeploymentsGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{service_user_id}/{parent_deployment}'.",
		)
		return
	}

	serviceUserID := parts[0]
	parentDeployment := parts[1]

	intID, err := r.client.GetDeploymentIntID(ctx, parentDeployment)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
		return
	}

	grant, err := r.client.GetServiceUserGrant(ctx, serviceUserID, "branch_deployments", intID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading branch deployments grant", err.Error())
		return
	}

	grantVal, customRoleIDVal := grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state := serviceUserBranchDeploymentsGrantResourceModel{
		ID:               types.StringValue(req.ID),
		ServiceUserID:    types.StringValue(serviceUserID),
		ParentDeployment: types.StringValue(parentDeployment),
		Grant:            grantVal,
		CustomRoleID:     customRoleIDVal,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
