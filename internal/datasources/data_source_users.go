package datasources

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &usersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *client.Client
}

type usersDataSourceModel struct {
	ID    types.String          `tfsdk:"id"`
	Users []userDataSourceModel `tfsdk:"users"`
}

func (d *usersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Dagster+ organization members.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source.",
				Computed:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "All users in the organization.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The user ID assigned by Dagster+.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address of the user.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The display name of the user.",
							Computed:    true,
						},
						"role": schema.StringAttribute{
							Description: "The organization-level role: VIEWER, EDITOR, ADMIN, or OWNER.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	users, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading users", err.Error())
		return
	}

	state := usersDataSourceModel{
		ID:    types.StringValue("users"),
		Users: make([]userDataSourceModel, len(users)),
	}
	for i, u := range users {
		state.Users[i] = userDataSourceModel{
			ID:    types.StringValue(u.ID),
			Email: types.StringValue(u.Email),
			Name:  types.StringValue(u.Name),
			Role:  types.StringValue(u.Role),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
