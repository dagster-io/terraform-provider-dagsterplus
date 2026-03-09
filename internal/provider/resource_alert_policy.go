package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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

type alertPolicyResourceModel struct {
	ID                  types.String             `tfsdk:"id"`
	Deployment          types.String             `tfsdk:"deployment"`
	Name                types.String             `tfsdk:"name"`
	Description         types.String             `tfsdk:"description"`
	PolicyType          types.String             `tfsdk:"policy_type"`
	Enabled             types.Bool               `tfsdk:"enabled"`
	EventTypes          types.List               `tfsdk:"event_types"`
	Asset               []assetConfigModel       `tfsdk:"asset"`
	NotificationService notificationServiceModel `tfsdk:"notification_service"`
}

type assetConfigModel struct {
	AllAssets      types.Bool   `tfsdk:"all_assets"`
	AssetSelection types.String `tfsdk:"asset_selection"`
	AssetKey       types.String `tfsdk:"asset_key"`
	HealthStatus   types.String `tfsdk:"health_status"`
}

type notificationServiceModel struct {
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
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the alert policy is active.",
				Required:    true,
			},
			"event_types": schema.ListAttribute{
				Description: "Event types that trigger the alert (e.g. JOB_FAILURE, JOB_SUCCESS). " +
					"For asset policies using the asset block, event_types are derived from the health_status block " +
					"and do not need to be set explicitly.",
				Optional:    true,
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
						},
						"asset_selection": schema.StringAttribute{
							Description: "An asset selection string (e.g. 'tag:my-tag'). Mutually exclusive with all_assets and asset_key.",
							Optional:    true,
						},
						"asset_key": schema.StringAttribute{
							Description: "A specific asset key path (e.g. 'my_asset' or 'group/my_asset'). Mutually exclusive with all_assets and asset_selection.",
							Optional:    true,
						},
						"health_status": schema.StringAttribute{
							Description: "Trigger on a specific health status transition: degraded, warning, or healthy.",
							Optional:    true,
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
	var plan alertPolicyResourceModel
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

	resp.Diagnostics.Append(policyToModel(created, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state alertPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetAlertPolicy(ctx, state.Deployment.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert policy", err.Error())
		return
	}

	resp.Diagnostics.Append(policyToModel(policy, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *alertPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan alertPolicyResourceModel
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

	resp.Diagnostics.Append(policyToModel(updated, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *alertPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state alertPolicyResourceModel
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

	var state alertPolicyResourceModel
	state.Deployment = types.StringValue(parts[0])
	resp.Diagnostics.Append(policyToModel(policy, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// modelToPolicy converts the Terraform model to the client AlertPolicy type.
func modelToPolicy(ctx context.Context, model alertPolicyResourceModel) (client.AlertPolicy, diag.Diagnostics) {
	var diags diag.Diagnostics

	policy := client.AlertPolicy{
		Name:        model.Name.ValueString(),
		Description: model.Description.ValueString(),
		PolicyType:  model.PolicyType.ValueString(),
		Enabled:     model.Enabled.ValueBool(),
	}

	// Asset block — derives event_types and alert_targets.
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

		switch a.HealthStatus.ValueString() {
		case "degraded":
			policy.EventTypes = append(policy.EventTypes, "ASSET_HEALTH_DEGRADED")
		case "warning":
			policy.EventTypes = append(policy.EventTypes, "ASSET_HEALTH_WARNING")
		case "healthy":
			policy.EventTypes = append(policy.EventTypes, "ASSET_HEALTH_HEALTHY")
		}
	} else {
		// Non-asset policy: use explicit event_types.
		diags.Append(model.EventTypes.ElementsAs(ctx, &policy.EventTypes, false)...)
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

// policyToModel populates the Terraform model from the client AlertPolicy.
func policyToModel(policy *client.AlertPolicy, model *alertPolicyResourceModel) diag.Diagnostics {
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

	// Reconstruct asset block if policy has asset target or health status events.
	if model.PolicyType.ValueString() == "asset" {
		a := assetConfigModel{
			AllAssets:      types.BoolNull(),
			AssetSelection: types.StringNull(),
			AssetKey:       types.StringNull(),
			HealthStatus:   types.StringNull(),
		}

		switch policy.AlertTargetType {
		case "all_assets":
			a.AllAssets = types.BoolValue(true)
		case "asset_selection":
			a.AssetSelection = types.StringValue(policy.AlertTargetValue)
		case "asset_key":
			a.AssetKey = types.StringValue(policy.AlertTargetValue)
		}

		for _, et := range policy.EventTypes {
			switch et {
			case "ASSET_HEALTH_DEGRADED":
				a.HealthStatus = types.StringValue("degraded")
			case "ASSET_HEALTH_WARNING":
				a.HealthStatus = types.StringValue("warning")
			case "ASSET_HEALTH_HEALTHY":
				a.HealthStatus = types.StringValue("healthy")
			}
		}
		model.Asset = []assetConfigModel{a}
	} else {
		model.Asset = []assetConfigModel{}
	}

	// Notification service.
	ns := policy.NotificationService
	model.NotificationService = notificationServiceModel{
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
