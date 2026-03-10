resource "dagsterplus_team" "data_engineering" {
  name = "data-engineering"

  deployment_grant {
    deployment = "prod"
    grant      = "EDITOR"
  }

  deployment_grant {
    deployment = "staging"
    grant      = "ADMIN"
  }
}

resource "dagsterplus_team" "analysts" {
  name = "analysts"

  deployment_grant {
    deployment = "prod"
    grant      = "VIEWER"
  }
}
