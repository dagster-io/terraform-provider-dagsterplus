package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client"
)

// --- mock response helpers ---

func mockEmailNotification() map[string]any {
	return map[string]any{
		"__typename":     "EmailAlertPolicyNotification",
		"emailAddresses": []any{"test@example.com"},
	}
}

func mockAssetPolicyRaw(name string) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "test policy",
		"enabled": true, "eventTypes": []any{"ASSET_HEALTH_DEGRADED"},
		"policyOptions":       map[string]any{"consecutiveFailureThreshold": 0},
		"alertTargets":        []any{},
		"notificationService": mockEmailNotification(),
	}
}

func mockRunPolicyRaw(name string) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "",
		"enabled": true, "eventTypes": []any{"JOB_FAILURE", "JOB_SUCCESS"},
		"policyOptions": map[string]any{"consecutiveFailureThreshold": 0},
		"alertTargets": []any{
			map[string]any{
				"__typename": "RunResultTarget",
				"jobs":       []any{}, "tags": []any{}, "locationNames": []any{},
			},
		},
		"notificationService": mockEmailNotification(),
	}
}

func mockCodeLocationPolicyRaw(name string) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "",
		"enabled": true, "eventTypes": []any{"CODE_LOCATION_ERROR"},
		"policyOptions": map[string]any{"consecutiveFailureThreshold": 0},
		"alertTargets": []any{
			map[string]any{"__typename": "CodeLocationTarget", "locationNames": []any{}},
		},
		"notificationService": mockEmailNotification(),
	}
}

func mockAutomationPolicyRaw(name string) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "",
		"enabled": true, "eventTypes": []any{"TICK_FAILURE"},
		"policyOptions": map[string]any{"consecutiveFailureThreshold": 3},
		"alertTargets": []any{
			map[string]any{
				"__typename":         "ScheduleSensorTarget",
				"types":              []any{"SCHEDULE", "SENSOR"},
				"locationNames":      []any{},
				"schedulesOrSensors": []any{},
			},
		},
		"notificationService": mockEmailNotification(),
	}
}

// mockAgentDowntimePolicyRaw builds an AGENT_UNAVAILABLE policy. renotify is the value the
// API returns for renotifyIntervalMinutes; pass 0 to model the null the API sends when no
// interval is configured (genqlient decodes a null Int to 0).
func mockAgentDowntimePolicyRaw(name string, renotify int) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "",
		"enabled": true, "eventTypes": []any{"AGENT_UNAVAILABLE"},
		"policyOptions": map[string]any{
			"consecutiveFailureThreshold": 0,
			"renotifyIntervalMinutes":     renotify,
		},
		// Agent downtime policies never carry a target.
		"alertTargets":        []any{},
		"notificationService": mockEmailNotification(),
	}
}

func mockBudgetPolicyRaw(name string) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "",
		"enabled": true, "eventTypes": []any{"INSIGHTS_CONSUMPTION_EXCEEDED"},
		"policyOptions": map[string]any{"consecutiveFailureThreshold": 0},
		"alertTargets": []any{
			map[string]any{
				"__typename":          "InsightsDeploymentThresholdTarget",
				"metricName":          "__dagster_dagster_credits",
				"operator":            "GREATER_THAN",
				"threshold":           100.0,
				"selectionPeriodDays": 30,
			},
		},
		"notificationService": mockEmailNotification(),
	}
}

func mockInsightMetricPolicyRaw(name string) map[string]any {
	return map[string]any{
		"id": "policy-id-" + name, "name": name, "description": "",
		"enabled": true, "eventTypes": []any{"INSIGHTS_CONSUMPTION_EXCEEDED"},
		"policyOptions": map[string]any{"consecutiveFailureThreshold": 0},
		"alertTargets": []any{
			map[string]any{
				"__typename":          "InsightsAssetGroupThresholdTarget",
				"assetGroup":          "my-group",
				"metricName":          "__dagster_dagster_credits",
				"operator":            "GREATER_THAN",
				"threshold":           50.0,
				"selectionPeriodDays": 7,
			},
		},
		"notificationService": mockEmailNotification(),
	}
}

