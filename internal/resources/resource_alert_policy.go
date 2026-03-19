package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &alertPolicyResource{}
var _ resource.ResourceWithImportState = &alertPolicyResource{}

func NewAlertPolicyResource() resource.Resource {
	return &alertPolicyResource{}
}

type alertPolicyResource struct {
	client *client.Client
}

type AlertPolicyResourceModel struct {
	ID                  types.String               `tfsdk:"id"`
	Deployment          types.String               `tfsdk:"deployment"`
	Name                types.String               `tfsdk:"name"`
	Description         types.String               `tfsdk:"description"`
	PolicyType          types.String               `tfsdk:"policy_type"`
	Enabled             types.Bool                 `tfsdk:"enabled"`
	EventTypes          types.List                 `tfsdk:"event_types"`
	Asset               []AssetConfigModel         `tfsdk:"asset"`
	Run                 []RunConfigModel           `tfsdk:"run"`
	CodeLocation        []CodeLocationConfigModel  `tfsdk:"code_location"`
	Automation          []AutomationConfigModel    `tfsdk:"automation"`
	Budget              []BudgetConfigModel        `tfsdk:"budget"`
	InsightMetric       []InsightMetricConfigModel `tfsdk:"insight_metric"`
	NotificationService NotificationServiceModel   `tfsdk:"notification_service"`
}

type AssetConfigModel struct {
	AllAssets      types.Bool   `tfsdk:"all_assets"`
	AssetSelection types.String `tfsdk:"asset_selection"`
	AssetKey       types.String `tfsdk:"asset_key"`
	HealthStatus   types.String `tfsdk:"health_status"`
	SpecificEvents types.List   `tfsdk:"specific_events"`
}

type RunConfigModel struct {
	AllRuns        types.Bool   `tfsdk:"all_runs"`
	Tags           types.String `tfsdk:"tags"`
	JobNames       types.List   `tfsdk:"job_names"`
	CodeLocations  types.List   `tfsdk:"code_locations"`
	OnSuccess      types.Bool   `tfsdk:"on_success"`
	OnFailure      types.Bool   `tfsdk:"on_failure"`
	OnTimeoutHours types.Int64  `tfsdk:"on_timeout_hours"`
}

type CodeLocationConfigModel struct {
	AllLocations types.Bool   `tfsdk:"all_locations"`
	LocationName types.String `tfsdk:"location_name"`
}

type AutomationConfigModel struct {
	AllSchedulesAndSensors types.Bool  `tfsdk:"all_schedules_and_sensors"`
	CodeLocations          types.List  `tfsdk:"code_locations"`
	SchedulesAndSensors    types.List  `tfsdk:"schedules_and_sensors"`
	IncludeSchedules       types.Bool  `tfsdk:"include_schedules"`
	IncludeSensors         types.Bool  `tfsdk:"include_sensors"`
	MinConsecutiveFailures types.Int64 `tfsdk:"min_consecutive_failures"`
}

type BudgetConfigModel struct {
	Operator   types.String  `tfsdk:"operator"`
	Threshold  types.Float64 `tfsdk:"threshold"`
	PeriodDays types.Int64   `tfsdk:"period_days"`
}

type InsightMetricConfigModel struct {
	AssetGroup types.String  `tfsdk:"asset_group"`
	AssetKey   types.String  `tfsdk:"asset_key"`
	JobName    types.String  `tfsdk:"job_name"`
	Metric     types.String  `tfsdk:"metric"`
	Operator   types.String  `tfsdk:"operator"`
	Threshold  types.Float64 `tfsdk:"threshold"`
	PeriodDays types.Int64   `tfsdk:"period_days"`
}

// specificEventToAPI maps user-friendly event names to Dagster+ API event type strings.
var specificEventToAPI = map[string]string{
	"materialization_success": "ASSET_MATERIALIZATION_SUCCESS",
	"materialization_failure": "ASSET_MATERIALIZATION_FAILURE",
	"check_passed":            "ASSET_CHECK_PASSED",
	"check_warn":              "ASSET_CHECK_SEVERITY_WARN",
	"check_error":             "ASSET_CHECK_SEVERITY_ERROR",
	"check_execution_failed":  "ASSET_CHECK_EXECUTION_FAILED",
	"freshness_passing":       "ASSET_NOT_OVERDUE",
	"freshness_warning":       "ASSET_OVERDUE_WARN",
	"freshness_failure":       "ASSET_OVERDUE",
}

