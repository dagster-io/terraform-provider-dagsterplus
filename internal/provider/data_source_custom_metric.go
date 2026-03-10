package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &customMetricDataSource{}

func NewCustomMetricDataSource() datasource.DataSource {
	return &customMetricDataSource{}
}

type customMetricDataSource struct {
	client *client.Client
}

type customMetricDataSourceModel struct {
	ID              types.String  `tfsdk:"id"`
	MetadataKey     types.String  `tfsdk:"metadata_key"`
	DisplayName     types.String  `tfsdk:"display_name"`
	Description     types.String  `tfsdk:"description"`
	CreateTimestamp types.Float64 `tfsdk:"create_timestamp"`
	UpdateTimestamp types.Float64 `tfsdk:"update_timestamp"`
}

func (d *customMetricDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_metric"
}

func (d *customMetricDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ custom metric by metadata key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The custom metric ID assigned by Dagster+.",
				Computed:    true,
			},
			"metadata_key": schema.StringAttribute{
				Description: "The metadata key of the custom metric to look up.",
				Required:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The human-readable display name for the metric.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the metric.",
				Computed:    true,
			},
			"create_timestamp": schema.Float64Attribute{
				Description: "Unix timestamp of when the custom metric was created.",
				Computed:    true,
			},
			"update_timestamp": schema.Float64Attribute{
				Description: "Unix timestamp of when the custom metric was last updated.",
				Computed:    true,
			},
		},
	}
}

func (d *customMetricDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *customMetricDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config customMetricDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metrics, err := d.client.ListCustomMetrics(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom metrics", err.Error())
		return
	}

	for _, m := range metrics {
		if m.MetadataKey == config.MetadataKey.ValueString() {
			state := customMetricDataSourceModel{
				ID:              types.StringValue(m.ID),
				MetadataKey:     types.StringValue(m.MetadataKey),
				DisplayName:     types.StringValue(m.DisplayName),
				Description:     types.StringValue(m.Description),
				CreateTimestamp: types.Float64Value(m.CreateTimestamp),
				UpdateTimestamp: types.Float64Value(m.UpdateTimestamp),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Custom metric not found",
		fmt.Sprintf("No custom metric with metadata_key %q was found.", config.MetadataKey.ValueString()),
	)
}
