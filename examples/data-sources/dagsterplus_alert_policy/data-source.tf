# Look up an existing alert policy created outside Terraform
data "dagsterplus_alert_policy" "existing" {
  deployment  = "prod"
  name        = "asset-health-degraded"
  policy_type = "asset"
}

output "notification_type" {
  value = data.dagsterplus_alert_policy.existing.notification_service.type
}

output "event_types" {
  value = data.dagsterplus_alert_policy.existing.event_types
}
