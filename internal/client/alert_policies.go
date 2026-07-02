package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dagster-io/terraform-provider-dagsterplus/internal/client/schema"
)

// AlertPolicy represents a Dagster+ alert policy.
type AlertPolicy struct {
	ID                  string
	Name                string
	Description         string
	PolicyType          string // "asset" | "run" | "code_location" | "automation" | "budget" | "insight_metric"
	Enabled             bool
	EventTypes          []string
	AlertTargetType     string // "all_assets" | "asset_selection" | "asset_key"
	AlertTargetValue    string // asset selection string or asset key (empty for all_assets)
	NotificationService AlertPolicyNotification
	Run                 *RunAlertConfig
	CodeLocation        *CodeLocationAlertConfig
	Automation          *AutomationAlertConfig
	Budget              *BudgetAlertConfig
	InsightMetric       *InsightMetricAlertConfig
}

// AlertPolicyNotification holds the notification service config for an alert policy.
// Exactly one notification type should be populated, indicated by Type.
type AlertPolicyNotification struct {
	Type                  string            // "email" | "email_owners" | "slack" | "microsoft_teams" | "pagerduty" | "webhook"
	EmailAddresses        []string          // email
	DefaultEmailAddresses []string          // email_owners (optional fallback addresses)
	SlackWorkspaceName    string            // slack
	SlackChannelName      string            // slack
	WebhookURL            string            // microsoft_teams, webhook
	IntegrationKey        string            // pagerduty
	WebhookHeaders        map[string]string // webhook (arbitrary request headers)
	WebhookBodyTemplate   string            // webhook (templated request body)
}

// RunTargetFilter filters run targets.
type RunTargetFilter struct {
	Tags          string // "key=value"; empty = no filter
	JobNames      []string
	LocationNames []string
}

// RunAlertConfig holds the configuration for a run-type alert policy.
type RunAlertConfig struct {
	AllRuns        bool
	Filter         RunTargetFilter
	OnSuccess      bool
	OnFailure      bool
	OnTimeoutHours int // 0 = not set
}

// CodeLocationAlertConfig holds the configuration for a code_location-type alert policy.
type CodeLocationAlertConfig struct {
	AllLocations bool
	LocationName string
}

// AutomationAlertConfig holds the configuration for an automation-type alert policy.
type AutomationAlertConfig struct {
	AllSchedulesAndSensors bool
	LocationNames          []string
	SchedulesAndSensors    []string
	IncludeSchedules       bool
	IncludeSensors         bool
	MinConsecutiveFailures int
}

// BudgetAlertConfig holds the configuration for a budget-type alert policy.
type BudgetAlertConfig struct {
	Operator   string // "greater_than" | "less_than"
	Threshold  float64
	PeriodDays int
}

// InsightMetricAlertConfig holds the configuration for an insight_metric-type alert policy.
type InsightMetricAlertConfig struct {
	AssetGroup string
	AssetKey   string // slash-delimited
	JobName    string
	Metric     string // "dagster_credits"
	Operator   string // "greater_than" | "less_than"
	Threshold  float64
	PeriodDays int
}

// metricNameToAPI maps user-friendly metric names to Dagster+ API metric name strings.
var metricNameToAPI = map[string]string{
	"dagster_credits": "__dagster_dagster_credits",
}

// apiToMetricName is the reverse of metricNameToAPI.
var apiToMetricName = func() map[string]string {
	m := make(map[string]string, len(metricNameToAPI))
	for k, v := range metricNameToAPI {
		m[v] = k
	}
	return m
}()

