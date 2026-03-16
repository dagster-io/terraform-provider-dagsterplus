data "dagsterplus_code_locations" "prod" {
  deployment = "prod"
}

output "location_names" {
  value = [for l in data.dagsterplus_code_locations.prod.code_locations : l.name]
}
