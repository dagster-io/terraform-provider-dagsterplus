# Import a custom role by its ID (visible in the Dagster+ UI URL or API response).
terraform import dagsterplus_role.data_engineer <role-id>

# Note: role_type is returned by the API and round-trips correctly through import.
