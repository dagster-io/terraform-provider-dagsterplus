package resources

import (
	"context"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// helpers

func strPtr(s string) types.String { return types.StringValue(s) }

func makeRunModel(onSuccess, onFailure bool) AlertPolicyResourceModel {
	return AlertPolicyResourceModel{
		Name:        strPtr("test-run-policy"),
		Description: strPtr(""),
		PolicyType:  strPtr("run"),
		Enabled:     types.BoolValue(true),
		Run: []RunConfigModel{
			{
				AllRuns:        types.BoolValue(true),
				Tags:           types.StringNull(),
				JobNames:       types.ListNull(types.StringType),
				CodeLocations:  types.ListNull(types.StringType),
				OnSuccess:      types.BoolValue(onSuccess),
				OnFailure:      types.BoolValue(onFailure),
				OnTimeoutHours: types.Int64Null(),
			},
		},
		NotificationService: NotificationServiceModel{
			Type:                  strPtr("email"),
			EmailAddresses:        types.ListValueMust(types.StringType, []attr.Value{strPtr("test@example.com")}),
			DefaultEmailAddresses: types.ListNull(types.StringType),
			SlackWorkspaceName:    types.StringNull(),
			SlackChannelName:      types.StringNull(),
			WebhookURL:            types.StringNull(),
			IntegrationKey:        types.StringNull(),
		},
	}
}

// --- modelToPolicy tests ---

func TestModelToPolicy_RunOnFailure(t *testing.T) {
	ctx := context.Background()
	m := makeRunModel(false, true)

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if policy.PolicyType != "run" {
		t.Errorf("expected PolicyType run, got %q", policy.PolicyType)
	}
	if policy.Run == nil {
		t.Fatal("expected Run config to be set")
	}
	if policy.Run.OnFailure != true {
		t.Error("expected OnFailure true")
	}
	if policy.Run.OnSuccess != false {
		t.Error("expected OnSuccess false")
	}
	if !containsString(policy.EventTypes, "JOB_FAILURE") {
		t.Errorf("expected JOB_FAILURE in event types, got %v", policy.EventTypes)
	}
	if containsString(policy.EventTypes, "JOB_SUCCESS") {
		t.Errorf("did not expect JOB_SUCCESS in event types, got %v", policy.EventTypes)
	}
}

func TestModelToPolicy_RunOnSuccess(t *testing.T) {
	ctx := context.Background()
	m := makeRunModel(true, false)

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if !containsString(policy.EventTypes, "JOB_SUCCESS") {
		t.Errorf("expected JOB_SUCCESS in event types, got %v", policy.EventTypes)
	}
	if containsString(policy.EventTypes, "JOB_FAILURE") {
		t.Errorf("did not expect JOB_FAILURE in event types")
	}
}

func TestModelToPolicy_AssetAllAssets(t *testing.T) {
	ctx := context.Background()
	m := AlertPolicyResourceModel{
		Name:       strPtr("asset-policy"),
		PolicyType: strPtr("asset"),
		Enabled:    types.BoolValue(true),
		Asset: []AssetConfigModel{
			{
				AllAssets:      types.BoolValue(true),
				AssetSelection: types.StringNull(),
				AssetKey:       types.StringNull(),
				HealthStatus:   types.StringNull(),
				SpecificEvents: types.ListNull(types.StringType),
			},
		},
		NotificationService: NotificationServiceModel{
			Type:                  strPtr("slack"),
			EmailAddresses:        types.ListNull(types.StringType),
			DefaultEmailAddresses: types.ListNull(types.StringType),
			SlackWorkspaceName:    strPtr("my-workspace"),
			SlackChannelName:      strPtr("#alerts"),
			WebhookURL:            types.StringNull(),
			IntegrationKey:        types.StringNull(),
		},
	}

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if policy.AlertTargetType != "all_assets" {
		t.Errorf("expected AlertTargetType all_assets, got %q", policy.AlertTargetType)
	}
	if policy.NotificationService.Type != "slack" {
		t.Errorf("expected notification type slack, got %q", policy.NotificationService.Type)
	}
	if policy.NotificationService.SlackWorkspaceName != "my-workspace" {
		t.Errorf("expected SlackWorkspaceName my-workspace, got %q", policy.NotificationService.SlackWorkspaceName)
	}
	if policy.NotificationService.SlackChannelName != "#alerts" {
		t.Errorf("expected SlackChannelName #alerts, got %q", policy.NotificationService.SlackChannelName)
	}
}

func TestModelToPolicy_AssetSpecificEvents(t *testing.T) {
	ctx := context.Background()
	specificEvents := types.ListValueMust(types.StringType, []attr.Value{
		strPtr("materialization_success"),
		strPtr("materialization_failure"),
	})
	m := AlertPolicyResourceModel{
		Name:       strPtr("asset-events-policy"),
		PolicyType: strPtr("asset"),
		Enabled:    types.BoolValue(true),
		Asset: []AssetConfigModel{
			{
				AllAssets:      types.BoolNull(),
				AssetSelection: types.StringNull(),
				AssetKey:       strPtr("my_asset"),
				HealthStatus:   types.StringNull(),
				SpecificEvents: specificEvents,
			},
		},
		NotificationService: NotificationServiceModel{
			Type:                  strPtr("pagerduty"),
			IntegrationKey:        strPtr("pd-key-123"),
			EmailAddresses:        types.ListNull(types.StringType),
			DefaultEmailAddresses: types.ListNull(types.StringType),
			SlackWorkspaceName:    types.StringNull(),
			SlackChannelName:      types.StringNull(),
			WebhookURL:            types.StringNull(),
		},
	}

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if policy.AlertTargetType != "asset_key" {
		t.Errorf("expected AlertTargetType asset_key, got %q", policy.AlertTargetType)
	}
	if policy.AlertTargetValue != "my_asset" {
		t.Errorf("expected AlertTargetValue my_asset, got %q", policy.AlertTargetValue)
	}
	if !containsString(policy.EventTypes, "ASSET_MATERIALIZATION_SUCCESS") {
		t.Errorf("expected ASSET_MATERIALIZATION_SUCCESS in event types, got %v", policy.EventTypes)
	}
	if !containsString(policy.EventTypes, "ASSET_MATERIALIZATION_FAILURE") {
		t.Errorf("expected ASSET_MATERIALIZATION_FAILURE in event types, got %v", policy.EventTypes)
	}
	if policy.NotificationService.IntegrationKey != "pd-key-123" {
		t.Errorf("expected IntegrationKey pd-key-123, got %q", policy.NotificationService.IntegrationKey)
	}
}

func TestModelToPolicy_CodeLocation(t *testing.T) {
	ctx := context.Background()
	m := AlertPolicyResourceModel{
		Name:       strPtr("cl-policy"),
		PolicyType: strPtr("code_location"),
		Enabled:    types.BoolValue(true),
		CodeLocation: []CodeLocationConfigModel{
			{
				AllLocations: types.BoolValue(false),
				LocationName: strPtr("my_repo"),
			},
		},
		NotificationService: NotificationServiceModel{
			Type:                  strPtr("email"),
			EmailAddresses:        types.ListValueMust(types.StringType, []attr.Value{strPtr("oncall@example.com")}),
			DefaultEmailAddresses: types.ListNull(types.StringType),
			SlackWorkspaceName:    types.StringNull(),
			SlackChannelName:      types.StringNull(),
			WebhookURL:            types.StringNull(),
			IntegrationKey:        types.StringNull(),
		},
	}

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if policy.CodeLocation == nil {
		t.Fatal("expected CodeLocation config to be set")
	}
	if policy.CodeLocation.LocationName != "my_repo" {
		t.Errorf("expected LocationName my_repo, got %q", policy.CodeLocation.LocationName)
	}
	if !containsString(policy.EventTypes, "CODE_LOCATION_ERROR") {
		t.Errorf("expected CODE_LOCATION_ERROR in event types, got %v", policy.EventTypes)
	}
}

func TestModelToPolicy_NotificationSlack(t *testing.T) {
	ctx := context.Background()
	m := AlertPolicyResourceModel{
		Name:       strPtr("slack-policy"),
		PolicyType: strPtr("run"),
		Enabled:    types.BoolValue(true),
		Run: []RunConfigModel{
			{
				AllRuns:        types.BoolValue(true),
				Tags:           types.StringNull(),
				JobNames:       types.ListNull(types.StringType),
				CodeLocations:  types.ListNull(types.StringType),
				OnSuccess:      types.BoolValue(true),
				OnFailure:      types.BoolValue(false),
				OnTimeoutHours: types.Int64Null(),
			},
		},
		NotificationService: NotificationServiceModel{
			Type:                  strPtr("slack"),
			EmailAddresses:        types.ListNull(types.StringType),
			DefaultEmailAddresses: types.ListNull(types.StringType),
			SlackWorkspaceName:    strPtr("dagster-ws"),
			SlackChannelName:      strPtr("#run-alerts"),
			WebhookURL:            types.StringNull(),
			IntegrationKey:        types.StringNull(),
		},
	}

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	ns := policy.NotificationService
	if ns.Type != "slack" {
		t.Errorf("expected notification type slack, got %q", ns.Type)
	}
	if ns.SlackWorkspaceName != "dagster-ws" {
		t.Errorf("expected SlackWorkspaceName dagster-ws, got %q", ns.SlackWorkspaceName)
	}
	if ns.SlackChannelName != "#run-alerts" {
		t.Errorf("expected SlackChannelName #run-alerts, got %q", ns.SlackChannelName)
	}
}

func TestModelToPolicy_NotificationEmail(t *testing.T) {
	ctx := context.Background()
	m := makeRunModel(true, true)
	m.NotificationService = NotificationServiceModel{
		Type: strPtr("email"),
		EmailAddresses: types.ListValueMust(types.StringType, []attr.Value{
			strPtr("a@example.com"),
			strPtr("b@example.com"),
		}),
		DefaultEmailAddresses: types.ListNull(types.StringType),
		SlackWorkspaceName:    types.StringNull(),
		SlackChannelName:      types.StringNull(),
		WebhookURL:            types.StringNull(),
		IntegrationKey:        types.StringNull(),
	}

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	ns := policy.NotificationService
	if ns.Type != "email" {
		t.Errorf("expected notification type email, got %q", ns.Type)
	}
	if len(ns.EmailAddresses) != 2 {
		t.Fatalf("expected 2 email addresses, got %d", len(ns.EmailAddresses))
	}
}

func TestModelToPolicy_NotificationPagerDuty(t *testing.T) {
	ctx := context.Background()
	m := makeRunModel(false, true)
	m.NotificationService = NotificationServiceModel{
		Type:                  strPtr("pagerduty"),
		EmailAddresses:        types.ListNull(types.StringType),
		DefaultEmailAddresses: types.ListNull(types.StringType),
		SlackWorkspaceName:    types.StringNull(),
		SlackChannelName:      types.StringNull(),
		WebhookURL:            types.StringNull(),
		IntegrationKey:        strPtr("my-pd-key"),
	}

	policy, diags := modelToPolicy(ctx, m)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}

	if policy.NotificationService.IntegrationKey != "my-pd-key" {
		t.Errorf("expected IntegrationKey my-pd-key, got %q", policy.NotificationService.IntegrationKey)
	}
}

