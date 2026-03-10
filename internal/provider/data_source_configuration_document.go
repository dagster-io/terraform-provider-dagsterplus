package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"sigs.k8s.io/yaml"
)

var _ datasource.DataSource = &configurationDocumentDataSource{}

func NewConfigurationDocumentDataSource() datasource.DataSource {
	return &configurationDocumentDataSource{}
}

type configurationDocumentDataSource struct {
	client *client.Client
}

type configurationDocumentDataSourceModel struct {
	YAMLBody types.String `tfsdk:"yaml_body"`
	JSON     types.String `tfsdk:"json"`
}

func (d *configurationDocumentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration_document"
}

func (d *configurationDocumentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Converts a YAML configuration document to JSON for use with the Dagster+ API.",
		Attributes: map[string]schema.Attribute{
			"yaml_body": schema.StringAttribute{
				Description: "The configuration document as a YAML string.",
				Required:    true,
			},
			"json": schema.StringAttribute{
				Description: "The configuration document as a JSON string.",
				Computed:    true,
			},
		},
	}
}

func (d *configurationDocumentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *configurationDocumentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data configurationDocumentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jsonBytes, err := yaml.YAMLToJSON([]byte(data.YAMLBody.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to parse YAML", err.Error())
		return
	}

	data.JSON = types.StringValue(string(jsonBytes))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
