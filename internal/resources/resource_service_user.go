package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &serviceUserResource{}
var _ resource.ResourceWithImportState = &serviceUserResource{}

func NewServiceUserResource() resource.Resource {
	return &serviceUserResource{}
}

type serviceUserResource struct {
	client *client.Client
}

type serviceUserResourceModel struct {
	ID                        types.String                  `tfsdk:"id"`
	Name                      types.String                  `tfsdk:"name"`
	Description               types.String                  `tfsdk:"description"`
	OrganizationGrant         []orgGrantModel               `tfsdk:"organization_grant"`
	AllBranchDeploymentsGrant []orgGrantModel               `tfsdk:"all_branch_deployments_grant"`
	DeploymentGrant           []deploymentGrantModel        `tfsdk:"deployment_grant"`
	BranchDeploymentsGrant    []branchDeploymentsGrantModel `tfsdk:"branch_deployments_grant"`
}

func (r *serviceUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (r *serviceUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	orgGrantBlock := schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"grant": schema.StringAttribute{
				Description: withEnumValues(orgGrantDescription, orgGrantLevels),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(orgGrantLevels...),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("custom_role_id")),
				},
			},
			"custom_role_id": schema.StringAttribute{
				Description: "The ID of a custom role to assign. Exactly one of grant or custom_role_id must be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("grant")),
				},
			},
		},
	}
	deploymentScopeGrantBlock := schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"grant": schema.StringAttribute{
				Description: withEnumValues(grantLevelDescription, grantLevels),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(grantLevels...),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("custom_role_id")),
				},
			},
			"custom_role_id": schema.StringAttribute{
				Description: "The ID of a custom role to assign. Exactly one of grant or custom_role_id must be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("grant")),
				},
			},
		},
	}
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ service user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The service user ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the service user.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "An optional description for the service user.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"organization_grant": schema.ListNestedBlock{
				Description:  "Organization-level permission grant for the service user. At most one block is allowed.",
				NestedObject: orgGrantBlock,
			},
			"all_branch_deployments_grant": schema.ListNestedBlock{
				Description:  "Permission grant for the service user across all branch deployments. At most one block is allowed.",
				NestedObject: deploymentScopeGrantBlock,
			},
			"deployment_grant": schema.ListNestedBlock{
				Description: "Deployment-level permission grant for the service user. One block per deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"deployment": schema.StringAttribute{
							Description: "The name of the deployment to grant access to.",
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
							Description: "The ID of a custom role to assign. Exactly one of grant or custom_role_id must be set.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("grant")),
							},
						},
					},
				},
			},
			"branch_deployments_grant": schema.ListNestedBlock{
				Description: "Permission grant across all branch deployments of a specific parent (full) deployment. One block per parent deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"parent_deployment": schema.StringAttribute{
							Description: "The name of the full (parent) deployment whose branch deployments this grant applies to.",
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
							Description: "The ID of a custom role to assign. Exactly one of grant or custom_role_id must be set.",
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

func (r *serviceUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func serviceUserGrantToOrgModel(g *client.ServiceUserGrant) orgGrantModel {
	grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
	return orgGrantModel{Grant: grantVal, CustomRoleID: customRoleIDVal}
}

func (r *serviceUserResource) applyGrants(ctx context.Context, suID string, plan *serviceUserResourceModel, diags *diag.Diagnostics) {
	if len(plan.OrganizationGrant) > 0 {
		g := plan.OrganizationGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
			ServiceUserID:   suID,
			DeploymentScope: "organization",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			diags.AddError("Error setting organization grant", err.Error())
			return
		}
		plan.OrganizationGrant = []orgGrantModel{serviceUserGrantToOrgModel(result)}
	}
	if len(plan.AllBranchDeploymentsGrant) > 0 {
		g := plan.AllBranchDeploymentsGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
			ServiceUserID:   suID,
			DeploymentScope: "all_branch_deployments",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			diags.AddError("Error setting all-branch-deployments grant", err.Error())
			return
		}
		plan.AllBranchDeploymentsGrant = []orgGrantModel{serviceUserGrantToOrgModel(result)}
	}
	for i, g := range plan.DeploymentGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.Deployment.ValueString())
		if err != nil {
			diags.AddError("Error resolving deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
			ServiceUserID:   suID,
			DeploymentScope: "deployment",
			DeploymentID:    intID,
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		}); err != nil {
			diags.AddError("Error setting deployment grant", err.Error())
			return
		}
		plan.DeploymentGrant[i] = g
	}
	for i, g := range plan.BranchDeploymentsGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.ParentDeployment.ValueString())
		if err != nil {
			diags.AddError("Error resolving parent deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetServiceUserGrant(ctx, client.ServiceUserGrant{
			ServiceUserID:   suID,
			DeploymentScope: "branch_deployments",
			DeploymentID:    intID,
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		}); err != nil {
			diags.AddError("Error setting branch deployments grant", err.Error())
			return
		}
		plan.BranchDeploymentsGrant[i] = g
	}
}