// newListSrv returns a mock server that responds to any query with the given alert policies list.
func newListSrv(policies []any) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{"alertPolicies": policies})
	}))
}

// --- ListAlertPolicies ---

func TestListAlertPolicies_Success(t *testing.T) {
	srv := newListSrv([]any{mockAssetPolicyRaw("asset-policy"), mockRunPolicyRaw("run-policy")})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 2 {
		t.Fatalf("len = %d, want 2", len(policies))
	}
	if policies[0].Name != "asset-policy" {
		t.Errorf("Name[0] = %q, want asset-policy", policies[0].Name)
	}
	if policies[1].Name != "run-policy" {
		t.Errorf("Name[1] = %q, want run-policy", policies[1].Name)
	}
}

func TestListAlertPolicies_Empty(t *testing.T) {
	srv := newListSrv([]any{})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 0 {
		t.Fatalf("len = %d, want 0", len(policies))
	}
}

func TestListAlertPolicies_UsesDeploymentEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gqlOK(w, map[string]any{"alertPolicies": []any{}})
	}))
	defer srv.Close()

	if _, err := newClient(srv).ListAlertPolicies(context.Background(), "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/prod/graphql" {
		t.Errorf("path = %q, want /prod/graphql", gotPath)
	}
}

func TestListAlertPolicies_GraphQLError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlErr(w, "something went wrong")
	}))
	defer srv.Close()

	_, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "ListAlertPolicies") {
		t.Errorf("error should mention ListAlertPolicies, got: %v", err)
	}
}

// --- GetAlertPolicy ---

func TestGetAlertPolicy_Found(t *testing.T) {
	srv := newListSrv([]any{
		mockAssetPolicyRaw("target-policy"),
		mockRunPolicyRaw("other-policy"),
	})
	defer srv.Close()

	p, err := newClient(srv).GetAlertPolicy(context.Background(), "prod", "target-policy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "target-policy" {
		t.Errorf("Name = %q, want target-policy", p.Name)
	}
}

func TestGetAlertPolicy_NotFound(t *testing.T) {
	srv := newListSrv([]any{mockAssetPolicyRaw("other-policy")})
	defer srv.Close()

	_, err := newClient(srv).GetAlertPolicy(context.Background(), "prod", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

// --- CreateOrUpdateAlertPolicy ---

func TestCreateOrUpdateAlertPolicy_Success(t *testing.T) {
	var callCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		b := parseBody(t, r)
		if callCount == 1 {
			if !strings.Contains(b.Query, "CreateOrUpdateAlertPolicy") {
				t.Errorf("call 1: expected CreateOrUpdateAlertPolicy, got: %s", b.Query)
			}
			gqlOK(w, map[string]any{
				"createOrUpdateAlertPolicyFromDocument": map[string]any{
					"__typename": "AlertPolicy",
					"id":         "policy-id",
					"name":       "my-asset-policy",
				},
			})
		} else {
			// Second call: ListAlertPolicies (triggered by GetAlertPolicy inside CreateOrUpdate)
			gqlOK(w, map[string]any{"alertPolicies": []any{mockAssetPolicyRaw("my-asset-policy")}})
		}
	}))
	defer srv.Close()

	policy := client.AlertPolicy{
		Name:       "my-asset-policy",
		PolicyType: "asset",
		Enabled:    true,
		EventTypes: []string{"ASSET_HEALTH_DEGRADED"},
		NotificationService: client.AlertPolicyNotification{
			Type:           "email",
			EmailAddresses: []string{"test@example.com"},
		},
	}

	result, err := newClient(srv).CreateOrUpdateAlertPolicy(context.Background(), "prod", policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "my-asset-policy" {
		t.Errorf("Name = %q, want my-asset-policy", result.Name)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (mutation + list)", callCount)
	}
}