func alertFieldsToPolicy(f schema.AlertPolicyFields) AlertPolicy {
	ns := AlertPolicyNotification{}
	switch n := f.NotificationService.(type) {
	case *schema.AlertPolicyFieldsNotificationServiceEmailAlertPolicyNotification:
		ns.Type = "email"
		ns.EmailAddresses = n.EmailAddresses
	case *schema.AlertPolicyFieldsNotificationServiceEmailOwnersAlertPolicyNotification:
		ns.Type = "email_owners"
		ns.DefaultEmailAddresses = n.DefaultEmailAddresses
	case *schema.AlertPolicyFieldsNotificationServiceSlackAlertPolicyNotification:
		ns.Type = "slack"
		ns.SlackWorkspaceName = n.SlackWorkspaceName
		ns.SlackChannelName = n.SlackChannelName
	case *schema.AlertPolicyFieldsNotificationServiceMicrosoftTeamsAlertPolicyNotification:
		ns.Type = "microsoft_teams"
		ns.WebhookURL = n.WebhookUrl
	case *schema.AlertPolicyFieldsNotificationServicePagerdutyAlertPolicyNotification:
		ns.Type = "pagerduty"
		ns.IntegrationKey = n.IntegrationKey
	case *schema.AlertPolicyFieldsNotificationServiceWebhookAlertPolicyNotification:
		ns.Type = "webhook"
		ns.WebhookURL = n.WebhookUrl
		ns.WebhookBodyTemplate = n.BodyTemplate
		if len(n.Headers) > 0 {
			ns.WebhookHeaders = make(map[string]string, len(n.Headers))
			for _, h := range n.Headers {
				ns.WebhookHeaders[h.Key] = h.Value
			}
		}
	}

	eventTypes := make([]string, len(f.EventTypes))
	for i, et := range f.EventTypes {
		eventTypes[i] = string(et)
	}

	targetType := "all_assets"
	targetValue := ""
	var runCfg *RunAlertConfig
	var codeCfg *CodeLocationAlertConfig
	var autoCfg *AutomationAlertConfig
	var budgetCfg *BudgetAlertConfig
	var insightCfg *InsightMetricAlertConfig

	for _, t := range f.AlertTargets {
		switch target := t.(type) {
		case *schema.AlertPolicyFieldsAlertTargetsAssetSelectionTarget:
			targetType = "asset_selection"
			targetValue = target.AssetSelectionString
		case *schema.AlertPolicyFieldsAlertTargetsAssetKeyTarget:
			targetType = "asset_key"
			targetValue = strings.Join(target.AssetKey.Path, "/")
		case *schema.AlertPolicyFieldsAlertTargetsRunResultTarget:
			if runCfg == nil {
				runCfg = &RunAlertConfig{}
			}
			runCfg.Filter.LocationNames = target.CodeLocationNames
			for _, j := range target.Jobs {
				runCfg.Filter.JobNames = append(runCfg.Filter.JobNames, j.JobName)
			}
			if len(target.Tags) > 0 {
				runCfg.Filter.Tags = target.Tags[0].Key + "=" + target.Tags[0].Value
			}
			runCfg.AllRuns = len(target.Jobs) == 0 && len(target.Tags) == 0 && len(target.CodeLocationNames) == 0
		case *schema.AlertPolicyFieldsAlertTargetsLongRunningJobThresholdTarget:
			if runCfg == nil {
				runCfg = &RunAlertConfig{}
				runCfg.Filter.LocationNames = target.CodeLocationNames
				for _, j := range target.Jobs {
					runCfg.Filter.JobNames = append(runCfg.Filter.JobNames, j.JobName)
				}
				if len(target.Tags) > 0 {
					runCfg.Filter.Tags = target.Tags[0].Key + "=" + target.Tags[0].Value
				}
				runCfg.AllRuns = len(target.Jobs) == 0 && len(target.Tags) == 0 && len(target.CodeLocationNames) == 0
			}
			runCfg.OnTimeoutHours = int(target.ThresholdSeconds) / 3600
		case *schema.AlertPolicyFieldsAlertTargetsCodeLocationTarget:
			codeCfg = &CodeLocationAlertConfig{}
			if len(target.CodeLocationNames) == 0 {
				codeCfg.AllLocations = true
			} else {
				codeCfg.LocationName = target.CodeLocationNames[0]
			}
		case *schema.AlertPolicyFieldsAlertTargetsScheduleSensorTarget:
			autoCfg = &AutomationAlertConfig{
				MinConsecutiveFailures: f.PolicyOptions.ConsecutiveFailureThreshold,
				LocationNames:          target.CodeLocationNames,
			}
			for _, tt := range target.Types {
				switch tt {
				case "SCHEDULE":
					autoCfg.IncludeSchedules = true
				case "SENSOR":
					autoCfg.IncludeSensors = true
				}
			}
			for _, s := range target.SchedulesSensors {
				autoCfg.SchedulesAndSensors = append(autoCfg.SchedulesAndSensors, s.Name)
			}
			autoCfg.AllSchedulesAndSensors = len(target.CodeLocationNames) == 0 && len(target.SchedulesSensors) == 0
		case *schema.AlertPolicyFieldsAlertTargetsInsightsDeploymentThresholdTarget:
			metric := target.MetricName
			if friendly, ok := apiToMetricName[metric]; ok {
				metric = friendly
			}
			op := strings.ToLower(string(target.Operator))
			budgetCfg = &BudgetAlertConfig{
				Operator:   op,
				Threshold:  target.Threshold,
				PeriodDays: target.SelectionPeriodDays,
			}
			insightCfg = &InsightMetricAlertConfig{
				Metric:     metric,
				Operator:   op,
				Threshold:  target.Threshold,
				PeriodDays: target.SelectionPeriodDays,
			}
		case *schema.AlertPolicyFieldsAlertTargetsInsightsAssetGroupThresholdTarget:
			metric := target.MetricName
			if friendly, ok := apiToMetricName[metric]; ok {
				metric = friendly
			}
			insightCfg = &InsightMetricAlertConfig{
				AssetGroup: target.AssetGroup,
				Metric:     metric,
				Operator:   strings.ToLower(string(target.Operator)),
				Threshold:  target.Threshold,
				PeriodDays: target.SelectionPeriodDays,
			}
		case *schema.AlertPolicyFieldsAlertTargetsInsightsAssetThresholdTarget:
			metric := target.MetricName
			if friendly, ok := apiToMetricName[metric]; ok {
				metric = friendly
			}
			insightCfg = &InsightMetricAlertConfig{
				AssetKey:   strings.Join(target.AssetKey.Path, "/"),
				Metric:     metric,
				Operator:   strings.ToLower(string(target.Operator)),
				Threshold:  target.Threshold,
				PeriodDays: target.SelectionPeriodDays,
			}
		case *schema.AlertPolicyFieldsAlertTargetsInsightsJobThresholdTarget:
			metric := target.MetricName
			if friendly, ok := apiToMetricName[metric]; ok {
				metric = friendly
			}
			insightCfg = &InsightMetricAlertConfig{
				JobName:    target.JobName,
				Metric:     metric,
				Operator:   strings.ToLower(string(target.Operator)),
				Threshold:  target.Threshold,
				PeriodDays: target.SelectionPeriodDays,
			}
		}
	}

	// Set OnSuccess/OnFailure on run config from event types.
	if runCfg != nil {
		for _, et := range eventTypes {
			switch et {
			case "JOB_SUCCESS":
				runCfg.OnSuccess = true
			case "JOB_FAILURE":
				runCfg.OnFailure = true
			}
		}
	}

	return AlertPolicy{
		ID:                  f.Id,
		Name:                f.Name,
		Description:         f.Description,
		Enabled:             f.Enabled,
		EventTypes:          eventTypes,
		AlertTargetType:     targetType,
		AlertTargetValue:    targetValue,
		NotificationService: ns,
		Run:                 runCfg,
		CodeLocation:        codeCfg,
		Automation:          autoCfg,
		Budget:              budgetCfg,
		InsightMetric:       insightCfg,
	}
}

