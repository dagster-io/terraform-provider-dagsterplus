# Singleton resource — only one per organization
resource "dagsterplus_organization_settings" "this" {
  settings_json = jsonencode({
    sso_default_role = "VIEWER"
  })
}
