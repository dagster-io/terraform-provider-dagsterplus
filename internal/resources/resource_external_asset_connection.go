package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &externalAssetConnectionResource{}
var _ resource.ResourceWithImportState = &externalAssetConnectionResource{}

func NewExternalAssetConnectionResource() resource.Resource {
	return &externalAssetConnectionResource{}
}

type externalAssetConnectionResource struct {
	client *client.Client
}

type externalAssetConnectionResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ConnectionType   types.String `tfsdk:"connection_type"`
	SourceConfigYaml types.String `tfsdk:"source_config_yaml"`
	CronString       types.String `tfsdk:"cron_string"`
	Timezone         types.String `tfsdk:"timezone"`
	ScheduleStatus   types.String `tfsdk:"schedule_status"`
}

func (r *externalAssetConnectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_asset_connection"
}

func (r *externalAssetConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ external asset connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The connection ID assigned by Dagster+.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the external asset connection.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description for the connection.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_type": schema.StringAttribute{
				Description: withEnumValues("The type of external connection. Changing this forces a new resource.", externalAssetConnectionTypes),
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(externalAssetConnectionTypes...),
				},
			},
			"source_config_yaml": schema.StringAttribute{
				Description: "YAML configuration for the connection source. May contain credentials — treated as sensitive.",
				Required:    true,
				Sensitive:   true,
			},
			"cron_string": schema.StringAttribute{
				Description: "Cron expression for the connection schedule.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timezone": schema.StringAttribute{
				Description: "Timezone for the connection schedule.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule_status": schema.StringAttribute{
				Description: "The current schedule status: `RUNNING` or `STOPPED`.",
				Computed:    true,
			},
		},
	}
}

func (r *externalAssetConnectionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *externalAssetConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan externalAssetConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := r.client.CreateExternalAssetConnection(ctx, externalAssetConnectionFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating external asset connection", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, externalAssetConnectionToModel(conn))...)
}

func (r *externalAssetConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state externalAssetConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn, err := r.client.GetExternalAssetConnection(ctx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading external asset connection", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, externalAssetConnectionToModel(conn))...)
}

func (r *externalAssetConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan externalAssetConnectionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state externalAssetConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := externalAssetConnectionFromModel(plan)
	c.ID = state.ID.ValueString()

	conn, err := r.client.UpdateExternalAssetConnection(ctx, c)
	if err != nil {
		resp.Diagnostics.AddError("Error updating external asset connection", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, externalAssetConnectionToModel(conn))...)
}

func (r *externalAssetConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state externalAssetConnectionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteExternalAssetConnection(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError("Error deleting external asset connection", err.Error())
	}
}

func (r *externalAssetConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	conn, err := r.client.GetExternalAssetConnection(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing external asset connection", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, externalAssetConnectionToModel(conn))...)
}

func externalAssetConnectionFromModel(m externalAssetConnectionResourceModel) client.ExternalAssetConnection {
	return client.ExternalAssetConnection{
		Name:             m.Name.ValueString(),
		Description:      m.Description.ValueString(),
		ConnectionType:   m.ConnectionType.ValueString(),
		SourceConfigYaml: m.SourceConfigYaml.ValueString(),
		CronString:       m.CronString.ValueString(),
		Timezone:         m.Timezone.ValueString(),
	}
}

func externalAssetConnectionToModel(c *client.ExternalAssetConnection) externalAssetConnectionResourceModel {
	return externalAssetConnectionResourceModel{
		ID:               types.StringValue(c.ID),
		Name:             types.StringValue(c.Name),
		Description:      types.StringValue(c.Description),
		ConnectionType:   types.StringValue(c.ConnectionType),
		SourceConfigYaml: types.StringValue(c.SourceConfigYaml),
		CronString:       types.StringValue(c.CronString),
		Timezone:         types.StringValue(c.Timezone),
		ScheduleStatus:   types.StringValue(c.ScheduleStatus),
	}
}
