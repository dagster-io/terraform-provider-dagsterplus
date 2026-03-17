package datasources

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &userDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.Client
}

type userDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Email             types.String `tfsdk:"email"`
	Name              types.String `tfsdk:"name"`
	Role              types.String `tfsdk:"role"`
	Picture           types.String `tfsdk:"picture"`
	IsScimProvisioned types.Bool   `tfsdk:"is_scim_provisioned"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ organization member by email address.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The user ID assigned by Dagster+.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the user to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The display name of the user.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The organization-level role: VIEWER, EDITOR, ADMIN, or OWNER.",
				Computed:    true,
			},
			"picture": schema.StringAttribute{
				Description: "URL to the user's profile picture.",
				Computed:    true,
			},
			"is_scim_provisioned": schema.BoolAttribute{
				Description: "Whether this user was provisioned through SCIM.",
				Computed:    true,
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading users", err.Error())
		return
	}

	for _, u := range users {
		if u.Email == config.Email.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, userDataSourceModel{
				ID:                types.StringValue(u.ID),
				Email:             types.StringValue(u.Email),
				Name:              types.StringValue(u.Name),
				Role:              types.StringValue(u.Role),
				Picture:           types.StringValue(u.Picture),
				IsScimProvisioned: types.BoolValue(u.IsScimProvisioned),
			})...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"User not found",
		fmt.Sprintf("No user with email %q found in the organization.", config.Email.ValueString()),
	)
}
