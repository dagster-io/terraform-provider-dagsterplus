package provider

import (
	"context"
	"fmt"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &alertPoliciesDataSource{}

func NewAlertPoliciesDataSource() datasource.DataSource {
	return &alertPoliciesDataSource{}
}

type alertPoliciesDataSource struct {
	client *client.Client
}

type alertPoliciesDataSourceModel struct {
	ID            types.String               `tfsdk:"id"`
	Deployment    types.String               `tfsdk:"deployment"`
	AlertPolicies []alertPolicyResourceModel `tfsdk:"alert_policies"`
}

func (d *alertPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_policies"
}

func (d *alertPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Dagster+ alert policies in a deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for this data source.",
				Computed:    true,
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment to list alert policies for.",
				Required:    true,
			},
			"alert_policies": schema.ListNestedAttribute{
				Description: "All alert policies in the deployment.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier in the form {deployment}/{name}.",
							Computed:    true,
						},
						"deployment": schema.StringAttribute{
							Description: "The deployment this alert policy belongs to.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The alert policy name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A human-readable description of the alert policy.",
							Computed:    true,
						},
						"policy_type": schema.StringAttribute{
							Description: "The category of alert policy: asset, run, code_location, automation, budget, or insight_metric.",
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
						"asset": schema.ListNestedAttribute{
							Description: "Asset-specific configuration (populated when policy_type = asset).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
						"run": schema.ListNestedAttribute{
							Description: "Run-specific configuration (populated when policy_type = run).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
						"code_location": schema.ListNestedAttribute{
							Description: "Code location-specific configuration (populated when policy_type = code_location).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
						"automation": schema.ListNestedAttribute{
							Description: "Automation-specific configuration (populated when policy_type = automation).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
						"budget": schema.ListNestedAttribute{
							Description: "Budget-specific configuration (populated when policy_type = budget).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
						"insight_metric": schema.ListNestedAttribute{
							Description: "Insight metric-specific configuration (populated when policy_type = insight_metric).",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
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
						"notification_service": schema.SingleNestedAttribute{
							Description: "The notification channel configured for this alert policy.",
							Computed:    true,
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
									Description: "Microsoft Teams incoming webhook URL.",
									Computed:    true,
									Sensitive:   true,
								},
								"integration_key": schema.StringAttribute{
									Description: "PagerDuty integration key.",
									Computed:    true,
									Sensitive:   true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *alertPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *alertPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config alertPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment := config.Deployment.ValueString()
	policies, err := d.client.ListAlertPolicies(ctx, deployment)
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert policies", err.Error())
		return
	}

	state := alertPoliciesDataSourceModel{
		ID:            types.StringValue(deployment),
		Deployment:    config.Deployment,
		AlertPolicies: make([]alertPolicyResourceModel, len(policies)),
	}
	for i := range policies {
		state.AlertPolicies[i] = alertPolicyResourceModel{
			Deployment: types.StringValue(deployment),
			PolicyType: types.StringValue(policies[i].PolicyType),
		}
		resp.Diagnostics.Append(policyToModel(&policies[i], &state.AlertPolicies[i])...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
