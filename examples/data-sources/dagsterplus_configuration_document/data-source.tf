data "dagsterplus_configuration_document" "example" {
  yaml_body = <<-YAML
    run_queue:
      max_concurrent_runs: 10
    run_monitoring:
      enabled: true
  YAML
}

output "settings_json" {
  value = data.dagsterplus_configuration_document.example.json
}
