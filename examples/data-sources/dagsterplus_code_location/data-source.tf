data "dagsterplus_code_location" "pipeline" {
  deployment = "prod"
  name            = "my-pipeline"
}

output "pipeline_image" {
  value = data.dagsterplus_code_location.pipeline.image
}
