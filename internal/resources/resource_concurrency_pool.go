package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &concurrencyPoolResource{}
var _ resource.ResourceWithImportState = &concurrencyPoolResource{}

func NewConcurrencyPoolResource() resource.Resource {
	return &concurrencyPoolResource{}
}

type concurrencyPoolResource struct {
	client *client.Client
}

type concurrencyPoolResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Deployment types.String `tfsdk:"deployment"`
	Name       types.String `tfsdk:"name"`
	Limit      types.Int64  `tfsdk:"limit"`
}

func (r *concurrencyPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_concurrency_pool"
}

func (r *concurrencyPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a concurrency pool limit for a Dagster+ deployment. A pool is a named " +
			"concurrency key; assets and ops assigned to the pool (via the `pool=` argument in your " +
			"Dagster code) share the configured limit of in-progress executions. On destroy, the " +
			"explicit limit is removed and the pool reverts to the deployment's default pool limit.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier in the form '{deployment}/{name}'.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment the pool belongs to. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The pool name (concurrency key). Changing this forces a new resource.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"limit": schema.Int64Attribute{
				Description: "The maximum number of in-progress executions allowed for the pool (0-1000).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 1000),
				},
			},
		},
	}
}

func (r *concurrencyPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *concurrencyPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan concurrencyPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment := plan.Deployment.ValueString()
	name := plan.Name.ValueString()
	if err := r.client.SetConcurrencyPool(ctx, deployment, name, int(plan.Limit.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Error setting concurrency pool limit", err.Error())
		return
	}

	plan.ID = types.StringValue(concurrencyPoolID(deployment, name))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *concurrencyPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state concurrencyPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.GetConcurrencyPool(ctx, state.Deployment.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading concurrency pool", err.Error())
		return
	}

	// The pool no longer has an explicit limit (reverted to the default), so
	// treat it as deleted and let Terraform plan a recreate.
	if !pool.IsSet {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Limit = types.Int64Value(int64(pool.Limit))
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *concurrencyPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan concurrencyPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment := plan.Deployment.ValueString()
	name := plan.Name.ValueString()
	if err := r.client.SetConcurrencyPool(ctx, deployment, name, int(plan.Limit.ValueInt64())); err != nil {
		resp.Diagnostics.AddError("Error updating concurrency pool limit", err.Error())
		return
	}

	plan.ID = types.StringValue(concurrencyPoolID(deployment, name))
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *concurrencyPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state concurrencyPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteConcurrencyPool(ctx, state.Deployment.ValueString(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting concurrency pool", err.Error())
	}
}

func (r *concurrencyPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Expected ID in the form '{deployment}/{name}'.",
		)
		return
	}
	deployment := parts[0]
	name := parts[1]

	pool, err := r.client.GetConcurrencyPool(ctx, deployment, name)
	if err != nil {
		resp.Diagnostics.AddError("Error importing concurrency pool", err.Error())
		return
	}
	if !pool.IsSet {
		resp.Diagnostics.AddError(
			"Concurrency pool not found",
			fmt.Sprintf("Pool %q in deployment %q has no explicit limit set.", name, deployment),
		)
		return
	}

	state := concurrencyPoolResourceModel{
		ID:         types.StringValue(concurrencyPoolID(deployment, name)),
		Deployment: types.StringValue(deployment),
		Name:       types.StringValue(name),
		Limit:      types.Int64Value(int64(pool.Limit)),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func concurrencyPoolID(deployment, name string) string {
	return fmt.Sprintf("%s/%s", deployment, name)
}
