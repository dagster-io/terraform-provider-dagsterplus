package datasources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure deploymentDataSource satisfies the datasource.DataSource interface.
var _ datasource.DataSource = &deploymentDataSource{}

// NewDeploymentDataSource returns a new deployment data source.
func NewDeploymentDataSource() datasource.DataSource {
	return &deploymentDataSource{}
}

// deploymentDataSource defines the data source implementation.
type deploymentDataSource struct {
	client *client.Client
}

// deploymentDataSourceModel describes the data source data model.
type deploymentDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	Status       types.String `tfsdk:"status"`
	DeploymentID types.String `tfsdk:"deployment_id"`
}

func (d *deploymentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (d *deploymentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Deployment identifier (same as name).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the deployment to look up.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The deployment type: `SERVERLESS`, `HYBRID`, or `BRANCH`.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the deployment (e.g. `ACTIVE`, `PENDING_DELETION`).",
				Computed:    true,
			},
			"deployment_id": schema.StringAttribute{
				Description: "The numeric deployment ID as a decimal string.",
				Computed:    true,
			},
		},
	}
}

func (d *deploymentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *deploymentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config deploymentDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, err := d.client.GetDeployment(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployment", err.Error())
		return
	}

	state := deploymentDataSourceModel{
		ID:           types.StringValue(deployment.Name),
		Name:         types.StringValue(deployment.Name),
		Type:         types.StringValue(deployment.Type),
		Status:       types.StringValue(deployment.Status),
		DeploymentID: types.StringValue(strconv.FormatInt(deployment.IntID, 10)),
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