// --- PolicyToModel tests ---

func TestPolicyToModel_RunPolicy(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:        "my-run-policy",
		Description: "alerts on run failure",
		Enabled:     true,
		EventTypes:  []string{"JOB_FAILURE"},
		Run: &client.RunAlertConfig{
			AllRuns:   false,
			OnFailure: true,
			Filter: client.RunTargetFilter{
				JobNames: []string{"my_job"},
			},
		},
		NotificationService: client.AlertPolicyNotification{
			Type:           "pagerduty",
			IntegrationKey: "pd-key",
		},
	}

	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("run"),
	}

	diags := PolicyToModel(policy, model)
	if diags != nil && diags.HasError() {
		t.Fatalf("PolicyToModel diags: %v", diags)
	}

	if model.Name.ValueString() != "my-run-policy" {
		t.Errorf("expected Name my-run-policy, got %q", model.Name.ValueString())
	}
	if !model.Enabled.ValueBool() {
		t.Error("expected Enabled true")
	}
	if len(model.Run) != 1 {
		t.Fatalf("expected 1 Run block, got %d", len(model.Run))
	}
	r := model.Run[0]
	if r.OnFailure.IsNull() || !r.OnFailure.ValueBool() {
		t.Error("expected OnFailure true")
	}
	if len(model.Asset) != 0 {
		t.Errorf("expected empty Asset, got %d", len(model.Asset))
	}
	// PolicyType should be preserved (not overwritten by PolicyToModel)
	if model.PolicyType.ValueString() != "run" {
		t.Errorf("PolicyType should be preserved as run, got %q", model.PolicyType.ValueString())
	}
}

