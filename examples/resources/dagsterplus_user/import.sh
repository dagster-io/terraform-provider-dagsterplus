# Users are imported using the user ID assigned by Dagster+.
# The user ID can be found via the Dagster+ API (ListUsers query).
# Note: the `name` attribute is set by Dagster+ after the user accepts the invite
# and is not returned on import — add it to ignore_changes if needed.
terraform import dagsterplus_user.example <user_id>
