# Import a service token by its server-assigned ID
# Note: the token value cannot be recovered after import — add it to ImportStateVerifyIgnore
# Note: service_user_id must be set manually after import as it is not returned by the API
terraform import dagsterplus_service_token.ci_token <token-id>
