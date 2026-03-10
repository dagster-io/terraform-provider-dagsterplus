# User tokens are imported using the token ID assigned by Dagster+.
# The token ID can be found via the Dagster+ API (ListUserTokens query).
# IMPORTANT: The token value (secret) is only available at creation time and
# cannot be recovered after import. The `token` attribute will be empty after
# import — plan accordingly (e.g. use ignore_changes or recreate the token).
terraform import dagsterplus_user_token.example <token_id>
