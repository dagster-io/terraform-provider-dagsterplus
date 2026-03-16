package datasources

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &versionDataSource{}

func NewVersionDataSource() datasource.DataSource {
	return &versionDataSource{}
}

type versionDataSource struct {
	client *client.Client
}

type versionDataSourceModel struct {
	Deployment types.String `tfsdk:"deployment"`
	Version    types.String `tfsdk:"version"`
}

func (d *versionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_version"
}

func (d *versionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the version of Dagster running in a deployment.",
		Attributes: map[string]schema.Attribute{
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment to query.",
				Required:    true,
			},
			"version": schema.StringAttribute{
				Description: "The Dagster version string.",
				Computed:    true,
			},
		},
	}
}

func (d *versionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *versionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data versionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	version, err := d.client.GetVersion(ctx, data.Deployment.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading version", err.Error())
		return
	}

	data.Version = types.StringValue(version)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
