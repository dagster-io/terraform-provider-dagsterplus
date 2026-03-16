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
type agentTokenResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Token types.String `tfsdk:"token"`
}

func (r *agentTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_token"
}

func (r *agentTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ agent token. " +
			"The token value is only available at creation time; it cannot be recovered after import. " +
			"Changing the name forces a new resource.",
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

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *agentTokenResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// Name is RequiresReplace — Update is never called.
}

func (r *agentTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state agentTokenResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAgentToken(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting agent token", err.Error())
	}
}

func (r *agentTokenResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Token value is irrecoverable after creation.
	// Name may not be resolvable if the list API field name is unavailable.
	state := agentTokenResourceModel{
		ID:    types.StringValue(req.ID),
		Name:  types.StringValue(""),
		Token: types.StringValue(""),
	}

	if tok, err := r.client.GetAgentToken(ctx, req.ID); err == nil {
		state.Name = types.StringValue(tok.Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
