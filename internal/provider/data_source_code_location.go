package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &codeLocationDataSource{}

func NewCodeLocationDataSource() datasource.DataSource {
	return &codeLocationDataSource{}
}

type codeLocationDataSource struct {
	client *client.Client
}

type codeLocationDataSourceModel struct {
	ID               types.String    `tfsdk:"id"`
	DeploymentName   types.String    `tfsdk:"deployment"`
	Name             types.String    `tfsdk:"name"`
	Image            types.String    `tfsdk:"image"`
	CodeSource       codeSourceModel `tfsdk:"code_source"`
	WorkingDirectory types.String    `tfsdk:"working_directory"`
	ExecutablePath   types.String    `tfsdk:"executable_path"`
}

func (d *codeLocationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_location"
}

func (d *codeLocationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ code location from a deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form '{deployment}/{name}'.",
				Computed:    true,
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment the code location belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the code location to look up.",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "The Docker image used by the code location.",
				Computed:    true,
			},
			"working_directory": schema.StringAttribute{
				Description: "The working directory inside the container.",
				Computed:    true,
			},
			"executable_path": schema.StringAttribute{
				Description: "Path to the Python executable inside the container.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"code_source": schema.SingleNestedBlock{
				Description: "Describes where Dagster finds the code.",
				Attributes: map[string]schema.Attribute{
					"python_file": schema.StringAttribute{
						Description: "Path to a Python file containing the repository.",
						Computed:    true,
					},
					"package_name": schema.StringAttribute{
						Description: "Python package name containing the repository.",
						Computed:    true,
					},
					"module_name": schema.StringAttribute{
						Description: "Python module name containing the repository.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (d *codeLocationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *codeLocationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config codeLocationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cl, err := d.client.GetCodeLocation(ctx, config.DeploymentName.ValueString(), config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading code location", err.Error())
		return
	}

	state := codeLocationDataSourceModel{
		ID:               types.StringValue(cl.ID),
		DeploymentName:   config.DeploymentName,
		Name:             types.StringValue(cl.Name),
		Image:            types.StringValue(cl.Image),
		WorkingDirectory: types.StringValue(cl.WorkingDirectory),
		ExecutablePath:   types.StringValue(cl.ExecutablePath),
		CodeSource: codeSourceModel{
			PythonFile:  types.StringValue(cl.CodeSource.PythonFile),
			PackageName: types.StringValue(cl.CodeSource.PackageName),
			ModuleName:  types.StringValue(cl.CodeSource.ModuleName),
		},
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
