package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &userTokenDataSource{}

func NewUserTokenDataSource() datasource.DataSource {
	return &userTokenDataSource{}
}

type userTokenDataSource struct {
	client *client.Client
}

type userTokenDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	UserID types.String `tfsdk:"user_id"`
}

func (d *userTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_token"
}

func (d *userTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ user token by label. " +
			"The token value is not available via this data source — it is only accessible at creation time.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The token ID assigned by Dagster+.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The label of the user token to look up.",
				Required:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user whose tokens to search.",
				Required:    true,
			},
		},
	}
}

func (d *userTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userTokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokens, err := d.client.ListUserTokens(ctx, config.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user tokens", err.Error())
		return
	}

	for _, t := range tokens {
		if t.Name == config.Name.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, userTokenDataSourceModel{
				ID:     types.StringValue(t.ID),
				Name:   types.StringValue(t.Name),
				UserID: config.UserID,
			})...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"User token not found",
		fmt.Sprintf("No user token with label %q found.", config.Name.ValueString()),
	)
}
