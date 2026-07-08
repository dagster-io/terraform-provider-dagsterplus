# Changelog

All notable changes to this provider will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/). This project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- `dagsterplus_concurrency_pool` resource for managing per-pool concurrency limits (0-1000) within a deployment. A pool is a named concurrency key shared by the assets/ops assigned to it via `pool=` in Dagster code. Supports import via `{deployment}/{name}`; on destroy the explicit limit is removed and the pool reverts to the deployment's default pool limit.
- `dagsterplus_alert_policy`: the `notification_service` block now supports a generic `webhook` notification type, mirroring the webhook channel available in the Dagster+ UI (e.g. for IncidentIO). Set `type = "webhook"` and configure `webhook_url` (the endpoint), `headers` (a `map(string)` of arbitrary request headers such as authentication tokens — marked sensitive), and `body_template` (a templated request body). This is distinct from the Teams-specific `microsoft_teams` type. Note that the webhook URL's domain must be allowlisted for your organization (a support-gated setting), and `body_template` tokens must be lowercase identifiers (e.g. `{{alert_summary}}`) or environment variables (e.g. `{{env.MY_SECRET}}`). Existing configs are unaffected; `webhook` is a new value added to the notification `type` enum. ([#40](https://github.com/dagster-io/terraform-provider-dagsterplus/issues/40))

## [0.1.7] - 2026-06-25

### Fixed
- All `team`/`user`/`service_user` grant resources and inline grant blocks (across the `organization`, `deployment`, `all_branch_deployments`, and `branch_deployments` scopes, plus per-location grants) now require exactly one of `grant` or `custom_role_id` to be set. Previously both were marked `Optional`, so a config that set neither (e.g. only `location_grants`) passed validation but sent an empty grant to the API, failing apply with an opaque `Value '' does not exist in 'PermissionGrant' enum` error. The requirement is now enforced at plan time with a clear validation message. Configs that already set one of the two attributes are unaffected. ([#37](https://github.com/dagster-io/terraform-provider-dagsterplus/issues/37))

## [0.1.6] - 2026-06-09

### Notes
- Migration note for `dagsterplus_agent_token`: the new `organization` attribute defaults to `true` and is now actively reconciled. If you previously removed a token's organization-scoped grant **outside Terraform** (e.g. via a `graphql_mutation` resource), Terraform will re-add it on the next apply — set `organization = false` in your config to keep the grant removed. Tokens that still hold their default organization grant are unaffected.

### Added
- Permission grants for `dagsterplus_agent_token`. Agent tokens only ever carry the `AGENT` permission, so grants are expressed as inline attributes on the token resource rather than nested blocks: `organization` (bool, default `true`), `all_branch_deployments` (bool, default `false`), `deployment_grants` (set of deployment names), and `branch_deployments_grants` (set of parent deployment names). Setting `organization = false` removes the organization-scoped grant that Dagster+ creates automatically with every new token.
- `dagsterplus_agent_token_deployment_grant` — standalone resource granting an agent token the `AGENT` permission on a single deployment, for composing grants in a separate module/lifecycle from the token (e.g. `for_each` over deployments). Imported via `{agent_token_id}/{deployment}`.
- Full matrix of permission-grant resources for `team`, `user`, and `service_user` principals across four scopes (`organization`, `deployment`, `all_branch_deployments`, `branch_deployments`). Each scope is available both as an inline block on the parent resource and as a standalone `{principal}_{scope}_grant` resource for out-of-band lifecycle management.
- `branch_deployments_grant { parent_deployment = … }` — new scope expressing grants that apply to all branch deployments of a specific parent (full) deployment. Symmetric with `deployment_grant { deployment = … }`.

### Changed
- Attribute documentation now lists the full set of valid values for every enum attribute, so users can discover them from the registry docs without trial-and-error. Affected attributes: `dagsterplus_role` `permissions` (the complete 30-value permission list) and `role_type`; the `grant` permission level (`VIEWER`/`LAUNCHER`/`EDITOR`/`ADMIN`) and organization `grant` (`ADMIN`) on all `team`/`user`/`service_user` grant resources and inline blocks; `dagsterplus_external_asset_connection` `connection_type`; and `dagsterplus_alert_policy` `policy_type`, budget/insight-metric `operator`, and notification `type`. Validator behavior is unchanged — the accepted values are identical to previous releases.

## [0.1.5] - 2026-05-01

### Added
- `pkg/provider` package exporting a top-level `New` constructor to support the Pulumi Terraform bridge.

## [0.1.4] - 2026-04-28

### Breaking Changes
- `dagsterplus_code_location`: `working_directory`, `executable_path`, `attribute`, and `agent_queue` are no longer `Computed`. These fields were previously `Optional+Computed`, meaning Terraform would persist whatever the API returned even if the field was omitted from config. They are now `Optional`-only and will be `null` in state when not set. Users with existing state that has these fields populated from a prior computed read should set the values explicitly in their config (or run `terraform state rm` and re-import) to avoid a persistent plan diff.

### Fixed
- Code location fields no longer incorrectly show as unknown after creation; optional fields now preserve state correctly.
- Provider now strips trailing slashes from `base_url` to prevent request failures when the URL is supplied with a trailing `/`.
- HTTP 500 responses containing a `PythonError` body are now handled gracefully instead of causing an opaque client error.
- `dagsterplus_user` resource no longer removes itself from state when the invited user is in a pending-invite state; state is preserved and refreshed once the invite is accepted.

### Changed
- Acceptance tests now run in CI against the live API on every pull request.

## [0.1.3] - 2026-04-14

### Added
- `container_context` support on `dagsterplus_code_location_from_document`, allowing container image and environment configuration.

### Fixed
- `dagsterplus_code_location_from_document` now accepts camelCase keys in the document (snake_case keys were removed as a simplification).
- Improved error handling: `InvalidLocationError` responses now surface the full message and error list.

## [0.1.2] - 2026-04-02

### Changed
- Provider documentation now includes a complete index template and improved formatting on the Terraform Registry.

## [0.1.1] - 2026-04-02

### Added
- Automated documentation generation via `tfplugindocs` in CI (docs are now auto-committed on changes to provider code or examples).

## [0.1.0] - 2026-04-02

### Added
- Initial public release.
- **21 resources:** `dagsterplus_deployment`, `dagsterplus_code_location`, `dagsterplus_team`, `dagsterplus_user`, `dagsterplus_alert_policy`, `dagsterplus_agent_token`, `dagsterplus_user_token`, `dagsterplus_role`, `dagsterplus_scim_settings`, `dagsterplus_atlan_integration`, `dagsterplus_github_integration`, `dagsterplus_secret`, `dagsterplus_deployment_settings`, `dagsterplus_service_user`, `dagsterplus_service_token`, `dagsterplus_organization_settings`, `dagsterplus_custom_metric`, `dagsterplus_external_asset_connection`, `dagsterplus_team_membership`, `dagsterplus_team_deployment_grant`, `dagsterplus_code_location_from_document`.
- **20 data sources** for all major resource types.
- Import support for all resources.
- Migration guide from `datarootsio/dagster` provider.
- Full CI pipeline (build, vet, format, unit tests, module tidy, generated code check).
- GoReleaser-based release pipeline with GPG-signed checksums.

[Unreleased]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.7...HEAD
[0.1.7]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/dagster-io/terraform-provider-dagsterplus/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/dagster-io/terraform-provider-dagsterplus/releases/tag/v0.1.0
