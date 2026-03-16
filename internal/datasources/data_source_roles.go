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

var _ datasource.DataSource = &rolesDataSource{}

func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

type rolesDataSource struct {
	client *client.Client
}

type rolesDataSourceModel struct {
	ID    types.String                  `tfsdk:"id"`
	Roles []resources.RoleResourceModel `tfsdk:"roles"`
}

func (d *rolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *rolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Dagster+ custom roles in the organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source.",
				Computed:    true,
			},
			"roles": schema.ListNestedAttribute{
				Description: "All custom roles in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The role ID assigned by Dagster+.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The role name.",
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *rolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	roles, err := d.client.ListRoles(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading roles", err.Error())
		return
	}

	state := rolesDataSourceModel{
		ID:    types.StringValue("roles"),
		Roles: make([]resources.RoleResourceModel, len(roles)),
	}
	for i := range roles {
		resp.Diagnostics.Append(resources.RoleToModel(ctx, &roles[i], &state.Roles[i])...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