func TestCreateOrUpdateAlertPolicy_InvalidPolicyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateAlertPolicyFromDocument": map[string]any{
				"__typename": "InvalidAlertPolicyError",
				"message":    "policy is invalid",
				"errors":     []any{"field X is required"},
			},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateOrUpdateAlertPolicy(context.Background(), "prod", client.AlertPolicy{Name: "bad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid policy") {
		t.Errorf("error should mention invalid policy, got: %v", err)
	}
}

func TestCreateOrUpdateAlertPolicy_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"createOrUpdateAlertPolicyFromDocument": map[string]any{"__typename": "UnknownType"},
		})
	}))
	defer srv.Close()

	_, err := newClient(srv).CreateOrUpdateAlertPolicy(context.Background(), "prod", client.AlertPolicy{Name: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// genqlient itself returns an unmarshal error for unknown typenames — just verify it's non-nil.
	if err == nil {
		t.Fatal("expected error for unknown typename, got nil")
	}
}

// --- DeleteAlertPolicy ---

func TestDeleteAlertPolicy_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := parseBody(t, r)
		if !strings.Contains(b.Query, "DeleteAlertPolicy") {
			t.Errorf("expected DeleteAlertPolicy mutation, got: %s", b.Query)
		}
		if b.Variables["name"] != "my-policy" {
			t.Errorf("name = %v, want my-policy", b.Variables["name"])
		}
		gqlOK(w, map[string]any{
			"deleteAlertPolicy": map[string]any{
				"__typename":      "DeleteAlertPolicySuccess",
				"alertPolicyName": "my-policy",
			},
		})
	}))
	defer srv.Close()

	if err := newClient(srv).DeleteAlertPolicy(context.Background(), "prod", "my-policy"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteAlertPolicy_UnexpectedTypename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gqlOK(w, map[string]any{
			"deleteAlertPolicy": map[string]any{"__typename": "UnauthorizedError"},
		})
	}))
	defer srv.Close()

	err := newClient(srv).DeleteAlertPolicy(context.Background(), "prod", "my-policy")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Policy type parsing ---

func TestListAlertPolicies_ParseRunPolicy(t *testing.T) {
	srv := newListSrv([]any{mockRunPolicyRaw("run-policy")})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := policies[0]
	if p.Run == nil {
		t.Fatal("Run is nil, want non-nil")
	}
	if !p.Run.AllRuns {
		t.Errorf("AllRuns = false, want true (no jobs/tags/locations in mock)")
	}
	if !p.Run.OnFailure {
		t.Errorf("OnFailure = false, want true (JOB_FAILURE in eventTypes)")
	}
	if !p.Run.OnSuccess {
		t.Errorf("OnSuccess = false, want true (JOB_SUCCESS in eventTypes)")
	}
}

func TestListAlertPolicies_ParseCodeLocationPolicy(t *testing.T) {
	srv := newListSrv([]any{mockCodeLocationPolicyRaw("cl-policy")})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := policies[0]
	if p.CodeLocation == nil {
		t.Fatal("CodeLocation is nil, want non-nil")
	}
	if !p.CodeLocation.AllLocations {
		t.Errorf("AllLocations = false, want true (empty locationNames)")
	}
}

func TestListAlertPolicies_ParseAutomationPolicy(t *testing.T) {
	srv := newListSrv([]any{mockAutomationPolicyRaw("auto-policy")})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := policies[0]
	if p.Automation == nil {
		t.Fatal("Automation is nil, want non-nil")
	}
	if !p.Automation.AllSchedulesAndSensors {
		t.Errorf("AllSchedulesAndSensors = false, want true")
	}
	if !p.Automation.IncludeSchedules {
		t.Errorf("IncludeSchedules = false, want true")
	}
	if !p.Automation.IncludeSensors {
		t.Errorf("IncludeSensors = false, want true")
	}
	if p.Automation.MinConsecutiveFailures != 3 {
		t.Errorf("MinConsecutiveFailures = %d, want 3", p.Automation.MinConsecutiveFailures)
	}
}