func TestPolicyToModel_PolicyTypePreserved(t *testing.T) {
	// PolicyToModel must NOT overwrite policy_type since the API does not return it.
	policy := &client.AlertPolicy{
		Name:             "asset-policy",
		Description:      "",
		Enabled:          false,
		EventTypes:       []string{"ASSET_MATERIALIZATION_SUCCESS"},
		AlertTargetType:  "all_assets",
		AlertTargetValue: "",
		NotificationService: client.AlertPolicyNotification{
			Type:           "email",
			EmailAddresses: []string{"test@example.com"},
		},
	}

	model := &AlertPolicyResourceModel{
		Deployment: strPtr("staging"),
		PolicyType: strPtr("asset"), // set before call, should stay
	}

	PolicyToModel(policy, model)

	if model.PolicyType.ValueString() != "asset" {
		t.Errorf("PolicyType should remain asset, got %q", model.PolicyType.ValueString())
	}
}

func TestPolicyToModel_AssetHealthStatus(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:             "health-policy",
		Description:      "",
		Enabled:          true,
		EventTypes:       []string{"ASSET_HEALTH_DEGRADED"},
		AlertTargetType:  "all_assets",
		AlertTargetValue: "",
		NotificationService: client.AlertPolicyNotification{
			Type:           "email",
			EmailAddresses: []string{"a@example.com"},
		},
	}

	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("asset"),
	}

	PolicyToModel(policy, model)

	if len(model.Asset) != 1 {
		t.Fatalf("expected 1 Asset block, got %d", len(model.Asset))
	}
	a := model.Asset[0]
	if a.HealthStatus.ValueString() != "degraded" {
		t.Errorf("expected health_status degraded, got %q", a.HealthStatus.ValueString())
	}
}

