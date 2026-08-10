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

var _ datasource.DataSource = &alertPolicyDataSource{}

// alertPolicyDataSourceModel mirrors resources.AlertPolicyResourceModel but uses a pointer for
// notification_service so the framework can decode a null block when the attribute is
// absent from data source config.
type alertPolicyDataSourceModel struct {
	ID                  types.String                         `tfsdk:"id"`
	Deployment          types.String                         `tfsdk:"deployment"`
	Name                types.String                         `tfsdk:"name"`
	Description         types.String                         `tfsdk:"description"`
	PolicyType          types.String                         `tfsdk:"policy_type"`
	Enabled             types.Bool                           `tfsdk:"enabled"`
	EventTypes          types.List                           `tfsdk:"event_types"`
	Asset               []resources.AssetConfigModel         `tfsdk:"asset"`
	Run                 []resources.RunConfigModel           `tfsdk:"run"`
	CodeLocation        []resources.CodeLocationConfigModel  `tfsdk:"code_location"`
	Automation          []resources.AutomationConfigModel    `tfsdk:"automation"`
	AgentDowntime       []resources.AgentDowntimeConfigModel `tfsdk:"agent_downtime"`
	Budget              []resources.BudgetConfigModel        `tfsdk:"budget"`
	InsightMetric       []resources.InsightMetricConfigModel `tfsdk:"insight_metric"`
	NotificationService *resources.NotificationServiceModel  `tfsdk:"notification_service"`
}

func NewAlertPolicyDataSource() datasource.DataSource {
	return &alertPolicyDataSource{}
}

type alertPolicyDataSource struct {
	client *client.Client
}

func (d *alertPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_policy"
}

