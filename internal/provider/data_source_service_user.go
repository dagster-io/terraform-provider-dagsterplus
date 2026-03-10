package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &serviceUserDataSource{}

func NewServiceUserDataSource() datasource.DataSource {
	return &serviceUserDataSource{}
}

type serviceUserDataSource struct {
	client *client.Client
}

type serviceUserDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (d *serviceUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_user"
}

func (d *serviceUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ service user by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The service user ID assigned by Dagster+.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the service user to look up.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The service user description.",
				Computed:    true,
			},
		},
	}
}

func (d *serviceUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	u, err := d.client.GetServiceUserByName(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading service user", err.Error())
		return
	}

	state := serviceUserDataSourceModel{
		ID:          types.StringValue(u.ID),
		Name:        types.StringValue(u.Name),
		Description: types.StringValue(u.Description),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