// CreateOrUpdateAlertPolicy creates or updates an alert policy in a deployment.
func (c *Client) CreateOrUpdateAlertPolicy(ctx context.Context, deployment string, policy AlertPolicy) (*AlertPolicy, error) {
	docBytes, err := json.Marshal(buildPolicyDoc(policy))
	if err != nil {
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: marshalling policy document: %w", err)
	}

	resp, err := schema.CreateOrUpdateAlertPolicy(ctx, c.gqlClient(deployment), json.RawMessage(docBytes))
	if err != nil {
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: %w", err)
	}

	switch r := resp.CreateOrUpdateAlertPolicyFromDocument.(type) {
	case *schema.CreateOrUpdateAlertPolicyCreateOrUpdateAlertPolicyFromDocumentAlertPolicy:
		return c.GetAlertPolicy(ctx, deployment, r.Name)
	case *schema.CreateOrUpdateAlertPolicyCreateOrUpdateAlertPolicyFromDocumentInvalidAlertPolicyError:
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: invalid policy: %s %v", r.Message, r.Errors)
	default:
		return nil, fmt.Errorf("CreateOrUpdateAlertPolicy: unexpected result type %T", resp.CreateOrUpdateAlertPolicyFromDocument)
	}
}

// GetAlertPolicy retrieves an alert policy by name from a deployment.
func (c *Client) GetAlertPolicy(ctx context.Context, deployment, name string) (*AlertPolicy, error) {
	policies, err := c.ListAlertPolicies(ctx, deployment)
	if err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i].Name == name {
			return &policies[i], nil
		}
	}
	return nil, fmt.Errorf("GetAlertPolicy: policy %q not found in deployment %q", name, deployment)
}

