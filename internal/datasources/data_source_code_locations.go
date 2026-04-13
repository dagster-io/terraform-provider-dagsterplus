package datasources

import (
	"context"
	"fmt"

	"encoding/json"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	resources "github.com/dagster-io/terraform-provider-dagsterplus/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &codeLocationsDataSource{}

func NewCodeLocationsDataSource() datasource.DataSource {
	return &codeLocationsDataSource{}
}

type codeLocationsDataSource struct {
	client *client.Client
}

type codeLocationsDataSourceModel struct {
	ID             types.String                  `tfsdk:"id"`
	DeploymentName types.String                  `tfsdk:"deployment"`
	CodeLocations  []codeLocationDataSourceModel `tfsdk:"code_locations"`
}

func (d *codeLocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_code_locations"
}

func (d *codeLocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Dagster+ code locations in a deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source.",
				Computed:    true,
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment to list code locations for.",
				Required:    true,
			},
			"code_locations": schema.ListNestedAttribute{
				Description: "All code locations in the deployment.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier in the form `{deployment}/{name}`.",
							Computed:    true,
						},
						"deployment": schema.StringAttribute{
							Description: "The deployment this code location belongs to.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The code location name.",
							Computed:    true,
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
						"attribute": schema.StringAttribute{
							Description: "Python attribute containing the Definitions object.",
							Computed:    true,
						},
						"agent_queue": schema.StringAttribute{
							Description: "Agent queue that runs this code location.",
							Computed:    true,
						},
						"container_context": schema.StringAttribute{
							Description: "JSON-encoded container context configuration for this code location. " +
								"Supports keys such as `k8s`, `ecs`, and `docker`, each with platform-specific settings.",
							Computed:   true,
							CustomType: jsontypes.NormalizedType{},
						},
						"code_source": schema.SingleNestedAttribute{
							Description: "Describes where Dagster finds the code.",
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *codeLocationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *codeLocationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config codeLocationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment := config.DeploymentName.ValueString()
	locs, err := d.client.ListCodeLocations(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error reading code locations", err.Error())
		return
	}

	state := codeLocationsDataSourceModel{
		ID:             types.StringValue(deployment),
		DeploymentName: config.DeploymentName,
		CodeLocations:  make([]codeLocationDataSourceModel, len(locs)),
	}
	for i, cl := range locs {
		m := codeLocationDataSourceModel{
			ID:               types.StringValue(cl.ID),
			DeploymentName:   types.StringValue(deployment),
			Name:             types.StringValue(cl.Name),
			Image:            types.StringValue(cl.Image),
			WorkingDirectory: types.StringValue(cl.WorkingDirectory),
			ExecutablePath:   types.StringValue(cl.ExecutablePath),
			AgentQueue:       types.StringValue(cl.AgentQueue),
			Attribute:        types.StringValue(cl.Attribute),
			CodeSource: &resources.CodeSourceModel{
				PythonFile:  resources.CodeSourceStringVal(cl.CodeSource.PythonFile),
				PackageName: resources.CodeSourceStringVal(cl.CodeSource.PackageName),
				ModuleName:  resources.CodeSourceStringVal(cl.CodeSource.ModuleName),
			},
			ContainerContext: jsontypes.NewNormalizedNull(),
		}
		if len(cl.ContainerContext) > 0 {
			ccBytes, err := json.Marshal(cl.ContainerContext)
			if err == nil {
				m.ContainerContext = jsontypes.NewNormalizedValue(string(ccBytes))
			}
		}
		state.CodeLocations[i] = m
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
