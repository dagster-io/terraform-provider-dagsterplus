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

var _ resource.Resource = &serviceUserAllBranchDeploymentsGrantResource{}
var _ resource.ResourceWithImportState = &serviceUserAllBranchDeploymentsGrantResource{}

func NewServiceUserAllBranchDeploymentsGrantResource() resource.Resource {
	return &serviceUserAllBranchDeploymentsGrantResource{}
}

type serviceUserAllBranchDeploymentsGrantResource struct {
	client *client.Client
}

type serviceUserAllBranchDeploymentsGrantResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ServiceUserID types.String `tfsdk:"service_user_id"`
	Grant         types.String `tfsdk:"grant"`
	CustomRoleID  types.String `tfsdk:"custom_role_id"`
}

func (r *serviceUserAllBranchDeploymentsGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user_all_branch_deployments_grant"
}

func (r *serviceUserAllBranchDeploymentsGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a permission grant for a Dagster+ service user across all branch deployments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier (the service_user_id).",
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

func (r *serviceUserAllBranchDeploymentsGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serviceUserAllBranchDeploymentsGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceUserAllBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
		ServiceUserID:   plan.ServiceUserID.ValueString(),
		DeploymentScope: "all_branch_deployments",
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting all branch deployments grant", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ServiceUserID.ValueString())
	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceUserAllBranchDeploymentsGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceUserAllBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant, err := r.client.GetServiceUserGrant(ctx, state.ServiceUserID.ValueString(), "all_branch_deployments", 0)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading all branch deployments grant", err.Error())
		return
	}

	state.Grant, state.CustomRoleID = grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serviceUserAllBranchDeploymentsGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceUserAllBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grantStr, customRoleID := resolveGrantFields(plan.Grant, plan.CustomRoleID)
	result, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
		ServiceUserID:   plan.ServiceUserID.ValueString(),
		DeploymentScope: "all_branch_deployments",
		Grant:           grantStr,
		CustomRoleID:    customRoleID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating all branch deployments grant", err.Error())
		return
	}

	plan.Grant, plan.CustomRoleID = grantFieldsFromAPI(result.Grant, result.CustomRoleID)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceUserAllBranchDeploymentsGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceUserAllBranchDeploymentsGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServiceUserGrant(ctx, state.ServiceUserID.ValueString(), "all_branch_deployments", 0); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing all branch deployments grant", err.Error())
	}
}

func (r *serviceUserAllBranchDeploymentsGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	serviceUserID := req.ID

	grant, err := r.client.GetServiceUserGrant(ctx, serviceUserID, "all_branch_deployments", 0)
	if err != nil {
		resp.Diagnostics.AddError("Error reading all branch deployments grant", err.Error())
		return
	}

	grantVal, customRoleIDVal := grantFieldsFromAPI(grant.Grant, grant.CustomRoleID)
	state := serviceUserAllBranchDeploymentsGrantResourceModel{
		ID:            types.StringValue(serviceUserID),
		ServiceUserID: types.StringValue(serviceUserID),
		Grant:         grantVal,
		CustomRoleID:  customRoleIDVal,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
