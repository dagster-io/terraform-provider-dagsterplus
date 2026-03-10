package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &agentTokenDataSource{}

func NewAgentTokenDataSource() datasource.DataSource {
	return &agentTokenDataSource{}
}

type agentTokenDataSource struct {
	client *client.Client
}

type agentTokenDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *agentTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_token"
}

func (d *agentTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ agent token by label. " +
			"The token value is not available via this data source — it is only accessible at creation time.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The token ID assigned by Dagster+.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The label of the agent token to look up.",
				Required:    true,
			},
		},
	}
}

func (d *agentTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *agentTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config agentTokenDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokens, err := d.client.ListAgentTokens(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent tokens", err.Error())
		return
	}

	for _, t := range tokens {
		if t.Name == config.Name.ValueString() {
			resp.Diagnostics.Append(resp.State.Set(ctx, agentTokenDataSourceModel{
				ID:   types.StringValue(t.ID),
				Name: types.StringValue(t.Name),
			})...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Agent token not found",
		fmt.Sprintf("No agent token with label %q found in the organization.", config.Name.ValueString()),
	)
}
