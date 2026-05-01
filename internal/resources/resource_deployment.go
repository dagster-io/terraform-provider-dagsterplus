package resources

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure deploymentResource satisfies the resource.Resource interface.
var _ resource.Resource = &deploymentResource{}
var _ resource.ResourceWithImportState = &deploymentResource{}

// NewDeploymentResource returns a new deployment resource.
func NewDeploymentResource() resource.Resource {
	return &deploymentResource{}
}

// deploymentResource defines the resource implementation.
type deploymentResource struct {
	client *client.Client
}

// deploymentResourceModel describes the resource data model.
type deploymentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Status       types.String `tfsdk:"status"`
	DeploymentID types.String `tfsdk:"deployment_id"`
}

func (r *deploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *deploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Deployment identifier (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the deployment. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "The current status of the deployment (e.g. ACTIVE, PENDING_DELETION).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment_id": schema.StringAttribute{
				Description: "The numeric deployment ID as a decimal string. Use this with GraphQL mutations such as `createOrUpdateAgentPermissions`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *deploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deploymentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.CreateDeployment(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating deployment", err.Error())
		return
	}

	plan.ID = types.StringValue(deployment.Name)
	plan.Name = types.StringValue(deployment.Name)
	plan.Status = types.StringValue(deployment.Status)
	plan.DeploymentID = types.StringValue(strconv.FormatInt(deployment.IntID, 10))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := r.client.GetDeployment(ctx, state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading deployment", err.Error())
		return
	}

	state.ID = types.StringValue(deployment.Name)
	state.Name = types.StringValue(deployment.Name)
	state.Status = types.StringValue(deployment.Status)
	state.DeploymentID = types.StringValue(strconv.FormatInt(deployment.IntID, 10))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *deploymentResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// name is RequiresReplace — Update is never called.
}

func (r *deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deploymentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDeployment(ctx, state.Name.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting deployment", err.Error())
	}
}

func (r *deploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	deployment, err := r.client.GetDeployment(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing deployment", err.Error())
		return
	}

	state := deploymentResourceModel{
		ID:           types.StringValue(deployment.Name),
		Name:         types.StringValue(deployment.Name),
		Status:       types.StringValue(deployment.Status),
		DeploymentID: types.StringValue(strconv.FormatInt(deployment.IntID, 10)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
