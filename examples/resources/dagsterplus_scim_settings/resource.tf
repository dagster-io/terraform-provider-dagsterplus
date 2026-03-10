# Enable SCIM sync for the organization.
resource "dagsterplus_scim_settings" "example" {
  enabled = true
}

# Disable SCIM sync.
resource "dagsterplus_scim_settings" "disabled" {
  enabled = false
}