func TestListAlertPolicies_ParseAgentDowntimePolicy(t *testing.T) {
	srv := newListSrv([]any{mockAgentDowntimePolicyRaw("agent-policy", 30)})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := policies[0]
	if p.AgentDowntime == nil {
		t.Fatal("AgentDowntime is nil, want non-nil")
	}
	if p.AgentDowntime.RenotifyIntervalMinutes != 30 {
		t.Errorf("RenotifyIntervalMinutes = %d, want 30", p.AgentDowntime.RenotifyIntervalMinutes)
	}
	if len(p.EventTypes) != 1 || p.EventTypes[0] != "AGENT_UNAVAILABLE" {
		t.Errorf("EventTypes = %v, want [AGENT_UNAVAILABLE]", p.EventTypes)
	}
}

// A policy with no renotify interval must leave AgentDowntime nil, so the provider renders
// an absent block rather than an empty one that would diff against the user's config forever.
func TestListAlertPolicies_ParseAgentDowntimePolicyWithoutRenotify(t *testing.T) {
	srv := newListSrv([]any{mockAgentDowntimePolicyRaw("agent-policy", 0)})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policies[0].AgentDowntime != nil {
		t.Errorf("AgentDowntime = %+v, want nil when the API returns no renotify interval", policies[0].AgentDowntime)
	}
}

// Only AGENT_UNAVAILABLE policies should pick up renotifyIntervalMinutes. The field is read
// from the shared fragment for every policy type, so a non-agent policy that happens to have
// one set must not be misread as an agent downtime policy.
func TestListAlertPolicies_RenotifyIgnoredOnNonAgentPolicy(t *testing.T) {
	raw := mockCodeLocationPolicyRaw("cl-policy")
	raw["policyOptions"] = map[string]any{
		"consecutiveFailureThreshold": 0,
		"renotifyIntervalMinutes":     15,
	}
	srv := newListSrv([]any{raw})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policies[0].AgentDowntime != nil {
		t.Errorf("AgentDowntime = %+v, want nil for a code_location policy", policies[0].AgentDowntime)
	}
	if policies[0].CodeLocation == nil {
		t.Error("CodeLocation is nil, want non-nil")
	}
}

func TestCreateOrUpdateAlertPolicy_AgentDowntimeDocument(t *testing.T) {
	tests := []struct {
		name         string
		config       *client.AgentDowntimeAlertConfig
		wantRenotify any // nil means policy_options must be absent entirely
	}{
		{"with renotify", &client.AgentDowntimeAlertConfig{RenotifyIntervalMinutes: 30}, float64(30)},
		{"without block", nil, nil},
		// A zero interval is rejected at the schema layer, but guard the client too: the
		// API accepts 0 and then reports it back as null, which would never round-trip.
		{"zero renotify", &client.AgentDowntimeAlertConfig{RenotifyIntervalMinutes: 0}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int
			var mutationVars map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount == 1 {
					mutationVars = parseBody(t, r).Variables
					gqlOK(w, map[string]any{
						"createOrUpdateAlertPolicyFromDocument": map[string]any{
							"__typename": "AlertPolicy", "id": "policy-id", "name": "agent-policy",
						},
					})
					return
				}
				gqlOK(w, map[string]any{"alertPolicies": []any{mockAgentDowntimePolicyRaw("agent-policy", 30)}})
			}))
			defer srv.Close()

			policy := client.AlertPolicy{
				Name:          "agent-policy",
				PolicyType:    "agent_downtime",
				Enabled:       true,
				AgentDowntime: tt.config,
				NotificationService: client.AlertPolicyNotification{
					Type:           "email",
					EmailAddresses: []string{"ops@example.com"},
				},
			}

			if _, err := newClient(srv).CreateOrUpdateAlertPolicy(context.Background(), "prod", policy); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			doc, ok := mutationVars["document"].(map[string]any)
			if !ok {
				t.Fatalf("document var is not a map: %T", mutationVars["document"])
			}

			events, ok := doc["event_types"].([]any)
			if !ok || len(events) != 1 || events[0] != "AGENT_UNAVAILABLE" {
				t.Errorf("event_types = %v, want [AGENT_UNAVAILABLE]", doc["event_types"])
			}

			// The API treats an absent alert_targets key as "all agents"; sending an
			// explicit empty list is unnecessary and sending a target is invalid.
			if _, present := doc["alert_targets"]; present {
				t.Errorf("alert_targets = %v, want the key to be absent", doc["alert_targets"])
			}

			opts, present := doc["policy_options"]
			if tt.wantRenotify == nil {
				if present {
					t.Errorf("policy_options = %v, want the key to be absent", opts)
				}
				return
			}
			optsMap, ok := opts.(map[string]any)
			if !ok {
				t.Fatalf("policy_options missing or wrong type: %T", opts)
			}
			if optsMap["renotify_interval_minutes"] != tt.wantRenotify {
				t.Errorf("renotify_interval_minutes = %v, want %v",
					optsMap["renotify_interval_minutes"], tt.wantRenotify)
			}
		})
	}
}