func TestPolicyToModel_CodeLocation(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:       "cl-policy",
		Enabled:    true,
		EventTypes: []string{"CODE_LOCATION_ERROR"},
		CodeLocation: &client.CodeLocationAlertConfig{
			AllLocations: false,
			LocationName: "my_repo",
		},
		NotificationService: client.AlertPolicyNotification{
			Type:           "email",
			EmailAddresses: []string{"a@example.com"},
		},
	}

	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("code_location"),
	}

	PolicyToModel(policy, model)

	if len(model.CodeLocation) != 1 {
		t.Fatalf("expected 1 CodeLocation block, got %d", len(model.CodeLocation))
	}
	cl := model.CodeLocation[0]
	if cl.LocationName.ValueString() != "my_repo" {
		t.Errorf("expected LocationName my_repo, got %q", cl.LocationName.ValueString())
	}
	if cl.AllLocations.ValueBool() {
		t.Error("expected AllLocations false")
	}
}

func TestPolicyToModel_NotificationSlack(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:       "slack-policy",
		Enabled:    true,
		EventTypes: []string{"JOB_FAILURE"},
		Run: &client.RunAlertConfig{
			AllRuns:   true,
			OnFailure: true,
		},
		NotificationService: client.AlertPolicyNotification{
			Type:               "slack",
			SlackWorkspaceName: "my-ws",
			SlackChannelName:   "#alerts",
		},
	}

	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("run"),
	}

	PolicyToModel(policy, model)

	ns := model.NotificationService
	if ns.Type.ValueString() != "slack" {
		t.Errorf("expected notification type slack, got %q", ns.Type.ValueString())
	}
	if ns.SlackWorkspaceName.ValueString() != "my-ws" {
		t.Errorf("expected SlackWorkspaceName my-ws, got %q", ns.SlackWorkspaceName.ValueString())
	}
	if ns.SlackChannelName.ValueString() != "#alerts" {
		t.Errorf("expected SlackChannelName #alerts, got %q", ns.SlackChannelName.ValueString())
	}
}