// apiToSpecificEvent is the reverse of specificEventToAPI.
var apiToSpecificEvent = func() map[string]string {
	m := make(map[string]string, len(specificEventToAPI))
	for k, v := range specificEventToAPI {
		m[v] = k
	}
	return m
}()

type NotificationServiceModel struct {
	Type                  types.String `tfsdk:"type"`
	EmailAddresses        types.List   `tfsdk:"email_addresses"`
	DefaultEmailAddresses types.List   `tfsdk:"default_email_addresses"`
	SlackWorkspaceName    types.String `tfsdk:"slack_workspace_name"`
	SlackChannelName      types.String `tfsdk:"slack_channel_name"`
	WebhookURL            types.String `tfsdk:"webhook_url"`
	IntegrationKey        types.String `tfsdk:"integration_key"`
}

func (r *alertPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert_policy"
}

func (r *alertPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dagster+ alert policy for a specific deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier in the form {deployment}/{name}.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deployment": schema.StringAttribute{
				Description: "The name of the deployment this alert policy belongs to (e.g. 'prod').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The unique name of the alert policy. Changing this forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "A human-readable description of the alert policy.",
				Optional:    true,
				Computed:    true,
			},
			"policy_type": schema.StringAttribute{
				Description: "The category of alert policy: asset, run, code_location, automation, budget, or insight_metric.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("asset", "run", "code_location", "automation", "budget", "insight_metric"),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert policy is active.",
				Required:    true,
			},
			"event_types": schema.ListAttribute{
				Description: "Event types that trigger the alert. Read-only — always derived by the provider from the policy-type-specific block.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"asset": schema.ListNestedBlock{
				Description: "Asset-specific configuration. Use when policy_type = asset. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_assets": schema.BoolAttribute{
							Description: "When true, the policy applies to all assets. Mutually exclusive with asset_selection and asset_key.",
							Optional:    true,
							Validators: []validator.Bool{
								boolvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("asset_selection"),
									path.MatchRelative().AtParent().AtName("asset_key"),
								),
							},
						},
						"asset_selection": schema.StringAttribute{
							Description: "An asset selection string (e.g. 'tag:my-tag'). Mutually exclusive with all_assets and asset_key.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_assets"),
									path.MatchRelative().AtParent().AtName("asset_key"),
								),
							},
						},
						"asset_key": schema.StringAttribute{
							Description: "A specific asset key path (e.g. 'my_asset' or 'group/my_asset'). Mutually exclusive with all_assets and asset_selection.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_assets"),
									path.MatchRelative().AtParent().AtName("asset_selection"),
								),
							},
						},
						"health_status": schema.StringAttribute{
							Description: "Trigger on a health status transition: degraded, warning, or healthy. Mutually exclusive with specific_events.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("specific_events"),
								),
							},
						},
						"specific_events": schema.ListAttribute{
							Description: "Trigger on specific asset events. Mutually exclusive with health_status. " +
								"Valid values — materializations: materialization_success, materialization_failure; " +
								"asset checks: check_passed, check_warn, check_error, check_execution_failed; " +
								"freshness: freshness_passing, freshness_warning, freshness_failure.",
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("health_status"),
								),
							},
						},
					},
				},
			},
			"run": schema.ListNestedBlock{
				Description: "Run-specific configuration. Use when policy_type = run. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_runs": schema.BoolAttribute{
							Description: "When true, the policy applies to all runs.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Bool{
								boolvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("tags"),
									path.MatchRelative().AtParent().AtName("job_names"),
									path.MatchRelative().AtParent().AtName("code_locations"),
								),
							},
						},
						"tags": schema.StringAttribute{
							Description: "Filter runs by tag in 'key=value' format.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_runs"),
								),
							},
						},
						"job_names": schema.ListAttribute{
							Description: "Filter runs by job names.",
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_runs"),
								),
							},
						},
						"code_locations": schema.ListAttribute{
							Description: "Filter runs by code location names.",
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_runs"),
								),
							},
						},
						"on_success": schema.BoolAttribute{
							Description: "Alert on run success.",
							Optional:    true,
							Computed:    true,
						},
						"on_failure": schema.BoolAttribute{
							Description: "Alert on run failure.",
							Optional:    true,
							Computed:    true,
						},
						"on_timeout_hours": schema.Int64Attribute{
							Description: "Alert when a run exceeds this many hours.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
					},
				},
			},
			"code_location": schema.ListNestedBlock{
				Description: "Code location-specific configuration. Use when policy_type = code_location. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_locations": schema.BoolAttribute{
							Description: "When true, the policy applies to all code locations.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Bool{
								boolvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("location_name"),
								),
							},
						},
						"location_name": schema.StringAttribute{
							Description: "A specific code location name to monitor.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_locations"),
								),
							},
						},
					},
				},
			},
			"automation": schema.ListNestedBlock{
				Description: "Automation-specific configuration. Use when policy_type = automation. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"all_schedules_and_sensors": schema.BoolAttribute{
							Description: "When true, the policy applies to all schedules and sensors.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Bool{
								boolvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("code_locations"),
									path.MatchRelative().AtParent().AtName("schedules_and_sensors"),
								),
							},
						},
						"code_locations": schema.ListAttribute{
							Description: "Filter by code location names.",
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_schedules_and_sensors"),
								),
							},
						},
						"schedules_and_sensors": schema.ListAttribute{
							Description: "Specific schedule or sensor names to monitor.",
							Optional:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("all_schedules_and_sensors"),
								),
							},
						},
						"include_schedules": schema.BoolAttribute{
							Description: "Include schedules in monitoring.",
							Optional:    true,
							Computed:    true,
						},
						"include_sensors": schema.BoolAttribute{
							Description: "Include sensors in monitoring.",
							Optional:    true,
							Computed:    true,
						},
						"min_consecutive_failures": schema.Int64Attribute{
							Description: "Minimum number of consecutive failures before alerting.",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 100),
							},
						},
					},
				},
			},
			"budget": schema.ListNestedBlock{
				Description: "Budget-specific configuration. Use when policy_type = budget. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"operator": schema.StringAttribute{
							Description: "Comparison operator: greater_than or less_than.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("greater_than", "less_than"),
							},
						},
						"threshold": schema.Float64Attribute{
							Description: "The budget threshold value.",
							Required:    true,
						},
						"period_days": schema.Int64Attribute{
							Description: "The period in days over which to measure.",
							Required:    true,
						},
					},
				},
			},
			"insight_metric": schema.ListNestedBlock{
				Description: "Insight metric-specific configuration. Use when policy_type = insight_metric. At most one block is allowed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"asset_group": schema.StringAttribute{
							Description: "Filter by asset group name.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("asset_key"),
									path.MatchRelative().AtParent().AtName("job_name"),
								),
							},
						},
						"asset_key": schema.StringAttribute{
							Description: "Filter by asset key (slash-delimited path).",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("asset_group"),
									path.MatchRelative().AtParent().AtName("job_name"),
								),
							},
						},
						"job_name": schema.StringAttribute{
							Description: "Filter by job name.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("asset_group"),
									path.MatchRelative().AtParent().AtName("asset_key"),
								),
							},
						},
						"metric": schema.StringAttribute{
							Description: "The metric to monitor (e.g. dagster_credits).",
							Required:    true,
						},
						"operator": schema.StringAttribute{
							Description: "Comparison operator: greater_than or less_than.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("greater_than", "less_than"),
							},
						},
						"threshold": schema.Float64Attribute{
							Description: "The metric threshold value.",
							Required:    true,
						},
						"period_days": schema.Int64Attribute{
							Description: "The period in days over which to measure.",
							Required:    true,
						},
					},
				},
			},
			"notification_service": schema.SingleNestedBlock{
				Description: "The notification channel for this alert policy. Exactly one notification type should be configured.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Notification type: email, email_owners, slack, microsoft_teams, or pagerduty.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("email", "email_owners", "slack", "microsoft_teams", "pagerduty"),
						},
					},
					"email_addresses": schema.ListAttribute{
						Description: "Email addresses to notify. Required when type = email.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"default_email_addresses": schema.ListAttribute{
						Description: "Fallback email addresses when no code owners are found. Used when type = email_owners.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"slack_workspace_name": schema.StringAttribute{
						Description: "Slack workspace name. Required when type = slack.",
						Optional:    true,
					},
					"slack_channel_name": schema.StringAttribute{
						Description: "Slack channel name (e.g. '#alerts'). Required when type = slack.",
						Optional:    true,
					},
					"webhook_url": schema.StringAttribute{
						Description: "Microsoft Teams incoming webhook URL. Required when type = microsoft_teams.",
						Optional:    true,
						Sensitive:   true,
					},
					"integration_key": schema.StringAttribute{
						Description: "PagerDuty integration key. Required when type = pagerduty.",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
		},
	}
}

