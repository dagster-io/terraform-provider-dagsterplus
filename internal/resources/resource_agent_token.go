package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure agentTokenResource satisfies the resource.Resource interface.
var _ resource.Resource = &agentTokenResource{}
var _ resource.ResourceWithImportState = &agentTokenResource{}

// NewAgentTokenResource returns a new agent token resource.
func NewAgentTokenResource() resource.Resource {
	return &agentTokenResource{}
}

// agentTokenResource defines the resource implementation.
type agentTokenResource struct {
	client *client.Client
}

// agentTokenResourceModel describes the resource data model.
//
// Agent tokens only ever carry the AGENT grant, so the inline grant attributes
// only express which scopes/deployments the token is granted on — there is no
// grant level, custom role, or location grant to configure.
type agentTokenResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Token                types.String `tfsdk:"token"`
	Organization         types.Bool   `tfsdk:"organization"`
	AllBranchDeployments types.Bool   `tfsdk:"all_branch_deployments"`
	DeploymentGrants     types.Set    `tfsdk:"deployment_grants"`
}

func (r *agentTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_token"
}

func (r *agentTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ agent token and its permission grants. " +
			"The token value is only available at creation time; it cannot be recovered after import. " +
			"Changing the name forces a new resource. Agent tokens only ever carry the AGENT permission, " +
			"so the grant attributes only control which scopes the token is granted on. " +
			"Dagster+ supports three agent-token grant scopes: the organization, specific full deployments, " +
			"and all branch deployments. Unlike user/team grants, agent tokens cannot be scoped to the branch " +
			"deployments of a single parent deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The token ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The label for the agent token. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The agent token value. Only populated on creation; not recoverable after import.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization": schema.BoolAttribute{
				Description: "Whether the token holds the organization-scoped AGENT grant. " +
					"Dagster+ creates this grant automatically with every new token; set to false to remove it. " +
					"Defaults to true.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"all_branch_deployments": schema.BoolAttribute{
				Description: "Whether the token holds the AGENT grant across all branch deployments. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"deployment_grants": schema.SetAttribute{
				Description: "Names of full deployments the token is granted the AGENT permission on.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *agentTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// applyGrants creates/updates every grant present in the plan. It never deletes;
// removals are reconciled by the caller (Create removes the default org grant,
// Update diffs against prior state).
func (r *agentTokenResource) applyGrants(ctx context.Context, tokenID string, plan *agentTokenResourceModel, diags *diag.Diagnostics) {
	if plan.Organization.ValueBool() {
		if _, err := r.client.SetAgentTokenGrant(ctx, client.AgentTokenGrant{
			AgentTokenID:    tokenID,
			DeploymentScope: "organization",
		}); err != nil {
			diags.AddError("Error setting organization grant", err.Error())
			return
		}
	}
	if plan.AllBranchDeployments.ValueBool() {
		if _, err := r.client.SetAgentTokenGrant(ctx, client.AgentTokenGrant{
			AgentTokenID:    tokenID,
			DeploymentScope: "all_branch_deployments",
		}); err != nil {
			diags.AddError("Error setting all-branch-deployments grant", err.Error())
			return
		}
	}

	var depNames []string
	diags.Append(plan.DeploymentGrants.ElementsAs(ctx, &depNames, false)...)
	if diags.HasError() {
		return
	}
	for _, name := range depNames {
		intID, err := r.client.GetDeploymentIntID(ctx, name)
		if err != nil {
			diags.AddError("Error resolving deployment", err.Error())
			return
		}
		if _, err := r.client.SetAgentTokenGrant(ctx, client.AgentTokenGrant{
			AgentTokenID:    tokenID,
			DeploymentScope: "deployment",
			DeploymentID:    intID,
		}); err != nil {
			diags.AddError("Error setting deployment grant", err.Error())
			return
		}
	}
}

func (r *agentTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan agentTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.CreateAgentToken(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent token", err.Error())
		return
	}

	plan.ID = types.StringValue(tok.ID)
	plan.Name = types.StringValue(tok.Name)
	plan.Token = types.StringValue(tok.Token)

	// Dagster+ auto-creates an organization-scoped AGENT grant with every new
	// token. If the user opted out, remove it before applying the rest.
	if !plan.Organization.ValueBool() {
		if err := r.client.DeleteAgentTokenGrant(ctx, tok.ID, "organization", 0); err != nil {
			resp.Diagnostics.AddError("Error removing organization grant", err.Error())
			return
		}
	}

	r.applyGrants(ctx, tok.ID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *agentTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state agentTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tok, err := r.client.GetAgentToken(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading agent token", err.Error())
		return
	}

	// Preserve the token value from state — the API never returns it after creation.
	state.ID = types.StringValue(tok.ID)
	state.Name = types.StringValue(tok.Name)
	// state.Token is intentionally not overwritten.

	_, orgErr := r.client.GetAgentTokenGrant(ctx, tok.ID, "organization", 0)
	state.Organization = types.BoolValue(orgErr == nil)

	_, allBranchErr := r.client.GetAgentTokenGrant(ctx, tok.ID, "all_branch_deployments", 0)
	state.AllBranchDeployments = types.BoolValue(allBranchErr == nil)

	state.DeploymentGrants = r.refreshDeploymentNames(ctx, tok.ID, state.DeploymentGrants, "deployment", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// refreshDeploymentNames keeps only the deployment names from prior state whose
// grant still exists on the token. Grants added outside Terraform are not
// surfaced (the resource manages only the names listed in config), matching the
// inline-grant behavior of the user / service-user resources.
func (r *agentTokenResource) refreshDeploymentNames(ctx context.Context, tokenID string, prior types.Set, scope string, diags *diag.Diagnostics) types.Set {
	if prior.IsNull() || prior.IsUnknown() {
		return prior
	}
	var names []string
	diags.Append(prior.ElementsAs(ctx, &names, false)...)
	if diags.HasError() {
		return prior
	}
	var kept []attr.Value
	for _, name := range names {
		intID, err := r.client.GetDeploymentIntID(ctx, name)
		if err != nil {
			continue
		}
		if _, err := r.client.GetAgentTokenGrant(ctx, tokenID, scope, intID); err == nil {
			kept = append(kept, types.StringValue(name))
		}
	}
	set, d := types.SetValue(types.StringType, kept)
	diags.Append(d...)
	return set
}

func (r *agentTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan agentTokenResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state agentTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokenID := state.ID.ValueString()

	// Remove org / all-branch grants that were enabled in state but disabled in plan.
	if !plan.Organization.ValueBool() && state.Organization.ValueBool() {
		if err := r.client.DeleteAgentTokenGrant(ctx, tokenID, "organization", 0); err != nil {
			resp.Diagnostics.AddError("Error removing organization grant", err.Error())
			return
		}
	}
	if !plan.AllBranchDeployments.ValueBool() && state.AllBranchDeployments.ValueBool() {
		if err := r.client.DeleteAgentTokenGrant(ctx, tokenID, "all_branch_deployments", 0); err != nil {
			resp.Diagnostics.AddError("Error removing all-branch-deployments grant", err.Error())
			return
		}
	}

	// Remove deployment grants no longer in plan.
	r.deleteRemovedDeploymentGrants(ctx, tokenID, plan.DeploymentGrants, state.DeploymentGrants, "deployment", &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.applyGrants(ctx, tokenID, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *agentTokenResource) deleteRemovedDeploymentGrants(ctx context.Context, tokenID string, plan, state types.Set, scope string, diags *diag.Diagnostics) {
	var planNames, stateNames []string
	diags.Append(plan.ElementsAs(ctx, &planNames, false)...)
	diags.Append(state.ElementsAs(ctx, &stateNames, false)...)
	if diags.HasError() {
		return
	}
	inPlan := make(map[string]bool, len(planNames))
	for _, n := range planNames {
		inPlan[n] = true
	}
	for _, name := range stateNames {
		if inPlan[name] {
			continue
		}
		intID, err := r.client.GetDeploymentIntID(ctx, name)
		if err != nil {
			diags.AddError("Error resolving deployment", err.Error())
			return
		}
		if err := r.client.DeleteAgentTokenGrant(ctx, tokenID, scope, intID); err != nil {
			diags.AddError("Error removing deployment grant", err.Error())
			return
		}
	}
}

func (r *agentTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state agentTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAgentToken(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting agent token", err.Error())
	}
}

func (r *agentTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Token value is irrecoverable after creation.
	state := agentTokenResourceModel{
		ID:                   types.StringValue(req.ID),
		Name:                 types.StringValue(""),
		Token:                types.StringValue(""),
		Organization:         types.BoolValue(true),
		AllBranchDeployments: types.BoolValue(false),
		DeploymentGrants:     types.SetNull(types.StringType),
	}

	if tok, err := r.client.GetAgentToken(ctx, req.ID); err == nil {
		state.Name = types.StringValue(tok.Name)
	}
	if _, err := r.client.GetAgentTokenGrant(ctx, req.ID, "organization", 0); err != nil {
		state.Organization = types.BoolValue(false)
	}
	if _, err := r.client.GetAgentTokenGrant(ctx, req.ID, "all_branch_deployments", 0); err == nil {
		state.AllBranchDeployments = types.BoolValue(true)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