func TestPolicyToModel_EventTypes(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:       "multi-event-policy",
		Enabled:    true,
		EventTypes: []string{"JOB_SUCCESS", "JOB_FAILURE"},
		Run: &client.RunAlertConfig{
			AllRuns:   true,
			OnSuccess: true,
			OnFailure: true,
		},
		NotificationService: client.AlertPolicyNotification{
			Type:           "email",
			EmailAddresses: []string{"x@example.com"},
		},
	}

	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("run"),
	}

	PolicyToModel(policy, model)

	var ets []string
	model.EventTypes.ElementsAs(context.Background(), &ets, false)
	if !containsString(ets, "JOB_SUCCESS") || !containsString(ets, "JOB_FAILURE") {
		t.Errorf("expected both JOB_SUCCESS and JOB_FAILURE in EventTypes, got %v", ets)
	}
}

// --- helper ---

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// --- agent_downtime ---

func makeAgentDowntimeModel(block []AgentDowntimeConfigModel) AlertPolicyResourceModel {
	return AlertPolicyResourceModel{
		Name:          strPtr("test-agent-policy"),
		Description:   strPtr(""),
		PolicyType:    strPtr("agent_downtime"),
		Enabled:       types.BoolValue(true),
		AgentDowntime: block,
		NotificationService: NotificationServiceModel{
			Type:                  strPtr("email"),
			EmailAddresses:        types.ListValueMust(types.StringType, []attr.Value{strPtr("ops@example.com")}),
			DefaultEmailAddresses: types.ListNull(types.StringType),
			SlackWorkspaceName:    types.StringNull(),
			SlackChannelName:      types.StringNull(),
			WebhookURL:            types.StringNull(),
			IntegrationKey:        types.StringNull(),
		},
	}
}

func TestModelToPolicy_AgentDowntimeWithRenotify(t *testing.T) {
	model := makeAgentDowntimeModel([]AgentDowntimeConfigModel{
		{RenotifyIntervalMinutes: types.Int64Value(30)},
	})

	policy, diags := modelToPolicy(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if policy.AgentDowntime == nil {
		t.Fatal("AgentDowntime is nil, want non-nil")
	}
	if policy.AgentDowntime.RenotifyIntervalMinutes != 30 {
		t.Errorf("RenotifyIntervalMinutes = %d, want 30", policy.AgentDowntime.RenotifyIntervalMinutes)
	}
	if len(policy.EventTypes) != 1 || policy.EventTypes[0] != "AGENT_UNAVAILABLE" {
		t.Errorf("EventTypes = %v, want [AGENT_UNAVAILABLE]", policy.EventTypes)
	}
}

func TestModelToPolicy_AgentDowntimeWithoutBlock(t *testing.T) {
	model := makeAgentDowntimeModel(nil)

	policy, diags := modelToPolicy(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if policy.AgentDowntime != nil {
		t.Errorf("AgentDowntime = %+v, want nil when the block is omitted", policy.AgentDowntime)
	}
	if policy.PolicyType != "agent_downtime" {
		t.Errorf("PolicyType = %q, want agent_downtime", policy.PolicyType)
	}
}

func TestPolicyToModel_AgentDowntimeWithRenotify(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:          "agent-policy",
		Enabled:       true,
		EventTypes:    []string{"AGENT_UNAVAILABLE"},
		AgentDowntime: &client.AgentDowntimeAlertConfig{RenotifyIntervalMinutes: 45},
		NotificationService: client.AlertPolicyNotification{
			Type: "email", EmailAddresses: []string{"ops@example.com"},
		},
	}
	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("agent_downtime"),
	}

	PolicyToModel(policy, model)

	if len(model.AgentDowntime) != 1 {
		t.Fatalf("len(AgentDowntime) = %d, want 1", len(model.AgentDowntime))
	}
	if got := model.AgentDowntime[0].RenotifyIntervalMinutes.ValueInt64(); got != 45 {
		t.Errorf("RenotifyIntervalMinutes = %d, want 45", got)
	}
}

// A policy with no renotify interval must render as zero blocks. Emitting an empty block
// here would diff forever against a config that legitimately omits it.
func TestPolicyToModel_AgentDowntimeWithoutRenotifyEmitsNoBlock(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:       "agent-policy",
		Enabled:    true,
		EventTypes: []string{"AGENT_UNAVAILABLE"},
		NotificationService: client.AlertPolicyNotification{
			Type: "email", EmailAddresses: []string{"ops@example.com"},
		},
	}
	model := &AlertPolicyResourceModel{
		Deployment: strPtr("prod"),
		PolicyType: strPtr("agent_downtime"),
	}

	PolicyToModel(policy, model)

	if len(model.AgentDowntime) != 0 {
		t.Errorf("len(AgentDowntime) = %d, want 0", len(model.AgentDowntime))
	}
}