func (r *alertPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *alertPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, diags := modelToPolicy(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOrUpdateAlertPolicy(ctx, plan.Deployment.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert policy", err.Error())
		return
	}

	resp.Diagnostics.Append(PolicyToModel(created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetAlertPolicy(ctx, state.Deployment.ValueString(), state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading alert policy", err.Error())
		return
	}

	resp.Diagnostics.Append(PolicyToModel(policy, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alertPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, diags := modelToPolicy(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.CreateOrUpdateAlertPolicy(ctx, plan.Deployment.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert policy", err.Error())
		return
	}

	resp.Diagnostics.Append(PolicyToModel(updated, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAlertPolicy(ctx, state.Deployment.ValueString(), state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting alert policy", err.Error())
	}
}

func (r *alertPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: deployment/policy-name")
		return
	}

	policy, err := r.client.GetAlertPolicy(ctx, parts[0], parts[1])
	if err != nil {
		resp.Diagnostics.AddError("Error importing alert policy", err.Error())
		return
	}

	var state AlertPolicyResourceModel
	state.Deployment = types.StringValue(parts[0])
	resp.Diagnostics.Append(PolicyToModel(policy, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// modelToPolicy converts the Terraform model to the client AlertPolicy type.
func modelToPolicy(ctx context.Context, model AlertPolicyResourceModel) (client.AlertPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	policy := client.AlertPolicy{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		PolicyType:  model.PolicyType.ValueString(),
		Enabled:     model.Enabled.ValueBool(),
	}

	if len(model.Asset) > 0 {
		a := model.Asset[0]

		switch {
		case !a.AllAssets.IsNull() && a.AllAssets.ValueBool():
			policy.AlertTargetType = "all_assets"
		case !a.AssetSelection.IsNull() && a.AssetSelection.ValueString() != "":
			policy.AlertTargetType = "asset_selection"
			policy.AlertTargetValue = a.AssetSelection.ValueString()
		case !a.AssetKey.IsNull() && a.AssetKey.ValueString() != "":
			policy.AlertTargetType = "asset_key"
			policy.AlertTargetValue = a.AssetKey.ValueString()
		}

		if !a.HealthStatus.IsNull() && a.HealthStatus.ValueString() != "" {
			switch a.HealthStatus.ValueString() {
			case "degraded":
				policy.EventTypes = append(policy.EventTypes, "ASSET_HEALTH_DEGRADED")
			case "warning":
				policy.EventTypes = append(policy.EventTypes, "ASSET_HEALTH_WARNING")
			case "healthy":
				policy.EventTypes = append(policy.EventTypes, "ASSET_HEALTH_HEALTHY")
			}
		} else if !a.SpecificEvents.IsNull() && !a.SpecificEvents.IsUnknown() {
			var events []string
			diags.Append(a.SpecificEvents.ElementsAs(ctx, &events, false)...)
			for _, e := range events {
				if apiEvent, ok := specificEventToAPI[e]; ok {
					policy.EventTypes = append(policy.EventTypes, apiEvent)
				}
			}
		}

	} else if len(model.Run) > 0 {
		r := model.Run[0]
		policy.Run = &client.RunAlertConfig{
			AllRuns:        r.AllRuns.ValueBool(),
			OnSuccess:      r.OnSuccess.ValueBool(),
			OnFailure:      r.OnFailure.ValueBool(),
			OnTimeoutHours: int(r.OnTimeoutHours.ValueInt64()),
		}
		if !r.Tags.IsNull() && r.Tags.ValueString() != "" {
			policy.Run.Filter.Tags = r.Tags.ValueString()
		}
		if !r.JobNames.IsNull() && !r.JobNames.IsUnknown() {
			diags.Append(r.JobNames.ElementsAs(ctx, &policy.Run.Filter.JobNames, false)...)
		}
		if !r.CodeLocations.IsNull() && !r.CodeLocations.IsUnknown() {
			diags.Append(r.CodeLocations.ElementsAs(ctx, &policy.Run.Filter.LocationNames, false)...)
		}
		if policy.Run.OnSuccess {
			policy.EventTypes = append(policy.EventTypes, "JOB_SUCCESS")
		}
		if policy.Run.OnFailure {
			policy.EventTypes = append(policy.EventTypes, "JOB_FAILURE")
		}

	} else if len(model.CodeLocation) > 0 {
		c := model.CodeLocation[0]
		policy.CodeLocation = &client.CodeLocationAlertConfig{
			AllLocations: c.AllLocations.ValueBool(),
			LocationName: c.LocationName.ValueString(),
		}
		policy.EventTypes = []string{"CODE_LOCATION_ERROR"}

	} else if len(model.Automation) > 0 {
		a := model.Automation[0]
		policy.Automation = &client.AutomationAlertConfig{
			AllSchedulesAndSensors: a.AllSchedulesAndSensors.ValueBool(),
			IncludeSchedules:       a.IncludeSchedules.ValueBool(),
			IncludeSensors:         a.IncludeSensors.ValueBool(),
			MinConsecutiveFailures: int(a.MinConsecutiveFailures.ValueInt64()),
		}
		if !a.CodeLocations.IsNull() && !a.CodeLocations.IsUnknown() {
			diags.Append(a.CodeLocations.ElementsAs(ctx, &policy.Automation.LocationNames, false)...)
		}
		if !a.SchedulesAndSensors.IsNull() && !a.SchedulesAndSensors.IsUnknown() {
			diags.Append(a.SchedulesAndSensors.ElementsAs(ctx, &policy.Automation.SchedulesAndSensors, false)...)
		}
		policy.EventTypes = []string{"TICK_FAILURE"}

	} else if len(model.Budget) > 0 {
		b := model.Budget[0]
		policy.Budget = &client.BudgetAlertConfig{
			Operator:   b.Operator.ValueString(),
			Threshold:  b.Threshold.ValueFloat64(),
			PeriodDays: int(b.PeriodDays.ValueInt64()),
		}
		policy.EventTypes = []string{"INSIGHTS_CONSUMPTION_EXCEEDED"}

	} else if len(model.InsightMetric) > 0 {
		im := model.InsightMetric[0]
		policy.InsightMetric = &client.InsightMetricAlertConfig{
			AssetGroup: im.AssetGroup.ValueString(),
			AssetKey:   im.AssetKey.ValueString(),
			JobName:    im.JobName.ValueString(),
			Metric:     im.Metric.ValueString(),
			Operator:   im.Operator.ValueString(),
			Threshold:  im.Threshold.ValueFloat64(),
			PeriodDays: int(im.PeriodDays.ValueInt64()),
		}
		policy.EventTypes = []string{"INSIGHTS_CONSUMPTION_EXCEEDED"}

	}

	// Notification service.
	ns := model.NotificationService
	policy.NotificationService = client.AlertPolicyNotification{Type: ns.Type.ValueString()}
	switch ns.Type.ValueString() {
	case "email":
		diags.Append(ns.EmailAddresses.ElementsAs(ctx, &policy.NotificationService.EmailAddresses, false)...)
	case "email_owners":
		if !ns.DefaultEmailAddresses.IsNull() && !ns.DefaultEmailAddresses.IsUnknown() {
			diags.Append(ns.DefaultEmailAddresses.ElementsAs(ctx, &policy.NotificationService.DefaultEmailAddresses, false)...)
		}
	case "slack":
		policy.NotificationService.SlackWorkspaceName = ns.SlackWorkspaceName.ValueString()
		policy.NotificationService.SlackChannelName = ns.SlackChannelName.ValueString()
	case "microsoft_teams":
		policy.NotificationService.WebhookURL = ns.WebhookURL.ValueString()
	case "pagerduty":
		policy.NotificationService.IntegrationKey = ns.IntegrationKey.ValueString()
	}

	return policy, diags
}

// PolicyToModel populates the Terraform model from the client AlertPolicy.
func PolicyToModel(policy *client.AlertPolicy, model *AlertPolicyResourceModel) diag.Diagnostics {
	model.ID = types.StringValue(fmt.Sprintf("%s/%s", model.Deployment.ValueString(), policy.Name))
	model.Name = types.StringValue(policy.Name)
	model.Description = types.StringValue(policy.Description)
	// policy_type is not returned by the API; preserve the existing value from config/state.
	model.Enabled = types.BoolValue(policy.Enabled)

	// Populate event_types from API response.
	eventElems := make([]attr.Value, len(policy.EventTypes))
	for i, e := range policy.EventTypes {
		eventElems[i] = types.StringValue(e)
	}
	model.EventTypes = types.ListValueMust(types.StringType, eventElems)

	// policy_type is not returned by the API. When it is absent (e.g. during
	// import), infer it from whichever config block the API populated.
	policyType := model.PolicyType.ValueString()
	if policyType == "" {
		switch {
		case policy.Run != nil:
			policyType = "run"
		case policy.CodeLocation != nil:
			policyType = "code_location"
		case policy.Automation != nil:
			policyType = "automation"
		case policy.Budget != nil:
			policyType = "budget"
		case policy.InsightMetric != nil:
			policyType = "insight_metric"
		default:
			policyType = "asset"
		}
		model.PolicyType = types.StringValue(policyType)
	}

	switch policyType {
	case "asset":
		a := AssetConfigModel{
			AllAssets:      types.BoolNull(),
			AssetSelection: types.StringNull(),
			AssetKey:       types.StringNull(),
			HealthStatus:   types.StringNull(),
			SpecificEvents: types.ListNull(types.StringType),
		}

		switch policy.AlertTargetType {
		case "all_assets":
			a.AllAssets = types.BoolValue(true)
		case "asset_selection":
			a.AssetSelection = types.StringValue(policy.AlertTargetValue)
		case "asset_key":
			a.AssetKey = types.StringValue(policy.AlertTargetValue)
		}

		healthStatusMap := map[string]string{
			"ASSET_HEALTH_DEGRADED": "degraded",
			"ASSET_HEALTH_WARNING":  "warning",
			"ASSET_HEALTH_HEALTHY":  "healthy",
		}
		var specificElems []attr.Value
		for _, et := range policy.EventTypes {
			if hs, ok := healthStatusMap[et]; ok {
				a.HealthStatus = types.StringValue(hs)
			} else if friendly, ok := apiToSpecificEvent[et]; ok {
				specificElems = append(specificElems, types.StringValue(friendly))
			}
		}
		if len(specificElems) > 0 {
			a.SpecificEvents = types.ListValueMust(types.StringType, specificElems)
		}

		model.Asset = []AssetConfigModel{a}
		model.Run = []RunConfigModel{}
		model.CodeLocation = []CodeLocationConfigModel{}
		model.Automation = []AutomationConfigModel{}
		model.Budget = []BudgetConfigModel{}
		model.InsightMetric = []InsightMetricConfigModel{}

	case "run":
		model.Asset = []AssetConfigModel{}
		if policy.Run != nil {
			model.Run = []RunConfigModel{runToModel(policy.Run)}
		} else {
			model.Run = []RunConfigModel{}
		}
		model.CodeLocation = []CodeLocationConfigModel{}
		model.Automation = []AutomationConfigModel{}
		model.Budget = []BudgetConfigModel{}
		model.InsightMetric = []InsightMetricConfigModel{}

	case "code_location":
		model.Asset = []AssetConfigModel{}
		model.Run = []RunConfigModel{}
		if policy.CodeLocation != nil {
			model.CodeLocation = []CodeLocationConfigModel{codeLocationToModel(policy.CodeLocation)}
		} else {
			model.CodeLocation = []CodeLocationConfigModel{}
		}
		model.Automation = []AutomationConfigModel{}
		model.Budget = []BudgetConfigModel{}
		model.InsightMetric = []InsightMetricConfigModel{}

	case "automation":
		model.Asset = []AssetConfigModel{}
		model.Run = []RunConfigModel{}
		model.CodeLocation = []CodeLocationConfigModel{}
		if policy.Automation != nil {
			model.Automation = []AutomationConfigModel{automationToModel(policy.Automation)}
		} else {
			model.Automation = []AutomationConfigModel{}
		}
		model.Budget = []BudgetConfigModel{}
		model.InsightMetric = []InsightMetricConfigModel{}

	case "budget":
		model.Asset = []AssetConfigModel{}
		model.Run = []RunConfigModel{}
		model.CodeLocation = []CodeLocationConfigModel{}
		model.Automation = []AutomationConfigModel{}
		if policy.Budget != nil {
			model.Budget = []BudgetConfigModel{budgetToModel(policy.Budget)}
		} else {
			model.Budget = []BudgetConfigModel{}
		}
		model.InsightMetric = []InsightMetricConfigModel{}

	case "insight_metric":
		model.Asset = []AssetConfigModel{}
		model.Run = []RunConfigModel{}
		model.CodeLocation = []CodeLocationConfigModel{}
		model.Automation = []AutomationConfigModel{}
		model.Budget = []BudgetConfigModel{}
		if policy.InsightMetric != nil {
			model.InsightMetric = []InsightMetricConfigModel{insightMetricToModel(policy.InsightMetric)}
		} else {
			model.InsightMetric = []InsightMetricConfigModel{}
		}

	default:
		model.Asset = []AssetConfigModel{}
		model.Run = []RunConfigModel{}
		model.CodeLocation = []CodeLocationConfigModel{}
		model.Automation = []AutomationConfigModel{}
		model.Budget = []BudgetConfigModel{}
		model.InsightMetric = []InsightMetricConfigModel{}
	}

	// Notification service.
	ns := policy.NotificationService
	model.NotificationService = NotificationServiceModel{
		Type:                  types.StringValue(ns.Type),
		EmailAddresses:        types.ListNull(types.StringType),
		DefaultEmailAddresses: types.ListNull(types.StringType),
		SlackWorkspaceName:    types.StringNull(),
		SlackChannelName:      types.StringNull(),
		WebhookURL:            types.StringNull(),
		IntegrationKey:        types.StringNull(),
	}
	switch ns.Type {
	case "email":
		elems := make([]attr.Value, len(ns.EmailAddresses))
		for i, a := range ns.EmailAddresses {
			elems[i] = types.StringValue(a)
		}
		model.NotificationService.EmailAddresses = types.ListValueMust(types.StringType, elems)
	case "email_owners":
		elems := make([]attr.Value, len(ns.DefaultEmailAddresses))
		for i, a := range ns.DefaultEmailAddresses {
			elems[i] = types.StringValue(a)
		}
		model.NotificationService.DefaultEmailAddresses = types.ListValueMust(types.StringType, elems)
	case "slack":
		model.NotificationService.SlackWorkspaceName = types.StringValue(ns.SlackWorkspaceName)
		model.NotificationService.SlackChannelName = types.StringValue(ns.SlackChannelName)
	case "microsoft_teams":
		model.NotificationService.WebhookURL = types.StringValue(ns.WebhookURL)
	case "pagerduty":
		model.NotificationService.IntegrationKey = types.StringValue(ns.IntegrationKey)
	}

	return nil
}

func runToModel(r *client.RunAlertConfig) RunConfigModel {
	m := RunConfigModel{
		AllRuns:        types.BoolValue(r.AllRuns),
		Tags:           types.StringNull(),
		JobNames:       types.ListNull(types.StringType),
		CodeLocations:  types.ListNull(types.StringType),
		OnSuccess:      types.BoolValue(r.OnSuccess),
		OnFailure:      types.BoolValue(r.OnFailure),
		OnTimeoutHours: types.Int64Null(),
	}
	if r.OnTimeoutHours > 0 {
		m.OnTimeoutHours = types.Int64Value(int64(r.OnTimeoutHours))
	}
	if r.Filter.Tags != "" {
		m.Tags = types.StringValue(r.Filter.Tags)
	}
	if len(r.Filter.JobNames) > 0 {
		elems := make([]attr.Value, len(r.Filter.JobNames))
		for i, j := range r.Filter.JobNames {
			elems[i] = types.StringValue(j)
		}
		m.JobNames = types.ListValueMust(types.StringType, elems)
	}
	if len(r.Filter.LocationNames) > 0 {
		elems := make([]attr.Value, len(r.Filter.LocationNames))
		for i, l := range r.Filter.LocationNames {
			elems[i] = types.StringValue(l)
		}
		m.CodeLocations = types.ListValueMust(types.StringType, elems)
	}
	return m
}

func codeLocationToModel(c *client.CodeLocationAlertConfig) CodeLocationConfigModel {
	locationName := types.StringNull()
	if c.LocationName != "" {
		locationName = types.StringValue(c.LocationName)
	}
	return CodeLocationConfigModel{
		AllLocations: types.BoolValue(c.AllLocations),
		LocationName: locationName,
	}
}

func automationToModel(a *client.AutomationAlertConfig) AutomationConfigModel {
	m := AutomationConfigModel{
		AllSchedulesAndSensors: types.BoolValue(a.AllSchedulesAndSensors),
		IncludeSchedules:       types.BoolValue(a.IncludeSchedules),
		IncludeSensors:         types.BoolValue(a.IncludeSensors),
		MinConsecutiveFailures: types.Int64Null(),
		CodeLocations:          types.ListNull(types.StringType),
		SchedulesAndSensors:    types.ListNull(types.StringType),
	}
	if a.MinConsecutiveFailures > 0 {
		m.MinConsecutiveFailures = types.Int64Value(int64(a.MinConsecutiveFailures))
	}
	if len(a.LocationNames) > 0 {
		elems := make([]attr.Value, len(a.LocationNames))
		for i, l := range a.LocationNames {
			elems[i] = types.StringValue(l)
		}
		m.CodeLocations = types.ListValueMust(types.StringType, elems)
	}
	if len(a.SchedulesAndSensors) > 0 {
		elems := make([]attr.Value, len(a.SchedulesAndSensors))
		for i, s := range a.SchedulesAndSensors {
			elems[i] = types.StringValue(s)
		}
		m.SchedulesAndSensors = types.ListValueMust(types.StringType, elems)
	}
	return m
}

func budgetToModel(b *client.BudgetAlertConfig) BudgetConfigModel {
	return BudgetConfigModel{
		Operator:   types.StringValue(b.Operator),
		Threshold:  types.Float64Value(b.Threshold),
		PeriodDays: types.Int64Value(int64(b.PeriodDays)),
	}
}

func insightMetricToModel(im *client.InsightMetricAlertConfig) InsightMetricConfigModel {
	assetGroup := types.StringNull()
	if im.AssetGroup != "" {
		assetGroup = types.StringValue(im.AssetGroup)
	}
	assetKey := types.StringNull()
	if im.AssetKey != "" {
		assetKey = types.StringValue(im.AssetKey)
	}
	jobName := types.StringNull()
	if im.JobName != "" {
		jobName = types.StringValue(im.JobName)
	}
	return InsightMetricConfigModel{
		AssetGroup: assetGroup,
		AssetKey:   assetKey,
		JobName:    jobName,
		Metric:     types.StringValue(im.Metric),
		Operator:   types.StringValue(im.Operator),
		Threshold:  types.Float64Value(im.Threshold),
		PeriodDays: types.Int64Value(int64(im.PeriodDays)),
	}
}
