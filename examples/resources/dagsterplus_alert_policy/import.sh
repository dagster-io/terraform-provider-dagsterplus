# Alert policies are imported using the format: <deployment>/<policy_name>
# Note: the `policy_type` attribute is not returned by the API and must be set
# manually after import. Add it to ignore_changes or set it explicitly.
terraform import dagsterplus_alert_policy.example prod/asset-health-degraded
