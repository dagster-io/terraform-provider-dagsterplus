package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &secretDataSource{}

func NewSecretDataSource() datasource.DataSource {
	return &secretDataSource{}
}

type secretDataSource struct {
	client *client.Client
}

type secretDataSourceModel struct {
	ID                            types.String `tfsdk:"id"`
	SecretName                    types.String `tfsdk:"secret_name"`
	SecretValue                   types.String `tfsdk:"secret_value"`
	FullDeploymentScope           types.Bool   `tfsdk:"full_deployment_scope"`
	AllBranchDeploymentsScope     types.Bool   `tfsdk:"all_branch_deployments_scope"`
	SpecificBranchDeploymentScope types.String `tfsdk:"specific_branch_deployment_scope"`
	LocalDeploymentScope          types.Bool   `tfsdk:"local_deployment_scope"`
	LocationNames                 types.List   `tfsdk:"location_names"`
}

func (d *secretDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (d *secretDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ environment secret by name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The secret ID assigned by Dagster+.",
				Computed:    true,
			},
			"secret_name": schema.StringAttribute{
				Description: "The name of the secret to look up.",
				Required:    true,
			},
			"secret_value": schema.StringAttribute{
				Description: "The value of the secret.",
				Computed:    true,
				Sensitive:   true,
			},
			"full_deployment_scope": schema.BoolAttribute{
				Description: "Whether the secret is available to the full deployment.",
				Computed:    true,
			},
			"all_branch_deployments_scope": schema.BoolAttribute{
				Description: "Whether the secret is available to all branch deployments.",
				Computed:    true,
			},
			"specific_branch_deployment_scope": schema.StringAttribute{
				Description: "A specific branch deployment name the secret is scoped to.",
				Computed:    true,
			},
			"local_deployment_scope": schema.BoolAttribute{
				Description: "Whether the secret is available to local deployments.",
				Computed:    true,
			},
			"location_names": schema.ListAttribute{
				Description: "Code location names the secret is scoped to.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *secretDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *secretDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config secretDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := d.client.GetSecretByName(ctx, config.SecretName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading secret", err.Error())
		return
	}

	state, diags := secretToModel(ctx, s)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map to data source model
	dsModel := secretDataSourceModel{
		ID:                            state.ID,
		SecretName:                    state.SecretName,
		SecretValue:                   state.SecretValue,
		FullDeploymentScope:           state.FullDeploymentScope,
		AllBranchDeploymentsScope:     state.AllBranchDeploymentsScope,
		SpecificBranchDeploymentScope: state.SpecificBranchDeploymentScope,
		LocalDeploymentScope:          state.LocalDeploymentScope,
		LocationNames:                 state.LocationNames,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, dsModel)...)
}
