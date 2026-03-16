package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &customMetricResource{}
var _ resource.ResourceWithImportState = &customMetricResource{}

func NewCustomMetricResource() resource.Resource {
	return &customMetricResource{}
}

type customMetricResource struct {
	client *client.Client
}

type customMetricResourceModel struct {
	ID              types.String  `tfsdk:"id"`
	MetadataKey     types.String  `tfsdk:"metadata_key"`
	DisplayName     types.String  `tfsdk:"display_name"`
	Description     types.String  `tfsdk:"description"`
	CreateTimestamp types.Float64 `tfsdk:"create_timestamp"`
	UpdateTimestamp types.Float64 `tfsdk:"update_timestamp"`
}

func (r *customMetricResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_metric"
}

func (r *customMetricResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ custom metric.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The custom metric ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"metadata_key": schema.StringAttribute{
				Description: "The metadata key used to identify the custom metric. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "An optional human-readable display name for the metric.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "An optional description for the metric.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"create_timestamp": schema.Float64Attribute{
				Description: "Unix timestamp of when the custom metric was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"update_timestamp": schema.Float64Attribute{
				Description: "Unix timestamp of when the custom metric was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *customMetricResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customMetricResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.CreateCustomMetric(ctx,
		plan.MetadataKey.ValueString(),
		plan.DisplayName.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error creating custom metric", err.Error())
		return
	}

	plan.ID = types.StringValue(m.ID)
	plan.MetadataKey = types.StringValue(m.MetadataKey)
	plan.DisplayName = types.StringValue(m.DisplayName)
	plan.Description = types.StringValue(m.Description)
	plan.CreateTimestamp = types.Float64Value(m.CreateTimestamp)
	plan.UpdateTimestamp = types.Float64Value(m.UpdateTimestamp)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *customMetricResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.GetCustomMetric(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading custom metric", err.Error())
		return
	}

	state.ID = types.StringValue(m.ID)
	state.MetadataKey = types.StringValue(m.MetadataKey)
	state.DisplayName = types.StringValue(m.DisplayName)
	state.Description = types.StringValue(m.Description)
	state.CreateTimestamp = types.Float64Value(m.CreateTimestamp)
	state.UpdateTimestamp = types.Float64Value(m.UpdateTimestamp)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *customMetricResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan customMetricResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state customMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.UpdateCustomMetric(ctx,
		state.ID.ValueString(),
		plan.DisplayName.ValueString(),
		plan.Description.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Error updating custom metric", err.Error())
		return
	}

	plan.ID = types.StringValue(m.ID)
	plan.MetadataKey = types.StringValue(m.MetadataKey)
	plan.DisplayName = types.StringValue(m.DisplayName)
	plan.Description = types.StringValue(m.Description)
	plan.CreateTimestamp = types.Float64Value(m.CreateTimestamp)
	plan.UpdateTimestamp = types.Float64Value(m.UpdateTimestamp)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *customMetricResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customMetricResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCustomMetric(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "PythonError") {
			return
		}
		resp.Diagnostics.AddError("Error deleting custom metric", err.Error())
	}
}

func (r *customMetricResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	m, err := r.client.GetCustomMetric(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing custom metric", err.Error())
		return
	}

	state := customMetricResourceModel{
		ID:              types.StringValue(m.ID),
		MetadataKey:     types.StringValue(m.MetadataKey),
		DisplayName:     types.StringValue(m.DisplayName),
		Description:     types.StringValue(m.Description),
		CreateTimestamp: types.Float64Value(m.CreateTimestamp),
		UpdateTimestamp: types.Float64Value(m.UpdateTimestamp),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