// On import policy_type is empty and must be inferred. Agent downtime policies carry no
// alert target and, without a renotify interval, no config struct either — so they can only
// be recognised by event type. Getting this wrong imports them as "asset".
func TestPolicyToModel_AgentDowntimeInferredOnImport(t *testing.T) {
	for _, tt := range []struct {
		name   string
		config *client.AgentDowntimeAlertConfig
	}{
		{"with renotify", &client.AgentDowntimeAlertConfig{RenotifyIntervalMinutes: 10}},
		{"without renotify", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			policy := &client.AlertPolicy{
				Name:          "agent-policy",
				Enabled:       true,
				EventTypes:    []string{"AGENT_UNAVAILABLE"},
				AgentDowntime: tt.config,
				// The client defaults this for every policy; it must not pull the
				// inference toward "asset".
				AlertTargetType: "all_assets",
				NotificationService: client.AlertPolicyNotification{
					Type: "email", EmailAddresses: []string{"ops@example.com"},
				},
			}
			model := &AlertPolicyResourceModel{Deployment: strPtr("prod")}

			PolicyToModel(policy, model)

			if model.PolicyType.ValueString() != "agent_downtime" {
				t.Errorf("PolicyType = %q, want agent_downtime", model.PolicyType.ValueString())
			}
			if len(model.Asset) != 0 {
				t.Errorf("len(Asset) = %d, want 0 — asset block leaked into an agent policy", len(model.Asset))
			}
		})
	}
}

// Switching a resource's policy_type must not leave the previous type's block in state.
func TestPolicyToModel_AgentDowntimeClearsOtherBlocks(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:          "agent-policy",
		Enabled:       true,
		EventTypes:    []string{"AGENT_UNAVAILABLE"},
		AgentDowntime: &client.AgentDowntimeAlertConfig{RenotifyIntervalMinutes: 5},
		NotificationService: client.AlertPolicyNotification{
			Type: "email", EmailAddresses: []string{"ops@example.com"},
		},
	}
	model := &AlertPolicyResourceModel{
		Deployment:   strPtr("prod"),
		PolicyType:   strPtr("agent_downtime"),
		CodeLocation: []CodeLocationConfigModel{{AllLocations: types.BoolValue(true)}},
		Budget:       []BudgetConfigModel{{Threshold: types.Float64Value(1)}},
	}

	PolicyToModel(policy, model)

	if len(model.CodeLocation) != 0 {
		t.Errorf("len(CodeLocation) = %d, want 0", len(model.CodeLocation))
	}
	if len(model.Budget) != 0 {
		t.Errorf("len(Budget) = %d, want 0", len(model.Budget))
	}
	if len(model.AgentDowntime) != 1 {
		t.Errorf("len(AgentDowntime) = %d, want 1", len(model.AgentDowntime))
	}
}

// The reverse of the above: a non-agent policy must not retain a stale agent_downtime block.
func TestPolicyToModel_NonAgentClearsAgentDowntimeBlock(t *testing.T) {
	policy := &client.AlertPolicy{
		Name:         "cl-policy",
		Enabled:      true,
		EventTypes:   []string{"CODE_LOCATION_ERROR"},
		CodeLocation: &client.CodeLocationAlertConfig{AllLocations: true},
		NotificationService: client.AlertPolicyNotification{
			Type: "email", EmailAddresses: []string{"ops@example.com"},
		},
	}
	model := &AlertPolicyResourceModel{
		Deployment:    strPtr("prod"),
		PolicyType:    strPtr("code_location"),
		AgentDowntime: []AgentDowntimeConfigModel{{RenotifyIntervalMinutes: types.Int64Value(30)}},
	}

	PolicyToModel(policy, model)

	if len(model.AgentDowntime) != 0 {
		t.Errorf("len(AgentDowntime) = %d, want 0", len(model.AgentDowntime))
	}
}
