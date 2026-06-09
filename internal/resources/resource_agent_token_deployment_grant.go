package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &agentTokenDeploymentGrantResource{}
var _ resource.ResourceWithImportState = &agentTokenDeploymentGrantResource{}

func NewAgentTokenDeploymentGrantResource() resource.Resource {
	return &agentTokenDeploymentGrantResource{}
}

type agentTokenDeploymentGrantResource struct {
	client *client.Client
}

type agentTokenDeploymentGrantResourceModel struct {
	ID           types.String `tfsdk:"id"`
	AgentTokenID types.String `tfsdk:"agent_token_id"`
	Deployment   types.String `tfsdk:"deployment"`
}

func (r *agentTokenDeploymentGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_token_deployment_grant"
}

func (r *agentTokenDeploymentGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Grants an agent token the AGENT permission on a specific deployment. " +
			"Agent tokens only ever carry the AGENT permission, so there is no grant level to configure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form `{agent_token_id}/{deployment}`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent_token_id": schema.StringAttribute{
				Description: "The ID of the agent token to grant access. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment to grant access to. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *agentTokenDeploymentGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *agentTokenDeploymentGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan agentTokenDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, plan.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	if _, err := r.client.SetAgentTokenGrant(ctx, client.AgentTokenGrant{
		AgentTokenID:    plan.AgentTokenID.ValueString(),
		DeploymentScope: "deployment",
		DeploymentID:    intID,
	}); err != nil {
		resp.Diagnostics.AddError("Error setting deployment grant", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.AgentTokenID.ValueString() + "/" + plan.Deployment.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *agentTokenDeploymentGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state agentTokenDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	if _, err := r.client.GetAgentTokenGrant(ctx, state.AgentTokenID.ValueString(), "deployment", intID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading deployment grant", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *agentTokenDeploymentGrantResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// All attributes are RequiresReplace — Update is never called.
}

func (r *agentTokenDeploymentGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state agentTokenDeploymentGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	intID, err := r.client.GetDeploymentIntID(ctx, state.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	if err := r.client.DeleteAgentTokenGrant(ctx, state.AgentTokenID.ValueString(), "deployment", intID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error removing deployment grant", err.Error())
	}
}

func (r *agentTokenDeploymentGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{agent_token_id}/{deployment}'.",
		)
		return
	}

	agentTokenID := parts[0]
	deployment := parts[1]

	intID, err := r.client.GetDeploymentIntID(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving deployment", err.Error())
		return
	}

	if _, err := r.client.GetAgentTokenGrant(ctx, agentTokenID, "deployment", intID); err != nil {
		resp.Diagnostics.AddError("Error reading deployment grant", err.Error())
		return
	}

	state := agentTokenDeploymentGrantResourceModel{
		ID:           types.StringValue(req.ID),
		AgentTokenID: types.StringValue(agentTokenID),
		Deployment:   types.StringValue(deployment),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