func (d *alertPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an existing Dagster+ alert policy from a deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form `{deployment}/{name}`.",
				Computed:    true,
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment the alert policy belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the alert policy to look up.",
				Required:    true,
			},
			"policy_type": schema.StringAttribute{
				Description: "The category of alert policy: asset, run, code_location, automation, agent_downtime, budget, or insight_metric.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A human-readable description of the alert policy.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert policy is active.",
				Computed:    true,
			},
			"event_types": schema.ListAttribute{
				Description: "Event types that trigger the alert.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"asset": schema.ListNestedBlock{
				Description: "Asset-specific configuration (populated when policy_type = asset).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_assets": schema.BoolAttribute{
							Description: "True when the policy applies to all assets.",
							Computed:    true,
						},
						"asset_selection": schema.StringAttribute{
							Description: "Asset selection string (e.g. 'tag:my-tag').",
							Computed:    true,
						},
						"asset_key": schema.StringAttribute{
							Description: "Specific asset key path.",
							Computed:    true,
						},
						"health_status": schema.StringAttribute{
							Description: "Health status transition: degraded, warning, or healthy.",
							Computed:    true,
						},
						"specific_events": schema.ListAttribute{
							Description: "Specific asset events that trigger the alert.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"run": schema.ListNestedBlock{
				Description: "Run-specific configuration (populated when policy_type = run).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_runs": schema.BoolAttribute{
							Description: "True when the policy applies to all runs.",
							Computed:    true,
						},
						"tags": schema.StringAttribute{
							Description: "Run filter tag in 'key=value' format.",
							Computed:    true,
						},
						"job_names": schema.ListAttribute{
							Description: "Job names filter.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"code_locations": schema.ListAttribute{
							Description: "Code location names filter.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"on_success": schema.BoolAttribute{
							Description: "Alert on run success.",
							Computed:    true,
						},
						"on_failure": schema.BoolAttribute{
							Description: "Alert on run failure.",
							Computed:    true,
						},
						"on_timeout_hours": schema.Int64Attribute{
							Description: "Alert when a run exceeds this many hours.",
							Computed:    true,
						},
					},
				},
			},
			"code_location": schema.ListNestedBlock{
				Description: "Code location-specific configuration (populated when policy_type = code_location).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_locations": schema.BoolAttribute{
							Description: "True when the policy applies to all code locations.",
							Computed:    true,
						},
						"location_name": schema.StringAttribute{
							Description: "Specific code location name.",
							Computed:    true,
						},
					},
				},
			},
			"automation": schema.ListNestedBlock{
				Description: "Automation-specific configuration (populated when policy_type = automation).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_schedules_and_sensors": schema.BoolAttribute{
							Description: "True when the policy applies to all schedules and sensors.",
							Computed:    true,
						},
						"code_locations": schema.ListAttribute{
							Description: "Code location names filter.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"schedules_and_sensors": schema.ListAttribute{
							Description: "Specific schedule or sensor names.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"include_schedules": schema.BoolAttribute{
							Description: "Whether schedules are included.",
							Computed:    true,
						},
						"include_sensors": schema.BoolAttribute{
							Description: "Whether sensors are included.",
							Computed:    true,
						},
						"min_consecutive_failures": schema.Int64Attribute{
							Description: "Minimum consecutive failures before alerting.",
							Computed:    true,
						},
					},
				},
			},
			"agent_downtime": schema.ListNestedBlock{
				Description: "Agent downtime configuration (populated when policy_type = agent_downtime and a renotify interval is set).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"renotify_interval_minutes": schema.Int64Attribute{
							Description: "Minutes between repeat notifications while the agent remains unavailable.",
							Computed:    true,
						},
					},
				},
			},
			"budget": schema.ListNestedBlock{
				Description: "Budget-specific configuration (populated when policy_type = budget).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"operator": schema.StringAttribute{
							Description: "Comparison operator: greater_than or less_than.",
							Computed:    true,
						},
						"threshold": schema.Float64Attribute{
							Description: "Budget threshold value.",
							Computed:    true,
						},
						"period_days": schema.Int64Attribute{
							Description: "Period in days over which to measure.",
							Computed:    true,
						},
					},
				},
			},
			"insight_metric": schema.ListNestedBlock{
				Description: "Insight metric-specific configuration (populated when policy_type = insight_metric).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"asset_group": schema.StringAttribute{
							Description: "Asset group name filter.",
							Computed:    true,
						},
						"asset_key": schema.StringAttribute{
							Description: "Asset key filter (slash-delimited path).",
							Computed:    true,
						},
						"job_name": schema.StringAttribute{
							Description: "Job name filter.",
							Computed:    true,
						},
						"metric": schema.StringAttribute{
							Description: "The metric being monitored.",
							Computed:    true,
						},
						"operator": schema.StringAttribute{
							Description: "Comparison operator: greater_than or less_than.",
							Computed:    true,
						},
						"threshold": schema.Float64Attribute{
							Description: "Metric threshold value.",
							Computed:    true,
						},
						"period_days": schema.Int64Attribute{
							Description: "Period in days over which to measure.",
							Computed:    true,
						},
					},
				},
			},
			"notification_service": schema.SingleNestedBlock{
				Description: "The notification channel configured for this alert policy.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Notification type: email, email_owners, slack, microsoft_teams, or pagerduty.",
						Computed:    true,
					},
					"email_addresses": schema.ListAttribute{
						Description: "Email addresses to notify.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"default_email_addresses": schema.ListAttribute{
						Description: "Fallback email addresses for email_owners notifications.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"slack_workspace_name": schema.StringAttribute{
						Description: "Slack workspace name.",
						Computed:    true,
					},
					"slack_channel_name": schema.StringAttribute{
						Description: "Slack channel name.",
						Computed:    true,
					},
					"webhook_url": schema.StringAttribute{
						Description: "Incoming webhook URL (Microsoft Teams or generic webhook).",
						Computed:    true,
						Sensitive:   true,
					},
					"integration_key": schema.StringAttribute{
						Description: "PagerDuty integration key.",
						Computed:    true,
						Sensitive:   true,
					},
					"headers": schema.MapAttribute{
						Description: "Arbitrary HTTP headers sent with each generic webhook request.",
						Computed:    true,
						Sensitive:   true,
						ElementType: types.StringType,
					},
					"body_template": schema.StringAttribute{
						Description: "Templated request body sent with each generic webhook request.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (d *alertPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config alertPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetAlertPolicy(ctx, config.Deployment.ValueString(), config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert policy", err.Error())
		return
	}

	// Convert to resource model to reuse resources.PolicyToModel, then copy into data source model.
	rm := resources.AlertPolicyResourceModel{
		Deployment: config.Deployment,
		PolicyType: config.PolicyType,
	}
	resp.Diagnostics.Append(resources.PolicyToModel(policy, &rm)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = rm.ID
	config.Name = rm.Name
	config.Description = rm.Description
	config.Enabled = rm.Enabled
	config.EventTypes = rm.EventTypes
	config.Asset = rm.Asset
	config.Run = rm.Run
	config.CodeLocation = rm.CodeLocation
	config.Automation = rm.Automation
	config.AgentDowntime = rm.AgentDowntime
	config.Budget = rm.Budget
	config.InsightMetric = rm.InsightMetric
	config.NotificationService = &rm.NotificationService
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
