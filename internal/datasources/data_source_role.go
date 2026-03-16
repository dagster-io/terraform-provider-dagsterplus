package datasources

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	resources "github.com/dagster-io/terraform-provider-dagsterplus/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &roleDataSource{}

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

type roleDataSource struct {
	client *client.Client
}

func (d *roleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ custom role by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The role ID assigned by Dagster+.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the role to look up.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A human-readable description of the role.",
				Computed:    true,
			},
			"icon": schema.StringAttribute{
				Description: "The icon name for the role.",
				Computed:    true,
			},
			"role_type": schema.StringAttribute{
				Description: "The scope of the role: deployment or organization.",
				Computed:    true,
			},
			"permissions": schema.SetAttribute{
				Description: "The permissions granted by this role.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *roleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config resources.RoleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRoleByName(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading role", err.Error())
		return
	}

	resp.Diagnostics.Append(resources.RoleToModel(ctx, role, &config)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
