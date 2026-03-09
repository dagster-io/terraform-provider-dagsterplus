resource "dagsterplus_team" "data_engineering" {
  name = "data-engineering"

  deployment_permission {
    deployment_name = "prod"
    role            = "EDITOR"
  }

  deployment_permission {
    deployment_name = "staging"
    role            = "ADMIN"
  }
}

resource "dagsterplus_team" "analysts" {
  name = "analysts"

  deployment_permission {
    deployment_name = "prod"
    role            = "VIEWER"
  }
}