// ListAlertPolicies returns all alert policies for a deployment.
func (c *Client) ListAlertPolicies(ctx context.Context, deployment string) ([]AlertPolicy, error) {
	resp, err := schema.ListAlertPolicies(ctx, c.gqlClient(deployment))
	if err != nil {
		return nil, fmt.Errorf("ListAlertPolicies: %w", err)
	}

	policies := make([]AlertPolicy, len(resp.AlertPolicies))
	for i, p := range resp.AlertPolicies {
		policies[i] = alertFieldsToPolicy(p.AlertPolicyFields)
	}
	return policies, nil
}

// DeleteAlertPolicy deletes an alert policy by name from a deployment.
func (c *Client) DeleteAlertPolicy(ctx context.Context, deployment, name string) error {
	resp, err := schema.DeleteAlertPolicy(ctx, c.gqlClient(deployment), name)
	if err != nil {
		return fmt.Errorf("DeleteAlertPolicy: %w", err)
	}

	switch resp.DeleteAlertPolicy.(type) {
	case *schema.DeleteAlertPolicyDeleteAlertPolicyDeleteAlertPolicySuccess:
		return nil
	default:
		return fmt.Errorf("DeleteAlertPolicy: unexpected result type %T", resp.DeleteAlertPolicy)
	}
}

