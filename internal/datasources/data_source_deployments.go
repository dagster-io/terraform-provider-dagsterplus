package datasources

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &deploymentsDataSource{}

func NewDeploymentsDataSource() datasource.DataSource {
	return &deploymentsDataSource{}
}

type deploymentsDataSource struct {
	client *client.Client
}

type deploymentsDataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Deployments []deploymentDataSourceModel `tfsdk:"deployments"`
}

func (d *deploymentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *deploymentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Dagster+ deployments in the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source.",
				Computed:    true,
			},
			"deployments": schema.ListNestedAttribute{
				Description: "All deployments in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Deployment identifier (same as name).",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The deployment name.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The deployment type: `SERVERLESS`, `HYBRID`, or `BRANCH`.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The deployment status.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *deploymentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deploymentsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	deployments, err := d.client.ListDeployments(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading deployments", err.Error())
		return
	}

	state := deploymentsDataSourceModel{
		ID:          types.StringValue("deployments"),
		Deployments: make([]deploymentDataSourceModel, len(deployments)),
	}
	for i, dep := range deployments {
		state.Deployments[i] = deploymentDataSourceModel{
			ID:   types.StringValue(dep.Name),
			Name: types.StringValue(dep.Name),
			Type: types.StringValue(dep.Type),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