func (r *serviceUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.CreateServiceUser(ctx, plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating service user", err.Error())
		return
	}

	plan.ID = types.StringValue(u.ID)
	plan.Name = types.StringValue(u.Name)
	plan.Description = types.StringValue(u.Description)

	r.applyGrants(ctx, u.ID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.GetServiceUser(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading service user", err.Error())
		return
	}

	state.ID = types.StringValue(u.ID)
	state.Name = types.StringValue(u.Name)
	state.Description = types.StringValue(u.Description)

	if len(state.OrganizationGrant) > 0 {
		if g, err := r.client.GetServiceUserGrant(ctx, u.ID, "organization", 0); err == nil {
			state.OrganizationGrant = []orgGrantModel{serviceUserGrantToOrgModel(g)}
		} else {
			state.OrganizationGrant = []orgGrantModel{}
		}
	}
	if len(state.AllBranchDeploymentsGrant) > 0 {
		if g, err := r.client.GetServiceUserGrant(ctx, u.ID, "all_branch_deployments", 0); err == nil {
			state.AllBranchDeploymentsGrant = []orgGrantModel{serviceUserGrantToOrgModel(g)}
		} else {
			state.AllBranchDeploymentsGrant = []orgGrantModel{}
		}
	}

	prevDep := state.DeploymentGrant
	state.DeploymentGrant = nil
	for _, existing := range prevDep {
		intID, err := r.client.GetDeploymentIntID(ctx, existing.Deployment.ValueString())
		if err != nil {
			continue
		}
		g, err := r.client.GetServiceUserGrant(ctx, u.ID, "deployment", intID)
		if err != nil {
			continue
		}
		grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
		state.DeploymentGrant = append(state.DeploymentGrant, deploymentGrantModel{
			Deployment:   existing.Deployment,
			Grant:        grantVal,
			CustomRoleID: customRoleIDVal,
		})
	}

	prevBranch := state.BranchDeploymentsGrant
	state.BranchDeploymentsGrant = nil
	for _, existing := range prevBranch {
		intID, err := r.client.GetDeploymentIntID(ctx, existing.ParentDeployment.ValueString())
		if err != nil {
			continue
		}
		g, err := r.client.GetServiceUserGrant(ctx, u.ID, "branch_deployments", intID)
		if err != nil {
			continue
		}
		grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
		state.BranchDeploymentsGrant = append(state.BranchDeploymentsGrant, branchDeploymentsGrantModel{
			ParentDeployment: existing.ParentDeployment,
			Grant:            grantVal,
			CustomRoleID:     customRoleIDVal,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serviceUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serviceUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state serviceUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := r.client.UpdateServiceUser(ctx, state.ID.ValueString(), plan.Name.ValueString(), plan.Description.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating service user", err.Error())
		return
	}

	plan.ID = types.StringValue(u.ID)
	plan.Name = types.StringValue(u.Name)
	plan.Description = types.StringValue(u.Description)

	// Delete org/all-branch grants that were in state but not in plan.
	if len(plan.OrganizationGrant) == 0 && len(state.OrganizationGrant) > 0 {
		if err := r.client.DeleteServiceUserGrant(ctx, u.ID, "organization", 0); err != nil {
			resp.Diagnostics.AddError("Error removing organization grant", err.Error())
			return
		}
	}
	if len(plan.AllBranchDeploymentsGrant) == 0 && len(state.AllBranchDeploymentsGrant) > 0 {
		if err := r.client.DeleteServiceUserGrant(ctx, u.ID, "all_branch_deployments", 0); err != nil {
			resp.Diagnostics.AddError("Error removing all-branch-deployments grant", err.Error())
			return
		}
	}

	// Delete deployment grants no longer in plan.
	planDep := make(map[string]bool, len(plan.DeploymentGrant))
	for _, g := range plan.DeploymentGrant {
		planDep[g.Deployment.ValueString()] = true
	}
	for _, g := range state.DeploymentGrant {
		if !planDep[g.Deployment.ValueString()] {
			intID, err := r.client.GetDeploymentIntID(ctx, g.Deployment.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error resolving deployment", err.Error())
				return
			}
			if err := r.client.DeleteServiceUserGrant(ctx, u.ID, "deployment", intID); err != nil {
				resp.Diagnostics.AddError("Error removing deployment grant", err.Error())
				return
			}
		}
	}

	// Delete branch deployments grants no longer in plan.
	planBranch := make(map[string]bool, len(plan.BranchDeploymentsGrant))
	for _, g := range plan.BranchDeploymentsGrant {
		planBranch[g.ParentDeployment.ValueString()] = true
	}
	for _, g := range state.BranchDeploymentsGrant {
		if !planBranch[g.ParentDeployment.ValueString()] {
			intID, err := r.client.GetDeploymentIntID(ctx, g.ParentDeployment.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
				return
			}
			if err := r.client.DeleteServiceUserGrant(ctx, u.ID, "branch_deployments", intID); err != nil {
				resp.Diagnostics.AddError("Error removing branch deployments grant", err.Error())
				return
			}
		}
	}

	r.applyGrants(ctx, u.ID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serviceUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteServiceUser(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting service user", err.Error())
	}
}

func (r *serviceUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	u, err := r.client.GetServiceUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing service user", err.Error())
		return
	}

	// Inline grant blocks are not auto-populated on import. Add them to your
	// config to bring the corresponding API grants under management — the next
	// Read will then refresh those blocks from state.
	state := serviceUserResourceModel{
		ID:          types.StringValue(u.ID),
		Name:        types.StringValue(u.Name),
		Description: types.StringValue(u.Description),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