// buildPolicyDoc constructs the GenericScalar document for create/update mutations.
func buildPolicyDoc(policy AlertPolicy) map[string]any {
	doc := map[string]any{
		"name":        policy.Name,
		"description": policy.Description,
		"enabled":     policy.Enabled,
	}

	switch policy.PolicyType {
	case "asset":
		doc["event_types"] = policy.EventTypes
		switch policy.AlertTargetType {
		case "asset_selection":
			doc["alert_targets"] = []any{
				map[string]any{"asset_selection_target": map[string]any{"asset_selection": policy.AlertTargetValue}},
			}
		case "asset_key":
			doc["alert_targets"] = []any{
				map[string]any{"asset_key_target": map[string]any{"asset_key": strings.Split(policy.AlertTargetValue, "/")}},
			}
		}

	case "run":
		var eventTypes []string
		if policy.Run != nil {
			if policy.Run.OnSuccess {
				eventTypes = append(eventTypes, "JOB_SUCCESS")
			}
			if policy.Run.OnFailure {
				eventTypes = append(eventTypes, "JOB_FAILURE")
			}
		}
		doc["event_types"] = eventTypes

		if policy.Run != nil {
			var targets []any
			if !policy.Run.AllRuns {
				runTarget := map[string]any{}
				if len(policy.Run.Filter.JobNames) > 0 {
					jobs := make([]map[string]any, len(policy.Run.Filter.JobNames))
					for i, j := range policy.Run.Filter.JobNames {
						jobs[i] = map[string]any{"job_name": j}
					}
					runTarget["jobs"] = jobs
				}
				if policy.Run.Filter.Tags != "" {
					parts := strings.SplitN(policy.Run.Filter.Tags, "=", 2)
					if len(parts) == 2 {
						runTarget["tags"] = []map[string]any{{"key": parts[0], "value": parts[1]}}
					}
				}
				if len(policy.Run.Filter.LocationNames) > 0 {
					runTarget["location_names"] = policy.Run.Filter.LocationNames
				}
				targets = append(targets, map[string]any{"run_result_target": runTarget})
			}
			if policy.Run.OnTimeoutHours > 0 {
				timeoutTarget := map[string]any{
					"threshold_seconds": policy.Run.OnTimeoutHours * 3600,
				}
				if !policy.Run.AllRuns && len(policy.Run.Filter.JobNames) > 0 {
					jobs := make([]map[string]any, len(policy.Run.Filter.JobNames))
					for i, j := range policy.Run.Filter.JobNames {
						jobs[i] = map[string]any{"job_name": j}
					}
					timeoutTarget["jobs"] = jobs
				}
				targets = append(targets, map[string]any{"long_running_job_threshold_target": timeoutTarget})
			}
			if len(targets) > 0 {
				doc["alert_targets"] = targets
			}
		}

	case "code_location":
		doc["event_types"] = []string{"CODE_LOCATION_ERROR"}
		if policy.CodeLocation != nil && !policy.CodeLocation.AllLocations {
			doc["alert_targets"] = []any{
				map[string]any{"code_location_target": map[string]any{
					"location_names": []string{policy.CodeLocation.LocationName},
				}},
			}
		}

	case "automation":
		doc["event_types"] = []string{"TICK_FAILURE"}
		if policy.Automation != nil {
			target := map[string]any{}
			var types []string
			if policy.Automation.IncludeSchedules {
				types = append(types, "SCHEDULE")
			}
			if policy.Automation.IncludeSensors {
				types = append(types, "SENSOR")
			}
			if len(types) > 0 {
				target["types"] = types
			}
			if len(policy.Automation.LocationNames) > 0 {
				target["location_names"] = policy.Automation.LocationNames
			}
			if len(policy.Automation.SchedulesAndSensors) > 0 {
				names := make([]map[string]any, len(policy.Automation.SchedulesAndSensors))
				for i, n := range policy.Automation.SchedulesAndSensors {
					names[i] = map[string]any{"name": n}
				}
				target["schedules_or_sensors"] = names
			}
			if !policy.Automation.AllSchedulesAndSensors || len(target) > 0 {
				doc["alert_targets"] = []any{
					map[string]any{"schedule_or_sensor_target": target},
				}
			}
			if policy.Automation.MinConsecutiveFailures > 0 {
				doc["policy_options"] = map[string]any{
					"consecutive_failure_threshold": policy.Automation.MinConsecutiveFailures,
				}
			}
		}

	case "budget":
		doc["event_types"] = []string{"INSIGHTS_CONSUMPTION_EXCEEDED"}
		if policy.Budget != nil {
			apiMetric := "__dagster_dagster_credits"
			doc["alert_targets"] = []any{
				map[string]any{"insights_deployment_threshold_target": map[string]any{
					"metric_name":           apiMetric,
					"operator":              strings.ToUpper(policy.Budget.Operator),
					"threshold":             policy.Budget.Threshold,
					"selection_period_days": policy.Budget.PeriodDays,
				}},
			}
		}

	case "insight_metric":
		doc["event_types"] = []string{"INSIGHTS_CONSUMPTION_EXCEEDED"}
		if policy.InsightMetric != nil {
			im := policy.InsightMetric
			apiMetric := im.Metric
			if v, ok := metricNameToAPI[im.Metric]; ok {
				apiMetric = v
			}
			thresholdDoc := map[string]any{
				"metric_name":           apiMetric,
				"operator":              strings.ToUpper(im.Operator),
				"threshold":             im.Threshold,
				"selection_period_days": im.PeriodDays,
			}
			var targetKey string
			switch {
			case im.AssetGroup != "":
				targetKey = "insights_asset_group_threshold_target"
				thresholdDoc["asset_group"] = im.AssetGroup
			case im.AssetKey != "":
				targetKey = "insights_asset_threshold_target"
				thresholdDoc["asset_key"] = strings.Split(im.AssetKey, "/")
			case im.JobName != "":
				targetKey = "insights_job_threshold_target"
				thresholdDoc["job_name"] = im.JobName
			default:
				targetKey = "insights_deployment_threshold_target"
			}
			doc["alert_targets"] = []any{
				map[string]any{targetKey: thresholdDoc},
			}
		}

	default:
		doc["event_types"] = policy.EventTypes
	}

	ns := policy.NotificationService
	switch ns.Type {
	case "email":
		doc["notification_service"] = map[string]any{
			"email": map[string]any{"email_addresses": ns.EmailAddresses},
		}
	case "email_owners":
		inner := map[string]any{}
		if len(ns.DefaultEmailAddresses) > 0 {
			inner["default_email_addresses"] = ns.DefaultEmailAddresses
		}
		doc["notification_service"] = map[string]any{"email_owners": inner}
	case "slack":
		doc["notification_service"] = map[string]any{
			"slack": map[string]any{
				"slack_workspace_name": ns.SlackWorkspaceName,
				"slack_channel_name":   ns.SlackChannelName,
			},
		}
	case "microsoft_teams":
		doc["notification_service"] = map[string]any{
			"microsoft_teams": map[string]any{"webhook_url": ns.WebhookURL},
		}
	case "pagerduty":
		doc["notification_service"] = map[string]any{
			"pagerduty": map[string]any{"integration_key": ns.IntegrationKey},
		}
	case "webhook":
		// The write document expects headers as a dict ({"Key": "Value"}), even
		// though the read path returns them as a KeyValuePair list.
		headers := make(map[string]any, len(ns.WebhookHeaders))
		for k, v := range ns.WebhookHeaders {
			headers[k] = v
		}
		doc["notification_service"] = map[string]any{
			"webhook": map[string]any{
				"webhook_url":   ns.WebhookURL,
				"headers":       headers,
				"body_template": ns.WebhookBodyTemplate,
			},
		}
	}

	return doc
}