func TestListAlertPolicies_ParseBudgetPolicy(t *testing.T) {
	srv := newListSrv([]any{mockBudgetPolicyRaw("budget-policy")})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := policies[0]
	if p.Budget == nil {
		t.Fatal("Budget is nil, want non-nil")
	}
	if p.Budget.Operator != "greater_than" {
		t.Errorf("Operator = %q, want greater_than", p.Budget.Operator)
	}
	if p.Budget.Threshold != 100.0 {
		t.Errorf("Threshold = %v, want 100.0", p.Budget.Threshold)
	}
	if p.Budget.PeriodDays != 30 {
		t.Errorf("PeriodDays = %d, want 30", p.Budget.PeriodDays)
	}
}

func TestListAlertPolicies_ParseInsightMetricPolicy(t *testing.T) {
	srv := newListSrv([]any{mockInsightMetricPolicyRaw("insight-policy")})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := policies[0]
	if p.InsightMetric == nil {
		t.Fatal("InsightMetric is nil, want non-nil")
	}
	if p.InsightMetric.AssetGroup != "my-group" {
		t.Errorf("AssetGroup = %q, want my-group", p.InsightMetric.AssetGroup)
	}
	if p.InsightMetric.Metric != "dagster_credits" {
		t.Errorf("Metric = %q, want dagster_credits (friendly name)", p.InsightMetric.Metric)
	}
	if p.InsightMetric.Operator != "greater_than" {
		t.Errorf("Operator = %q, want greater_than", p.InsightMetric.Operator)
	}
	if p.InsightMetric.Threshold != 50.0 {
		t.Errorf("Threshold = %v, want 50.0", p.InsightMetric.Threshold)
	}
	if p.InsightMetric.PeriodDays != 7 {
		t.Errorf("PeriodDays = %d, want 7", p.InsightMetric.PeriodDays)
	}
}

func TestListAlertPolicies_ParseWebhookNotification(t *testing.T) {
	raw := mockAssetPolicyRaw("webhook-policy")
	raw["notificationService"] = map[string]any{
		"__typename": "WebhookAlertPolicyNotification",
		"webhookUrl": "https://api.incident.io/v2/alert_events/http/abc",
		"headers": []any{
			map[string]any{"key": "Authorization", "value": "Bearer secret"},
			map[string]any{"key": "Content-Type", "value": "application/json"},
		},
		"bodyTemplate": `{"message": "{{ alert.name }}"}`,
	}
	srv := newListSrv([]any{raw})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := policies[0].NotificationService
	if ns.Type != "webhook" {
		t.Errorf("Type = %q, want webhook", ns.Type)
	}
	if ns.WebhookURL != "https://api.incident.io/v2/alert_events/http/abc" {
		t.Errorf("WebhookURL = %q, want the incident.io URL", ns.WebhookURL)
	}
	if ns.WebhookBodyTemplate != `{"message": "{{ alert.name }}"}` {
		t.Errorf("WebhookBodyTemplate = %q, unexpected", ns.WebhookBodyTemplate)
	}
	if len(ns.WebhookHeaders) != 2 {
		t.Fatalf("len(WebhookHeaders) = %d, want 2", len(ns.WebhookHeaders))
	}
	if ns.WebhookHeaders["Authorization"] != "Bearer secret" {
		t.Errorf("Authorization header = %q, want Bearer secret", ns.WebhookHeaders["Authorization"])
	}
	if ns.WebhookHeaders["Content-Type"] != "application/json" {
		t.Errorf("Content-Type header = %q, want application/json", ns.WebhookHeaders["Content-Type"])
	}
}

