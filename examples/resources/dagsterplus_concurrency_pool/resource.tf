# Limit the "database" pool to a single in-progress execution at a time.
# Assets/ops assigned to this pool in your Dagster code (e.g. @dg.asset(pool="database"))
# will share this limit.
resource "dagsterplus_concurrency_pool" "database" {
  deployment = "prod"
  name       = "database"
  limit      = 1
}

# A pool allowing more concurrency for a less contended resource.
resource "dagsterplus_concurrency_pool" "external_api" {
  deployment = "prod"
  name       = "external_api"
  limit      = 25
}
