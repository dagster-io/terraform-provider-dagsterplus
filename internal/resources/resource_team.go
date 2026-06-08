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

var _ resource.Resource = &teamResource{}
var _ resource.ResourceWithImportState = &teamResource{}

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *client.Client
}

type orgGrantModel struct {
	Grant        types.String `tfsdk:"grant"`
	CustomRoleID types.String `tfsdk:"custom_role_id"`
}

type deploymentGrantModel struct {
	Deployment   types.String `tfsdk:"deployment"`
	Grant        types.String `tfsdk:"grant"`
	CustomRoleID types.String `tfsdk:"custom_role_id"`
}

type branchDeploymentsGrantModel struct {
	ParentDeployment types.String `tfsdk:"parent_deployment"`
	Grant            types.String `tfsdk:"grant"`
	CustomRoleID     types.String `tfsdk:"custom_role_id"`
}

type memberModel struct {
	UserID types.String `tfsdk:"user_id"`
}

type teamResourceModel struct {
	ID                        types.String                  `tfsdk:"id"`
	Name                      types.String                  `tfsdk:"name"`
	OrganizationGrant         []orgGrantModel               `tfsdk:"organization_grant"`
	AllBranchDeploymentsGrant []orgGrantModel               `tfsdk:"all_branch_deployments_grant"`
	DeploymentGrant           []deploymentGrantModel        `tfsdk:"deployment_grant"`
	BranchDeploymentsGrant    []branchDeploymentsGrantModel `tfsdk:"branch_deployments_grant"`
	Members                   []memberModel                 `tfsdk:"member"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ team.",
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
			},
		},
		Blocks: map[string]schema.Block{
			"organization_grant": schema.ListNestedBlock{
				Description: "Organization-level permission grant for the team. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"grant": schema.StringAttribute{
							Description: withEnumValues(orgGrantDescription, orgGrantLevels),
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(orgGrantLevels...),
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
			"all_branch_deployments_grant": schema.ListNestedBlock{
				Description: "Permission grant for all branch deployments. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"grant": schema.StringAttribute{
							Description: withEnumValues(grantLevelDescription, grantLevels),
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(grantLevels...),
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
			"deployment_grant": schema.ListNestedBlock{
				Description: "Deployment-level permission grant for the team. One block per deployment.",
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
				Description: "Permission grant for all branch deployments of a specific parent (full) deployment. One block per parent deployment.",
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
			"member": schema.ListNestedBlock{
				Description: "A user who is a member of this team. One block per user.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"user_id": schema.StringAttribute{
							Description: "The ID of the user to add to the team.",
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

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.CreateTeam(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating team", err.Error())
		return
	}

	plan.ID = types.StringValue(team.ID)
	plan.Name = types.StringValue(team.Name)

	if len(plan.OrganizationGrant) > 0 {
		g := plan.OrganizationGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          team.ID,
			DeploymentScope: "organization",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error setting organization grant", err.Error())
			return
		}
		plan.OrganizationGrant = []orgGrantModel{grantToOrgModel(result)}
	}

	if len(plan.AllBranchDeploymentsGrant) > 0 {
		g := plan.AllBranchDeploymentsGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          team.ID,
			DeploymentScope: "all_branch_deployments",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error setting all-branch-deployments grant", err.Error())
			return
		}
		plan.AllBranchDeploymentsGrant = []orgGrantModel{grantToOrgModel(result)}
	}

	for i, g := range plan.DeploymentGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.Deployment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error resolving deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          team.ID,
			DeploymentScope: "deployment",
			DeploymentID:    intID,
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		}); err != nil {
			resp.Diagnostics.AddError("Error setting deployment grant", err.Error())
			return
		}
		plan.DeploymentGrant[i] = g
	}

	for i, g := range plan.BranchDeploymentsGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.ParentDeployment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          team.ID,
			DeploymentScope: "branch_deployments",
			DeploymentID:    intID,
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		}); err != nil {
			resp.Diagnostics.AddError("Error setting branch deployments grant", err.Error())
			return
		}
		plan.BranchDeploymentsGrant[i] = g
	}

	for _, m := range plan.Members {
		if err := r.client.AddUserToTeam(ctx, team.ID, m.UserID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error adding member to team", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.client.GetTeam(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading team", err.Error())
		return
	}

	state.ID = types.StringValue(team.ID)
	state.Name = types.StringValue(team.Name)

	// Only refresh org/branch grants if they are already in state (declared in config).
	// Orgs may have default grants applied to all teams by the API; fetching unconditionally
	// would cause perpetual drift for configs that don't declare these grants.
	if len(state.OrganizationGrant) > 0 {
		grant, err := r.client.GetTeamGrant(ctx, state.ID.ValueString(), "organization", 0)
		if err != nil {
			state.OrganizationGrant = []orgGrantModel{}
		} else {
			state.OrganizationGrant = []orgGrantModel{grantToOrgModel(grant)}
		}
	}

	if len(state.AllBranchDeploymentsGrant) > 0 {
		branchGrant, err := r.client.GetTeamGrant(ctx, state.ID.ValueString(), "all_branch_deployments", 0)
		if err != nil {
			state.AllBranchDeploymentsGrant = []orgGrantModel{}
		} else {
			state.AllBranchDeploymentsGrant = []orgGrantModel{grantToOrgModel(branchGrant)}
		}
	}

	// Rebuild deployment grants from API, preserving deployment names from prior state.
	prevDeploymentGrants := state.DeploymentGrant
	state.DeploymentGrant = nil
	for _, existing := range prevDeploymentGrants {
		intID, err := r.client.GetDeploymentIntID(ctx, existing.Deployment.ValueString())
		if err != nil {
			continue
		}
		g, err := r.client.GetTeamGrant(ctx, state.ID.ValueString(), "deployment", intID)
		if err != nil {
			continue // removed out-of-band; omit from state so Terraform detects drift
		}
		grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
		state.DeploymentGrant = append(state.DeploymentGrant, deploymentGrantModel{
			Deployment:   existing.Deployment,
			Grant:        grantVal,
			CustomRoleID: customRoleIDVal,
		})
	}

	// Rebuild branch deployments grants from API, preserving parent_deployment names from prior state.
	prevBranchGrants := state.BranchDeploymentsGrant
	state.BranchDeploymentsGrant = nil
	for _, existing := range prevBranchGrants {
		intID, err := r.client.GetDeploymentIntID(ctx, existing.ParentDeployment.ValueString())
		if err != nil {
			continue
		}
		g, err := r.client.GetTeamGrant(ctx, state.ID.ValueString(), "branch_deployments", intID)
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

	// Preserve members from prior state — the API query shape is not yet verified.
	// Members will be reconciled on Update if the config changes.

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan teamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		team, err := r.client.RenameTeam(ctx, plan.ID.ValueString(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error updating team", err.Error())
			return
		}
		plan.Name = types.StringValue(team.Name)
	}

	if len(plan.OrganizationGrant) > 0 {
		g := plan.OrganizationGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          plan.ID.ValueString(),
			DeploymentScope: "organization",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error setting organization grant", err.Error())
			return
		}
		plan.OrganizationGrant = []orgGrantModel{grantToOrgModel(result)}
	} else if len(state.OrganizationGrant) > 0 {
		if err := r.client.DeleteTeamGrant(ctx, plan.ID.ValueString(), "organization", 0); err != nil {
			resp.Diagnostics.AddError("Error removing organization grant", err.Error())
			return
		}
	}

	if len(plan.AllBranchDeploymentsGrant) > 0 {
		g := plan.AllBranchDeploymentsGrant[0]
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		result, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          plan.ID.ValueString(),
			DeploymentScope: "all_branch_deployments",
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error setting all-branch-deployments grant", err.Error())
			return
		}
		plan.AllBranchDeploymentsGrant = []orgGrantModel{grantToOrgModel(result)}
	} else if len(state.AllBranchDeploymentsGrant) > 0 {
		if err := r.client.DeleteTeamGrant(ctx, plan.ID.ValueString(), "all_branch_deployments", 0); err != nil {
			resp.Diagnostics.AddError("Error removing all-branch-deployments grant", err.Error())
			return
		}
	}

	// Build a set of deployment names in the plan for removal detection.
	planDeployments := make(map[string]bool, len(plan.DeploymentGrant))
	for _, g := range plan.DeploymentGrant {
		planDeployments[g.Deployment.ValueString()] = true
	}

	// Delete grants that were in state but are no longer in plan.
	for _, g := range state.DeploymentGrant {
		if !planDeployments[g.Deployment.ValueString()] {
			intID, err := r.client.GetDeploymentIntID(ctx, g.Deployment.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error resolving deployment", err.Error())
				return
			}
			if err := r.client.DeleteTeamGrant(ctx, plan.ID.ValueString(), "deployment", intID); err != nil {
				resp.Diagnostics.AddError("Error removing deployment grant", err.Error())
				return
			}
		}
	}

	// Upsert all grants in the plan.
	for i, g := range plan.DeploymentGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.Deployment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error resolving deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          plan.ID.ValueString(),
			DeploymentScope: "deployment",
			DeploymentID:    intID,
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		}); err != nil {
			resp.Diagnostics.AddError("Error setting deployment grant", err.Error())
			return
		}
		plan.DeploymentGrant[i] = g
	}

	// Sync branch_deployments_grant blocks: delete removed, upsert remaining.
	planBranchParents := make(map[string]bool, len(plan.BranchDeploymentsGrant))
	for _, g := range plan.BranchDeploymentsGrant {
		planBranchParents[g.ParentDeployment.ValueString()] = true
	}
	for _, g := range state.BranchDeploymentsGrant {
		if !planBranchParents[g.ParentDeployment.ValueString()] {
			intID, err := r.client.GetDeploymentIntID(ctx, g.ParentDeployment.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
				return
			}
			if err := r.client.DeleteTeamGrant(ctx, plan.ID.ValueString(), "branch_deployments", intID); err != nil {
				resp.Diagnostics.AddError("Error removing branch deployments grant", err.Error())
				return
			}
		}
	}
	for i, g := range plan.BranchDeploymentsGrant {
		intID, err := r.client.GetDeploymentIntID(ctx, g.ParentDeployment.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error resolving parent deployment", err.Error())
			return
		}
		grantStr, customRoleID := resolveGrantFields(g.Grant, g.CustomRoleID)
		if _, err := r.client.SetTeamGrant(ctx, client.TeamGrant{
			TeamID:          plan.ID.ValueString(),
			DeploymentScope: "branch_deployments",
			DeploymentID:    intID,
			Grant:           grantStr,
			CustomRoleID:    customRoleID,
		}); err != nil {
			resp.Diagnostics.AddError("Error setting branch deployments grant", err.Error())
			return
		}
		plan.BranchDeploymentsGrant[i] = g
	}

	// Sync members: remove those no longer in plan, add new ones.
	planMembers := make(map[string]bool, len(plan.Members))
	for _, m := range plan.Members {
		planMembers[m.UserID.ValueString()] = true
	}
	stateMembers := make(map[string]bool, len(state.Members))
	for _, m := range state.Members {
		stateMembers[m.UserID.ValueString()] = true
	}
	for userID := range stateMembers {
		if !planMembers[userID] {
			if err := r.client.RemoveUserFromTeam(ctx, plan.ID.ValueString(), userID); err != nil {
				resp.Diagnostics.AddError("Error removing member from team", err.Error())
				return
			}
		}
	}
	for userID := range planMembers {
		if !stateMembers[userID] {
			if err := r.client.AddUserToTeam(ctx, plan.ID.ValueString(), userID); err != nil {
				resp.Diagnostics.AddError("Error adding member to team", err.Error())
				return
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTeam(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting team", err.Error())
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	team, err := r.client.GetTeam(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing team", err.Error())
		return
	}

	state := teamResourceModel{
		ID:                        types.StringValue(team.ID),
		Name:                      types.StringValue(team.Name),
		OrganizationGrant:         []orgGrantModel{},
		AllBranchDeploymentsGrant: []orgGrantModel{},
		DeploymentGrant:           []deploymentGrantModel{},
		BranchDeploymentsGrant:    []branchDeploymentsGrantModel{},
		Members:                   []memberModel{},
	}

	if grant, err := r.client.GetTeamGrant(ctx, team.ID, "organization", 0); err == nil {
		state.OrganizationGrant = []orgGrantModel{grantToOrgModel(grant)}
	}

	if grant, err := r.client.GetTeamGrant(ctx, team.ID, "all_branch_deployments", 0); err == nil {
		state.AllBranchDeploymentsGrant = []orgGrantModel{grantToOrgModel(grant)}
	}

	if depGrants, err := r.client.ListTeamDeploymentGrants(ctx, team.ID); err == nil {
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

	if branchGrants, err := r.client.ListTeamBranchDeploymentsGrants(ctx, team.ID); err == nil {
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

func grantToOrgModel(g *client.TeamGrant) orgGrantModel {
	grantVal, customRoleIDVal := grantFieldsFromAPI(g.Grant, g.CustomRoleID)
	return orgGrantModel{Grant: grantVal, CustomRoleID: customRoleIDVal}
}

// resolveGrantFields converts model fields to the API grant+customRoleID pair.
// If custom_role_id is set, grant is implicitly "CUSTOM".
func resolveGrantFields(grant, customRoleID types.String) (string, string) {
	if !customRoleID.IsNull() && customRoleID.ValueString() != "" {
		return "CUSTOM", customRoleID.ValueString()
	}
	return grant.ValueString(), ""
}

// grantFieldsFromAPI converts API grant+customRoleID back to model fields.
// CUSTOM grants are represented solely by custom_role_id; grant is left null.
func grantFieldsFromAPI(grant, customRoleID string) (types.String, types.String) {
	if grant == "CUSTOM" {
		return types.StringNull(), nullableString(customRoleID)
	}
	return types.StringValue(grant), types.StringNull()
}

func nullableString(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