func TestCreateOrUpdateAlertPolicy_WebhookDocument(t *testing.T) {
	var callCount int
	var mutationVars map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			b := parseBody(t, r)
			mutationVars = b.Variables
			gqlOK(w, map[string]any{
				"createOrUpdateAlertPolicyFromDocument": map[string]any{
					"__typename": "AlertPolicy", "id": "policy-id", "name": "webhook-policy",
				},
			})
		} else {
			raw := mockAssetPolicyRaw("webhook-policy")
			raw["notificationService"] = map[string]any{
				"__typename":   "WebhookAlertPolicyNotification",
				"webhookUrl":   "https://example.com/hook",
				"headers":      []any{map[string]any{"key": "X-Token", "value": "abc"}},
				"bodyTemplate": "{}",
			}
			gqlOK(w, map[string]any{"alertPolicies": []any{raw}})
		}
	}))
	defer srv.Close()

	policy := client.AlertPolicy{
		Name:       "webhook-policy",
		PolicyType: "asset",
		Enabled:    true,
		EventTypes: []string{"ASSET_HEALTH_DEGRADED"},
		NotificationService: client.AlertPolicyNotification{
			Type:                "webhook",
			WebhookURL:          "https://example.com/hook",
			WebhookHeaders:      map[string]string{"X-Token": "abc"},
			WebhookBodyTemplate: "{}",
		},
	}

	if _, err := newClient(srv).CreateOrUpdateAlertPolicy(context.Background(), "prod", policy); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Inspect the document sent to the API.
	doc, ok := mutationVars["document"].(map[string]any)
	if !ok {
		t.Fatalf("document var is not a map: %T", mutationVars["document"])
	}
	ns, ok := doc["notification_service"].(map[string]any)
	if !ok {
		t.Fatalf("notification_service missing or wrong type: %T", doc["notification_service"])
	}
	webhook, ok := ns["webhook"].(map[string]any)
	if !ok {
		t.Fatalf("webhook key missing or wrong type: %T", ns["webhook"])
	}
	if webhook["webhook_url"] != "https://example.com/hook" {
		t.Errorf("webhook_url = %v, want https://example.com/hook", webhook["webhook_url"])
	}
	if webhook["body_template"] != "{}" {
		t.Errorf("body_template = %v, want {}", webhook["body_template"])
	}
	headers, ok := webhook["headers"].(map[string]any)
	if !ok {
		t.Fatalf("headers missing or wrong type: %T", webhook["headers"])
	}
	if len(headers) != 1 {
		t.Fatalf("len(headers) = %d, want 1", len(headers))
	}
	if headers["X-Token"] != "abc" {
		t.Errorf("headers[X-Token] = %v, want abc", headers["X-Token"])
	}
}

func TestListAlertPolicies_ParseSlackNotification(t *testing.T) {
	raw := mockAssetPolicyRaw("slack-policy")
	raw["notificationService"] = map[string]any{
		"__typename":         "SlackAlertPolicyNotification",
		"slackWorkspaceName": "my-workspace",
		"slackChannelName":   "#alerts",
	}
	srv := newListSrv([]any{raw})
	defer srv.Close()

	policies, err := newClient(srv).ListAlertPolicies(context.Background(), "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ns := policies[0].NotificationService
	if ns.Type != "slack" {
		t.Errorf("Type = %q, want slack", ns.Type)
	}
	if ns.SlackWorkspaceName != "my-workspace" {
		t.Errorf("SlackWorkspaceName = %q, want my-workspace", ns.SlackWorkspaceName)
	}
	if ns.SlackChannelName != "#alerts" {
		t.Errorf("SlackChannelName = %q, want #alerts", ns.SlackChannelName)
	}
}
