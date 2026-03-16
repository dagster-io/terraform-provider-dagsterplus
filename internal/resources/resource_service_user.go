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

var _ resource.Resource = &serviceUserResource{}
var _ resource.ResourceWithImportState = &serviceUserResource{}

func NewServiceUserResource() resource.Resource {
	return &serviceUserResource{}
}

type serviceUserResource struct {
	client *client.Client
}

type serviceUserResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *serviceUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (r *serviceUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

	state := serviceUserResourceModel{
		ID:          types.StringValue(u.ID),
		Name:        types.StringValue(u.Name),
		Description: types.StringValue(u.Description),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
