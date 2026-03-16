data "dagsterplus_alert_policies" "prod" {
  deployment = "prod"
}

output "policy_names" {
  value = [for p in data.dagsterplus_alert_policies.prod.alert_policies : p.name]
}
