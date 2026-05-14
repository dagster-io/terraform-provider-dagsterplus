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

// Ensure userResource satisfies the resource.Resource interface.
var _ resource.Resource = &userResource{}
var _ resource.ResourceWithImportState = &userResource{}

// NewUserResource returns a new user resource.
func NewUserResource() resource.Resource {
	return &userResource{}
}

// userResource defines the resource implementation.
type userResource struct {
	client *client.Client
}

// userResourceModel describes the resource data model.
type userResourceModel struct {
	ID                        types.String                  `tfsdk:"id"`
	Email                     types.String                  `tfsdk:"email"`
	Name                      types.String                  `tfsdk:"name"`
	Role                      types.String                  `tfsdk:"role"`
	Picture                   types.String                  `tfsdk:"picture"`
	OrganizationGrant         []orgGrantModel               `tfsdk:"organization_grant"`
	AllBranchDeploymentsGrant []orgGrantModel               `tfsdk:"all_branch_deployments_grant"`
	DeploymentGrant           []deploymentGrantModel        `tfsdk:"deployment_grant"`
	BranchDeploymentsGrant    []branchDeploymentsGrantModel `tfsdk:"branch_deployments_grant"`
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	grantBlock := schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"grant": schema.StringAttribute{
				Description: "Standard permission level: VIEWER, LAUNCHER, EDITOR, or ADMIN. Conflicts with custom_role_id.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("VIEWER", "LAUNCHER", "EDITOR", "ADMIN"),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("custom_role_id")),
				},
			},
			"custom_role_id": schema.StringAttribute{
				Description: "The ID of a custom role to assign. Conflicts with grant.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("grant")),
				},
			},
		},
	}
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ organization member. Invites the user on create and removes them on delete.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The user ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user to invite. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The display name of the user (set by Dagster+ after the user accepts the invite).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Description: "The organization-level role as returned by Dagster+. Read-only mirror of organization_grant.",
				Computed:    true,
			},
			"picture": schema.StringAttribute{
				Description: "URL to the user's profile picture.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"organization_grant": schema.ListNestedBlock{
				Description:  "Organization-level permission grant for the user. At most one block is allowed.",
				NestedObject: grantBlock,
			},
			"all_branch_deployments_grant": schema.ListNestedBlock{
				Description:  "Permission grant for the user across all branch deployments. At most one block is allowed.",
				NestedObject: grantBlock,
			},
			"deployment_grant": schema.ListNestedBlock{
				Description: "Deployment-level permission grant for the user. One block per deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"deployment": schema.StringAttribute{
							Description: "The name of the deployment to grant access to.",
							Required:    true,
						},
						"grant": schema.StringAttribute{
							Description: "Standard permission level: VIEWER, LAUNCHER, EDITOR, or ADMIN. Conflicts with custom_role_id.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("VIEWER", "LAUNCHER", "EDITOR", "ADMIN"),
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("custom_role_id")),
							},
						},
						"custom_role_id": schema.StringAttribute{
							Description: "The ID of a custom role to assign. Conflicts with grant.",
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
							Description: "Standard permission level: VIEWER, LAUNCHER, EDITOR, or ADMIN. Conflicts with custom_role_id.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("VIEWER", "LAUNCHER", "EDITOR", "ADMIN"),
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("custom_role_id")),
							},
						},
						"custom_role_id": schema.StringAttribute{
							Description: "The ID of a custom role to assign. Conflicts with grant.",
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

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func userGrantToOrgModel(g *client.UserGrant) orgGrantModel {
	grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
	return orgGrantModel{Grant: grantVal, CustomRoleID: customRoleIDVal}
}

func (r *userResource) applyGrants(ctx context.Context, userID, email string, plan *userResourceModel, diags *diag.Diagnostics) {
	if len(plan.OrganizationGrant) > 0 {
		g := plan.OrganizationGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetUserGrant(ctx, client.UserGrant{
			UserID:          userID,
			Email:           email,
			DeploymentScope: "organization",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			diags.AddError("Error setting organization grant", err.Error())
			return
		}
		plan.OrganizationGrant = []orgGrantModel{userGrantToOrgModel(result)}
	}
	if len(plan.AllBranchDeploymentsGrant) > 0 {
		g := plan.AllBranchDeploymentsGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetUserGrant(ctx, client.UserGrant{
			UserID:          userID,
			Email:           email,
			DeploymentScope: "all_branch_deployments",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			diags.AddError("Error setting all-branch-deployments grant", err.Error())
			return
		}
		plan.AllBranchDeploymentsGrant = []orgGrantModel{userGrantToOrgModel(result)}
	}
	for i, g := range plan.DeploymentGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.Deployment.ValueString())
		if err != nil {
			diags.AddError("Error resolving deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetUserGrant(ctx, client.UserGrant{
			UserID:          userID,
			Email:           email,
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
		if _, err := r.client.SetUserGrant(ctx, client.UserGrant{
			UserID:          userID,
			Email:           email,
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

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := plan.Email.ValueString()
	if _, err := r.client.GetUserByEmail(ctx, email); err == nil {
		resp.Diagnostics.AddError("Error inviting user", fmt.Sprintf("user with email %s is already registered", email))
		return
	}

	user, err := r.client.AddUser(ctx, email)
	if err != nil {
		resp.Diagnostics.AddError("Error inviting user", err.Error())
		return
	}

	plan.ID = types.StringValue(user.ID)
	plan.Name = types.StringValue(user.Name)
	plan.Role = types.StringValue(user.Role)
	plan.Picture = types.StringValue(user.Picture)

	r.applyGrants(ctx, user.ID, user.Email, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-read so .role reflects any org grant just applied.
	if refreshed, err := r.client.GetUser(ctx, user.ID); err == nil {
		plan.Role = types.StringValue(refreshed.Role)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	state.ID = types.StringValue(user.ID)
	state.Email = types.StringValue(user.Email)
	state.Name = types.StringValue(user.Name)
	state.Role = types.StringValue(user.Role)
	state.Picture = types.StringValue(user.Picture)

	if len(state.OrganizationGrant) > 0 {
		if g, err := r.client.GetUserGrant(ctx, user.ID, "organization", 0); err == nil {
			state.OrganizationGrant = []orgGrantModel{userGrantToOrgModel(g)}
		} else {
			state.OrganizationGrant = []orgGrantModel{}
		}
	}
	if len(state.AllBranchDeploymentsGrant) > 0 {
		if g, err := r.client.GetUserGrant(ctx, user.ID, "all_branch_deployments", 0); err == nil {
			state.AllBranchDeploymentsGrant = []orgGrantModel{userGrantToOrgModel(g)}
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
		g, err := r.client.GetUserGrant(ctx, user.ID, "deployment", intID)
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
		g, err := r.client.GetUserGrant(ctx, user.ID, "branch_deployments", intID)
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

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state userResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := state.Email.ValueString()

	// Delete org/all-branch grants no longer in plan.
	if len(plan.OrganizationGrant) == 0 && len(state.OrganizationGrant) > 0 {
		if err := r.client.DeleteUserGrant(ctx, email, "organization", 0); err != nil {
			resp.Diagnostics.AddError("Error removing organization grant", err.Error())
			return
		}
	}
	if len(plan.AllBranchDeploymentsGrant) == 0 && len(state.AllBranchDeploymentsGrant) > 0 {
		if err := r.client.DeleteUserGrant(ctx, email, "all_branch_deployments", 0); err != nil {
			resp.Diagnostics.AddError("Error removing all-branch-deployments grant", err.Error())
			return
		}
	}
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
			if err := r.client.DeleteUserGrant(ctx, email, "deployment", intID); err != nil {
				resp.Diagnostics.AddError("Error removing deployment grant", err.Error())
				return
			}
		}
	}
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
			if err := r.client.DeleteUserGrant(ctx, email, "branch_deployments", intID); err != nil {
				resp.Diagnostics.AddError("Error removing branch deployments grant", err.Error())
				return
			}
		}
	}

	r.applyGrants(ctx, state.ID.ValueString(), email, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve computed attrs from state and refresh role.
	plan.ID = state.ID
	plan.Name = state.Name
	plan.Picture = state.Picture
	if refreshed, err := r.client.GetUser(ctx, state.ID.ValueString()); err == nil {
		plan.Role = types.StringValue(refreshed.Role)
	} else {
		plan.Role = state.Role
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.RemoveUser(ctx, state.Email.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing user", err.Error())
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	user, err := r.client.GetUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing user", err.Error())
		return
	}

	state := userResourceModel{
		ID:                        types.StringValue(user.ID),
		Email:                     types.StringValue(user.Email),
		Name:                      types.StringValue(user.Name),
		Role:                      types.StringValue(user.Role),
		Picture:                   types.StringValue(user.Picture),
		OrganizationGrant:         []orgGrantModel{},
		AllBranchDeploymentsGrant: []orgGrantModel{},
		DeploymentGrant:           []deploymentGrantModel{},
		BranchDeploymentsGrant:    []branchDeploymentsGrantModel{},
	}

	if g, err := r.client.GetUserGrant(ctx, user.ID, "organization", 0); err == nil {
		state.OrganizationGrant = []orgGrantModel{userGrantToOrgModel(g)}
	}
	if g, err := r.client.GetUserGrant(ctx, user.ID, "all_branch_deployments", 0); err == nil {
		state.AllBranchDeploymentsGrant = []orgGrantModel{userGrantToOrgModel(g)}
	}
	if depGrants, err := r.client.ListUserDeploymentGrants(ctx, user.ID); err == nil {
		for _, g := range depGrants {
			depName, err := r.client.GetDeploymentNameByIntID(ctx, g.DeploymentID)
			if err != nil {
				continue
			}
			grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
			state.DeploymentGrant = append(state.DeploymentGrant, deploymentGrantModel{
				Deployment:   types.StringValue(depName),
				Grant:        grantVal,
				CustomRoleID: customRoleIDVal,
			})
		}
	}
	if branchGrants, err := r.client.ListUserBranchDeploymentsGrants(ctx, user.ID); err == nil {
		for _, g := range branchGrants {
			depName, err := r.client.GetDeploymentNameByIntID(ctx, g.DeploymentID)
			if err != nil {
				continue
			}
			grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
			state.BranchDeploymentsGrant = append(state.BranchDeploymentsGrant, branchDeploymentsGrantModel{
				ParentDeployment: types.StringValue(depName),
				Grant:            grantVal,
				CustomRoleID:     customRoleIDVal,
			})
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
